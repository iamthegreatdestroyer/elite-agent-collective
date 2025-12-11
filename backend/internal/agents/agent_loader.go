// Package agents provides the agent registry and HTTP handlers.
package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
	"gopkg.in/yaml.v3"
)

// AgentFileMetadata represents the YAML frontmatter of an agent file.
type AgentFileMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Codename    string `yaml:"codename"`
	Tier        int    `yaml:"tier"`
	ID          string `yaml:"id"`
	Category    string `yaml:"category"`
}

// LoadAgentFromFile loads an agent definition from a .agent.md file.
// The file should have YAML frontmatter followed by Markdown content.
func LoadAgentFromFile(filePath string) (*models.Agent, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent file: %w", err)
	}

	// Parse YAML frontmatter and content
	metadata, markdownContent, err := parseFrontmatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter from %s: %w", filePath, err)
	}

	// Extract fields from markdown
	philosophy := extractPhilosophy(markdownContent)
	examples := extractExamples(markdownContent)
	collaborators := extractCollaborators(markdownContent)
	directives := extractDirectives(markdownContent)

	// Create Agent model
	agent := &models.Agent{
		ID:            metadata.ID,
		Codename:      metadata.Codename,
		Tier:          metadata.Tier,
		Name:          metadata.Name,
		Specialty:     metadata.Description,
		Philosophy:    philosophy,
		Keywords:      []string{}, // Can be extracted from markdown if needed
		Directives:    directives,
		Examples:      examples,
		Collaborators: collaborators,
		Category:      metadata.Category,
		MarkdownPath:  filePath,
	}

	return agent, nil
}

// LoadAllAgentsFromDirectory loads all agent definitions from .github/agents/ directory.
func LoadAllAgentsFromDirectory(agentsDir string) ([]models.Agent, error) {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents directory: %w", err)
	}

	agents := make([]models.Agent, 0)
	var loadErrors []error

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .agent.md files
		if !strings.HasSuffix(entry.Name(), ".agent.md") {
			continue
		}

		filePath := filepath.Join(agentsDir, entry.Name())
		agent, err := LoadAgentFromFile(filePath)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load %s: %w", entry.Name(), err))
			continue
		}

		agents = append(agents, *agent)
	}

	if len(loadErrors) > 0 {
		// Log errors but don't fail completely if some agents fail to load
		fmt.Fprintf(os.Stderr, "Warnings loading agents:\n")
		for _, err := range loadErrors {
			fmt.Fprintf(os.Stderr, "  - %v\n", err)
		}
	}

	return agents, nil
}

// parseFrontmatter extracts YAML frontmatter from a file.
// Format: --- YAML --- Content
func parseFrontmatter(content string) (*AgentFileMetadata, string, error) {
	// Find frontmatter boundaries
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return nil, "", fmt.Errorf("missing frontmatter delimiter at start")
	}

	// Find closing delimiter
	endDelimiterIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endDelimiterIdx = i
			break
		}
	}

	if endDelimiterIdx == -1 {
		return nil, "", fmt.Errorf("missing closing frontmatter delimiter")
	}

	// Parse YAML metadata
	yamlContent := strings.Join(lines[1:endDelimiterIdx], "\n")
	var metadata AgentFileMetadata
	if err := yaml.Unmarshal([]byte(yamlContent), &metadata); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Get remaining content
	markdownContent := strings.Join(lines[endDelimiterIdx+1:], "\n")

	return &metadata, markdownContent, nil
}

// extractPhilosophy extracts the philosophy statement from markdown.
// Format: **Philosophy:** _"[Statement]"_
func extractPhilosophy(content string) string {
	re := regexp.MustCompile(`\*\*Philosophy:\*\*\s+_"([^"]+)"_`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractDirectives extracts directives from the Core Capabilities or similar section.
// Returns list of capability descriptions as directives.
func extractDirectives(content string) []string {
	var directives []string

	// Find "Core Capabilities" section
	lines := strings.Split(content, "\n")
	inCapabilities := false

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Check for Core Capabilities section header
		if strings.Contains(line, "Core Capabilities") {
			inCapabilities = true
			continue
		}

		// If we hit another section header, stop
		if inCapabilities && strings.HasPrefix(line, "##") && !strings.Contains(line, "Core Capabilities") {
			break
		}

		// Extract bullet points as directives
		if inCapabilities && strings.HasPrefix(line, "-") {
			directive := strings.TrimPrefix(strings.TrimSpace(line), "- ")
			if directive != "" && !strings.Contains(directive, "---") {
				directives = append(directives, directive)
			}
		}
	}

	return directives
}

// extractCapabilities extracts capability descriptions.
// Returns first 5 capabilities from Core Capabilities section.
func extractCapabilities(content string) []string {
	return extractDirectives(content) // Reuse directive extraction for capabilities
}

// extractExamples extracts invocation examples from the Invocation Examples section.
func extractExamples(content string) []string {
	var examples []string

	lines := strings.Split(content, "\n")
	inExamples := false

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Check for Invocation Examples section
		if strings.Contains(line, "Invocation Examples") {
			inExamples = true
			continue
		}

		// If we hit another section header, stop
		if inExamples && strings.HasPrefix(line, "##") && !strings.Contains(line, "Invocation") {
			break
		}

		// Extract backtick-quoted examples
		if inExamples && strings.HasPrefix(line, "@") {
			examples = append(examples, line)
		}
	}

	return examples
}

// extractCollaborators extracts collaborating agent names.
// Format: "Consults with @AGENT1, @AGENT2"
func extractCollaborators(content string) []string {
	var collaborators []string

	// Look for Multi-Agent Collaboration or similar sections
	re := regexp.MustCompile(`@([A-Z]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	seen := make(map[string]bool)
	for _, match := range matches {
		agentName := match[1]
		// Skip if it's the agent itself (would be in title/philosophy context)
		if agentName != "" && !seen[agentName] {
			seen[agentName] = true
			collaborators = append(collaborators, agentName)
		}
	}

	return collaborators
}

// ValidateAgentID validates and extracts numeric ID from agent metadata.
func ValidateAgentID(idStr string) (int, error) {
	idNum, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		return 0, fmt.Errorf("invalid agent ID: %s", idStr)
	}
	if idNum < 1 || idNum > 40 {
		return 0, fmt.Errorf("agent ID out of range (1-40): %d", idNum)
	}
	return idNum, nil
}
