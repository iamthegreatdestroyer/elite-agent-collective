package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// AgentPipeline executes agents sequentially, feeding each agent's output as
// context into the next agent's input (LangGraph-style sequential graph).
type AgentPipeline struct {
	registry *Registry
}

// NewAgentPipeline creates a new pipeline executor backed by registry.
func NewAgentPipeline(registry *Registry) *AgentPipeline {
	return &AgentPipeline{registry: registry}
}

// pipelineStep holds the result of a single agent execution in a pipeline.
type pipelineStep struct {
	agent   string
	output  string
	skipped bool
}

// Execute runs codenames in order. Each agent's response is appended as an
// assistant turn before invoking the next agent, so every agent in the
// pipeline sees all prior reasoning. Returns a combined response.
func (p *AgentPipeline) Execute(ctx context.Context, req *models.CopilotRequest, codenames []string) (*models.CopilotResponse, error) {
	var steps []pipelineStep
	cur := cloneRequest(req)

	for _, codename := range codenames {
		agent, err := p.registry.Get(codename)
		if err != nil {
			steps = append(steps, pipelineStep{agent: codename, skipped: true})
			continue
		}

		resp, err := agent.Handle(ctx, cur)
		if err != nil || len(resp.Choices) == 0 {
			steps = append(steps, pipelineStep{agent: codename, skipped: true})
			continue
		}

		output := resp.Choices[0].Message.Content
		steps = append(steps, pipelineStep{agent: codename, output: output})

		// Inject this agent's output as context for the next agent.
		cur = appendAssistant(cur, fmt.Sprintf("[%s]: %s", codename, output))
	}

	return copilot.NewResponse(formatPipeline(steps)), nil
}

func cloneRequest(req *models.CopilotRequest) *models.CopilotRequest {
	msgs := make([]models.Message, len(req.Messages))
	copy(msgs, req.Messages)
	return &models.CopilotRequest{Messages: msgs, Model: req.Model, Stream: req.Stream}
}

func appendAssistant(req *models.CopilotRequest, content string) *models.CopilotRequest {
	msgs := make([]models.Message, len(req.Messages)+1)
	copy(msgs, req.Messages)
	msgs[len(req.Messages)] = models.Message{Role: "assistant", Content: content}
	return &models.CopilotRequest{Messages: msgs, Model: req.Model, Stream: req.Stream}
}

func formatPipeline(steps []pipelineStep) string {
	var active []string
	for _, s := range steps {
		if !s.skipped {
			active = append(active, s.agent)
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Multi-Agent Pipeline: %s\n\n", strings.Join(active, " → ")))
	for _, s := range steps {
		if s.skipped {
			fmt.Fprintf(&sb, "### @%s *(unavailable — skipped)*\n\n", s.agent)
		} else {
			fmt.Fprintf(&sb, "### @%s\n\n%s\n\n---\n\n", s.agent, s.output)
		}
	}
	return strings.TrimSuffix(sb.String(), "---\n\n")
}
