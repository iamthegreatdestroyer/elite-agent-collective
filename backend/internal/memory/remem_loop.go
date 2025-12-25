// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the ReMem-Elite control loop: RETRIEVE → THINK → ACT → REFLECT → EVOLVE

package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// ============================================================================
// ReMem Control Loop Phases
// ============================================================================

// Phase represents a phase in the ReMem control loop.
type Phase string

const (
	PhaseRetrieve Phase = "RETRIEVE" // Sub-linear experience retrieval
	PhaseThink    Phase = "THINK"    // Reason with augmented context
	PhaseAct      Phase = "ACT"      // Execute agent with memory
	PhaseReflect  Phase = "REFLECT"  // Evaluate outcome
	PhaseEvolve   Phase = "EVOLVE"   // Update memory
)

// ============================================================================
// Context Constructor - Builds memory-augmented context
// ============================================================================

// ContextConstructor builds augmented context from retrieved experiences.
type ContextConstructor struct {
	maxExperiencesInContext int
	maxContextTokens        int
}

// NewContextConstructor creates a new context constructor with defaults.
func NewContextConstructor() *ContextConstructor {
	return &ContextConstructor{
		maxExperiencesInContext: 5,
		maxContextTokens:        4000, // Leave room for response
	}
}

// AugmentedContext contains the original request plus memory-enhanced context.
type AugmentedContext struct {
	// OriginalRequest is the user's request
	OriginalRequest *models.CopilotRequest

	// AgentExperiences are relevant experiences from the same agent
	AgentExperiences []*ExperienceTuple

	// TierExperiences are relevant experiences from same-tier agents
	TierExperiences []*ExperienceTuple

	// CollectiveBreakthroughs are cross-tier breakthrough discoveries
	CollectiveBreakthroughs []*Breakthrough

	// MemoryPrompt is the formatted memory injection for the agent
	MemoryPrompt string

	// RetrievalResult contains metadata about the retrieval
	RetrievalResult *RetrievalResult
}

// Build constructs an augmented context from the request and retrieved experiences.
func (c *ContextConstructor) Build(
	request *models.CopilotRequest,
	agentExperiences []*ExperienceTuple,
	tierExperiences []*ExperienceTuple,
	breakthroughs []*Breakthrough,
) *AugmentedContext {
	ctx := &AugmentedContext{
		OriginalRequest:         request,
		AgentExperiences:        agentExperiences,
		TierExperiences:         tierExperiences,
		CollectiveBreakthroughs: breakthroughs,
	}

	// Build memory prompt
	ctx.MemoryPrompt = c.buildMemoryPrompt(agentExperiences, tierExperiences, breakthroughs)

	return ctx
}

// buildMemoryPrompt formats experiences into a prompt injection.
func (c *ContextConstructor) buildMemoryPrompt(
	agentExps []*ExperienceTuple,
	tierExps []*ExperienceTuple,
	breakthroughs []*Breakthrough,
) string {
	var sb strings.Builder

	// Add agent-specific experiences
	if len(agentExps) > 0 {
		sb.WriteString("\n<RELEVANT_EXPERIENCES>\n")
		sb.WriteString("Based on similar tasks you've handled successfully:\n\n")

		for i, exp := range agentExps {
			if i >= c.maxExperiencesInContext {
				break
			}
			sb.WriteString(fmt.Sprintf("Experience %d (Success: %v, Fitness: %.2f):\n", i+1, exp.Success, exp.FitnessScore))
			sb.WriteString(fmt.Sprintf("  Strategy: %s\n", truncateString(exp.Strategy, 200)))
			if exp.Success {
				sb.WriteString(fmt.Sprintf("  Key insight: %s\n", extractKeyInsight(exp)))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("</RELEVANT_EXPERIENCES>\n")
	}

	// Add tier-shared knowledge
	if len(tierExps) > 0 {
		sb.WriteString("\n<TIER_KNOWLEDGE>\n")
		sb.WriteString("Insights from peer agents in your tier:\n\n")

		for i, exp := range tierExps {
			if i >= 3 { // Limit tier experiences
				break
			}
			sb.WriteString(fmt.Sprintf("From @%s: %s\n", exp.AgentID, truncateString(exp.Strategy, 150)))
		}
		sb.WriteString("</TIER_KNOWLEDGE>\n")
	}

	// Add breakthrough discoveries
	if len(breakthroughs) > 0 {
		sb.WriteString("\n<BREAKTHROUGHS>\n")
		sb.WriteString("Notable discoveries from the collective:\n\n")

		for i, bt := range breakthroughs {
			if i >= 2 { // Limit breakthroughs
				break
			}
			sb.WriteString(fmt.Sprintf("From @%s (Tier %d): %s\n",
				bt.OriginAgent, bt.OriginTier, truncateString(bt.Strategy, 150)))
		}
		sb.WriteString("</BREAKTHROUGHS>\n")
	}

	return sb.String()
}

// truncateString truncates a string to maxLen characters.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// extractKeyInsight extracts the key insight from an experience.
func extractKeyInsight(exp *ExperienceTuple) string {
	if insight, ok := exp.Metadata["key_insight"].(string); ok {
		return insight
	}
	// Default to first 100 chars of strategy
	return truncateString(exp.Strategy, 100)
}

// ============================================================================
// Memory Updater - Updates memory based on execution outcomes
// ============================================================================

// MemoryUpdater handles updating memory after execution.
type MemoryUpdater struct {
	retriever *SubLinearRetriever
	evaluator *OutcomeEvaluator
	mu        sync.Mutex
}

// NewMemoryUpdater creates a new memory updater.
func NewMemoryUpdater(retriever *SubLinearRetriever) *MemoryUpdater {
	return &MemoryUpdater{
		retriever: retriever,
		evaluator: NewOutcomeEvaluator(),
	}
}

// AddAndEvolve adds a new experience and evolves related experiences.
func (u *MemoryUpdater) AddAndEvolve(exp *ExperienceTuple, eval *Evaluation) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Update fitness based on evaluation
	exp.FitnessScore = u.computeFitness(exp, eval)
	exp.Success = eval.Success

	// Store evaluation in metadata
	exp.Metadata["evaluation_score"] = eval.Score
	exp.Metadata["evaluation_feedback"] = eval.Feedback

	// Add to retriever
	if err := u.retriever.Add(exp); err != nil {
		return fmt.Errorf("failed to add experience: %w", err)
	}

	// Update fitness of retrieved experiences that were used
	u.updateRetrievedFitness(exp, eval)

	return nil
}

// computeFitness calculates the fitness score for an experience.
func (u *MemoryUpdater) computeFitness(exp *ExperienceTuple, eval *Evaluation) float64 {
	baseFitness := eval.Score

	// Boost fitness for successful experiences
	if eval.Success {
		baseFitness = baseFitness*0.8 + 0.2
	}

	// Apply time decay (newer experiences slightly preferred)
	ageHours := float64(time.Now().UnixNano()-exp.Timestamp) / float64(time.Hour)
	decayFactor := 1.0 / (1.0 + ageHours/1000.0) // Slow decay over ~1000 hours

	return baseFitness * decayFactor
}

// updateRetrievedFitness updates fitness of experiences that were retrieved and used.
func (u *MemoryUpdater) updateRetrievedFitness(newExp *ExperienceTuple, eval *Evaluation) {
	// Get IDs of experiences that were used
	usedIDs, ok := newExp.Metadata["retrieved_experiences"].([]string)
	if !ok || len(usedIDs) == 0 {
		return
	}

	// Update fitness based on whether they contributed to success
	fitnessAdjustment := 0.05
	if !eval.Success {
		fitnessAdjustment = -0.02 // Smaller penalty for failure
	}

	for _, id := range usedIDs {
		// Get and update the experience
		exps := u.retriever.GetByAgent(newExp.AgentID)
		for _, exp := range exps {
			if exp.ID == id {
				exp.FitnessScore = clamp(exp.FitnessScore+fitnessAdjustment, 0.0, 1.0)
				break
			}
		}
	}
}

// clamp restricts a value to a range.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ============================================================================
// Outcome Evaluator - Evaluates execution outcomes
// ============================================================================

// OutcomeEvaluator evaluates the outcome of an agent execution.
type OutcomeEvaluator struct {
	// Configurable evaluation criteria
	successIndicators []string
	failureIndicators []string
}

// NewOutcomeEvaluator creates a new outcome evaluator with default indicators.
func NewOutcomeEvaluator() *OutcomeEvaluator {
	return &OutcomeEvaluator{
		successIndicators: []string{
			"successfully", "completed", "solved", "implemented",
			"fixed", "resolved", "created", "optimized",
		},
		failureIndicators: []string{
			"error", "failed", "couldn't", "unable", "sorry",
			"cannot", "impossible", "invalid",
		},
	}
}

// Evaluate assesses the outcome of an execution.
func (e *OutcomeEvaluator) Evaluate(
	request *models.CopilotRequest,
	response *models.CopilotResponse,
	trace *ExecutionTrace,
) *Evaluation {
	eval := &Evaluation{
		EvaluatedAt: time.Now().UnixNano(),
		Metrics:     make(map[string]float64),
	}

	if response == nil || len(response.Choices) == 0 {
		eval.Success = false
		eval.Score = 0.0
		eval.Feedback = "No response generated"
		return eval
	}

	content := response.Choices[0].Message.Content

	// Heuristic-based evaluation
	successScore := e.computeSuccessScore(content)
	completenessScore := e.computeCompletenessScore(request, content)
	relevanceScore := e.computeRelevanceScore(request, content)

	eval.Score = (successScore*0.4 + completenessScore*0.3 + relevanceScore*0.3)
	eval.Success = eval.Score >= 0.6

	eval.Metrics["success_indicators"] = successScore
	eval.Metrics["completeness"] = completenessScore
	eval.Metrics["relevance"] = relevanceScore
	eval.Metrics["response_length"] = float64(len(content))
	eval.Metrics["latency_ms"] = float64(trace.EndTime-trace.StartTime) / float64(time.Millisecond)

	if eval.Success {
		eval.Feedback = "Task completed successfully"
	} else {
		eval.Feedback = "Task may not have been fully completed"
	}

	return eval
}

// computeSuccessScore calculates a score based on success/failure indicators.
func (e *OutcomeEvaluator) computeSuccessScore(content string) float64 {
	lowerContent := strings.ToLower(content)

	successCount := 0
	for _, indicator := range e.successIndicators {
		if strings.Contains(lowerContent, indicator) {
			successCount++
		}
	}

	failureCount := 0
	for _, indicator := range e.failureIndicators {
		if strings.Contains(lowerContent, indicator) {
			failureCount++
		}
	}

	if successCount+failureCount == 0 {
		return 0.5 // Neutral
	}

	return float64(successCount) / float64(successCount+failureCount)
}

// computeCompletenessScore estimates how complete the response is.
func (e *OutcomeEvaluator) computeCompletenessScore(request *models.CopilotRequest, content string) float64 {
	// Simple heuristic: longer responses tend to be more complete
	// Optimal length varies, but we use a sigmoid around 500 chars
	length := float64(len(content))
	return 1.0 / (1.0 + exp(-0.01*(length-300)))
}

// computeRelevanceScore estimates how relevant the response is to the request.
func (e *OutcomeEvaluator) computeRelevanceScore(request *models.CopilotRequest, content string) float64 {
	if len(request.Messages) == 0 {
		return 0.5
	}

	// Extract keywords from request
	lastMessage := request.Messages[len(request.Messages)-1].Content
	requestWords := strings.Fields(strings.ToLower(lastMessage))

	// Count keyword matches in response
	lowerContent := strings.ToLower(content)
	matches := 0
	for _, word := range requestWords {
		if len(word) > 3 && strings.Contains(lowerContent, word) {
			matches++
		}
	}

	if len(requestWords) == 0 {
		return 0.5
	}

	return float64(matches) / float64(len(requestWords))
}

// exp is a helper for exponential function.
func exp(x float64) float64 {
	if x > 700 {
		return 1e308
	}
	if x < -700 {
		return 0
	}
	// Simple Taylor approximation for small values
	result := 1.0
	term := 1.0
	for i := 1; i <= 20; i++ {
		term *= x / float64(i)
		result += term
	}
	return result
}

// ============================================================================
// ReMem Controller - Main control loop orchestrator
// ============================================================================

// AgentExecutor defines the interface for executing agents.
type AgentExecutor interface {
	// Execute runs the agent and returns response with execution trace.
	Execute(ctx context.Context, augmentedCtx *AugmentedContext) (*models.CopilotResponse, *ExecutionTrace, error)
}

// EmbeddingService defines the interface for computing embeddings.
type EmbeddingService interface {
	// Embed computes an embedding vector for the given text.
	Embed(text string) ([]float32, error)
}

// ReMemController implements the Think-Act-Refine loop for Elite Agents.
type ReMemController struct {
	retriever        *SubLinearRetriever
	contextBuilder   *ContextConstructor
	updater          *MemoryUpdater
	evaluator        *OutcomeEvaluator
	impasseDetector  *ImpasseDetector
	consolidator     *MemoryConsolidator
	embeddingService EmbeddingService

	// Agent registry for tier lookups
	agentTiers map[string]int // agent_id -> tier_id

	// Breakthrough tracking
	breakthroughs  []*Breakthrough
	breakthroughMu sync.RWMutex

	// Evolution tracking
	generationCounter map[string]int // agent_id -> current generation
	genMu             sync.RWMutex

	// Configuration
	config *ReMemConfig
}

// ReMemConfig contains configuration for the ReMem controller.
type ReMemConfig struct {
	// MaxExperiencesPerAgent limits experiences stored per agent
	MaxExperiencesPerAgent int

	// MinFitnessThreshold for retrieval
	MinFitnessThreshold float64

	// BreakthroughThreshold for promoting experiences to collective
	BreakthroughThreshold float64

	// EmbeddingDimension is the size of embedding vectors
	EmbeddingDimension int
}

// DefaultReMemConfig returns the default configuration.
func DefaultReMemConfig() *ReMemConfig {
	return &ReMemConfig{
		MaxExperiencesPerAgent: 1000,
		MinFitnessThreshold:    0.3,
		BreakthroughThreshold:  0.9,
		EmbeddingDimension:     384, // Common small embedding dimension
	}
}

// NewReMemController creates a new ReMem controller.
func NewReMemController(config *ReMemConfig, embeddingService EmbeddingService) *ReMemController {
	if config == nil {
		config = DefaultReMemConfig()
	}

	retriever := NewSubLinearRetriever(config.EmbeddingDimension)
	goalStack := NewGoalStack(DefaultGoalStackConfig()) // Initialize goal stack for impasse resolution
	impasseDetector := NewImpasseDetector(DefaultImpasseDetectorConfig(), goalStack)
	consolidator := NewMemoryConsolidator(DefaultConsolidatorConfig())

	return &ReMemController{
		retriever:         retriever,
		contextBuilder:    NewContextConstructor(),
		updater:           NewMemoryUpdater(retriever),
		evaluator:         NewOutcomeEvaluator(),
		impasseDetector:   impasseDetector,
		consolidator:      consolidator,
		embeddingService:  embeddingService,
		agentTiers:        initializeAgentTiers(),
		breakthroughs:     make([]*Breakthrough, 0),
		generationCounter: make(map[string]int),
		config:            config,
	}
}

// initializeAgentTiers creates the agent-to-tier mapping for all 40 agents.
func initializeAgentTiers() map[string]int {
	return map[string]int{
		// Tier 1: Foundational
		"APEX": 1, "CIPHER": 1, "ARCHITECT": 1, "AXIOM": 1, "VELOCITY": 1,
		// Tier 2: Specialists
		"QUANTUM": 2, "TENSOR": 2, "FORTRESS": 2, "NEURAL": 2, "CRYPTO": 2,
		"FLUX": 2, "PRISM": 2, "SYNAPSE": 2, "CORE": 2, "HELIX": 2,
		"VANGUARD": 2, "ECLIPSE": 2,
		// Tier 3: Innovators
		"NEXUS": 3, "GENESIS": 3,
		// Tier 4: Meta
		"OMNISCIENT": 4,
		// Tier 5: Domain Specialists
		"ATLAS": 5, "FORGE": 5, "SENTRY": 5, "VERTEX": 5, "STREAM": 5,
		// Tier 6: Emerging Tech
		"PHOTON": 6, "LATTICE": 6, "MORPH": 6, "PHANTOM": 6, "ORBIT": 6,
		// Tier 7: Human-Centric
		"CANVAS": 7, "LINGUA": 7, "SCRIBE": 7, "MENTOR": 7, "BRIDGE": 7,
		// Tier 8: Enterprise
		"AEGIS": 8, "LEDGER": 8, "PULSE": 8, "ARBITER": 8, "ORACLE": 8,
	}
}

// ExecuteWithMemory runs the full ReMem loop for an agent invocation.
func (c *ReMemController) ExecuteWithMemory(
	ctx context.Context,
	agentID string,
	request *models.CopilotRequest,
	executor AgentExecutor,
) (*models.CopilotResponse, error) {
	tierID := c.getAgentTier(agentID)

	// =========================================================================
	// Phase 1: RETRIEVE - Sub-linear experience retrieval
	// =========================================================================
	queryCtx := c.buildQueryContext(agentID, tierID, request)
	retrievalResult, err := c.retriever.Retrieve(queryCtx)
	if err != nil {
		// Non-fatal: continue without memory augmentation
		retrievalResult = &RetrievalResult{Experiences: []*ExperienceTuple{}}
	}

	// Detect retrieval-based impasses
	goalID := fmt.Sprintf("goal-%s-%d", agentID, time.Now().UnixNano())
	if len(retrievalResult.Experiences) == 0 {
		// No match impasse
		impasse := c.impasseDetector.DetectNoMatch(goalID, "No relevant experiences found")
		if impasse != nil {
			// Attempt to resolve
			if _, err := c.impasseDetector.Resolve(impasse.ID); err == nil {
				// Resolution successful - could modify retrieval strategy
			}
		}
	} else if len(retrievalResult.Experiences) >= 2 {
		// Check for tie
		scores := make([]float64, len(retrievalResult.Experiences))
		candidates := make([]string, len(retrievalResult.Experiences))
		for i, exp := range retrievalResult.Experiences {
			scores[i] = exp.FitnessScore
			candidates[i] = exp.AgentID
		}
		// Tie detection is handled internally by DetectTie
		if impasse := c.impasseDetector.DetectTie(goalID, candidates, scores); impasse != nil {
			// Resolution will select among candidates
			c.impasseDetector.Resolve(impasse.ID)
		}
	}

	// Get tier-shared experiences
	tierExperiences := c.getTierExperiences(agentID, tierID, queryCtx)

	// Get collective breakthroughs
	breakthroughs := c.getApplicableBreakthroughs(tierID)

	// =========================================================================
	// Phase 2: THINK - Build augmented context with experiences
	// =========================================================================
	augmentedCtx := c.contextBuilder.Build(
		request,
		retrievalResult.Experiences,
		tierExperiences,
		breakthroughs,
	)
	augmentedCtx.RetrievalResult = retrievalResult

	// =========================================================================
	// Phase 3: ACT - Execute agent with memory-augmented context
	// =========================================================================
	response, trace, err := executor.Execute(ctx, augmentedCtx)
	if err != nil {
		// Detect execution failure impasse
		goalID := fmt.Sprintf("goal-%s-%d", agentID, time.Now().UnixNano())
		impasse := c.impasseDetector.DetectFailure(goalID, agentID, err.Error())
		if impasse != nil {
			// Attempt to resolve the failure
			if resolution, resolveErr := c.impasseDetector.Resolve(impasse.ID); resolveErr == nil {
				if resolution.Success {
					// Learn from the resolution
					// Could potentially retry with different strategy
				}
			}
		}
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	// =========================================================================
	// Phase 4: REFLECT - Evaluate outcome
	// =========================================================================
	evaluation := c.evaluator.Evaluate(request, response, trace)

	// =========================================================================
	// Phase 5: EVOLVE - Update memory based on outcome
	// =========================================================================
	newExperience := c.createExperience(agentID, tierID, request, response, trace, retrievalResult)

	// Add to consolidation buffer for offline processing
	c.consolidator.AddToBuffer(newExperience)

	// Store in primary memory
	if err := c.updater.AddAndEvolve(newExperience, evaluation); err != nil {
		// Log but don't fail the request
		// In production, this should be logged properly
	}

	// Check if this is a breakthrough worth sharing
	if evaluation.Score >= c.config.BreakthroughThreshold && evaluation.Success {
		c.promoteToBreakthrough(newExperience, tierID)
	}

	return response, nil
}

// buildQueryContext creates a query context for retrieval.
func (c *ReMemController) buildQueryContext(agentID string, tierID int, request *models.CopilotRequest) *QueryContext {
	input := ""
	if len(request.Messages) > 0 {
		input = request.Messages[len(request.Messages)-1].Content
	}

	queryCtx := NewQueryContext(agentID, tierID, input)
	queryCtx.MinFitnessScore = c.config.MinFitnessThreshold

	// Compute embedding if service available
	if c.embeddingService != nil {
		if embedding, err := c.embeddingService.Embed(input); err == nil {
			queryCtx.Embedding = embedding
		}
	}

	return queryCtx
}

// getTierExperiences retrieves experiences from same-tier agents.
func (c *ReMemController) getTierExperiences(agentID string, tierID int, query *QueryContext) []*ExperienceTuple {
	tierExps := c.retriever.GetByTier(tierID)

	// Filter out own experiences and apply query filters
	filtered := make([]*ExperienceTuple, 0)
	for _, exp := range tierExps {
		if exp.AgentID != agentID && exp.FitnessScore >= query.MinFitnessScore {
			filtered = append(filtered, exp)
		}
	}

	// Limit to top 3
	if len(filtered) > 3 {
		filtered = filtered[:3]
	}

	return filtered
}

// getApplicableBreakthroughs returns breakthroughs applicable to the given tier.
func (c *ReMemController) getApplicableBreakthroughs(tierID int) []*Breakthrough {
	c.breakthroughMu.RLock()
	defer c.breakthroughMu.RUnlock()

	applicable := make([]*Breakthrough, 0)
	for _, bt := range c.breakthroughs {
		for _, applicableTier := range bt.ApplicableTiers {
			if applicableTier == tierID {
				applicable = append(applicable, bt)
				break
			}
		}
	}

	// Limit to most recent 2
	if len(applicable) > 2 {
		applicable = applicable[len(applicable)-2:]
	}

	return applicable
}

// createExperience creates a new experience from an execution.
func (c *ReMemController) createExperience(
	agentID string,
	tierID int,
	request *models.CopilotRequest,
	response *models.CopilotResponse,
	trace *ExecutionTrace,
	retrieval *RetrievalResult,
) *ExperienceTuple {
	input := ""
	if len(request.Messages) > 0 {
		input = request.Messages[len(request.Messages)-1].Content
	}

	output := ""
	if len(response.Choices) > 0 {
		output = response.Choices[0].Message.Content
	}

	exp := NewExperienceTuple(agentID, tierID, input, output, trace.Strategy)
	exp.EvolutionGen = c.getCurrentGeneration(agentID)

	// Store which experiences were used
	usedIDs := make([]string, 0, len(retrieval.Experiences))
	for _, e := range retrieval.Experiences {
		usedIDs = append(usedIDs, e.ID)
	}
	exp.Metadata["retrieved_experiences"] = usedIDs
	exp.Metadata["retrieval_method"] = retrieval.RetrievalMethod

	// Compute embedding if service available
	if c.embeddingService != nil {
		if embedding, err := c.embeddingService.Embed(input); err == nil {
			exp.Embedding = embedding
		}
	}

	return exp
}

// promoteToBreakthrough promotes a high-performing experience to collective breakthrough.
func (c *ReMemController) promoteToBreakthrough(exp *ExperienceTuple, originTier int) {
	c.breakthroughMu.Lock()
	defer c.breakthroughMu.Unlock()

	bt := &Breakthrough{
		ID:                 "bt_" + exp.ID,
		OriginExperienceID: exp.ID,
		OriginAgent:        exp.AgentID,
		OriginTier:         originTier,
		Strategy:           exp.Strategy,
		Embedding:          exp.Embedding,
		ApplicableTiers:    c.computeApplicableTiers(exp, originTier),
		DiscoveredAt:       time.Now().UnixNano(),
	}

	c.breakthroughs = append(c.breakthroughs, bt)

	// Limit breakthroughs to prevent unbounded growth
	if len(c.breakthroughs) > 100 {
		c.breakthroughs = c.breakthroughs[len(c.breakthroughs)-100:]
	}
}

// computeApplicableTiers determines which tiers can benefit from a breakthrough.
func (c *ReMemController) computeApplicableTiers(exp *ExperienceTuple, originTier int) []int {
	// Simple heuristic: share with same tier and adjacent tiers
	applicable := []int{originTier}
	if originTier > 1 {
		applicable = append(applicable, originTier-1)
	}
	if originTier < 8 {
		applicable = append(applicable, originTier+1)
	}

	// Foundational breakthroughs (Tier 1) apply to all
	if originTier == 1 {
		applicable = []int{1, 2, 3, 4, 5, 6, 7, 8}
	}

	// Meta breakthroughs (Tier 4) apply to all
	if originTier == 4 {
		applicable = []int{1, 2, 3, 4, 5, 6, 7, 8}
	}

	return applicable
}

// getAgentTier returns the tier for an agent.
func (c *ReMemController) getAgentTier(agentID string) int {
	if tier, ok := c.agentTiers[agentID]; ok {
		return tier
	}
	return 1 // Default to tier 1
}

// getCurrentGeneration returns the current evolution generation for an agent.
func (c *ReMemController) getCurrentGeneration(agentID string) int {
	c.genMu.RLock()
	defer c.genMu.RUnlock()
	return c.generationCounter[agentID]
}

// IncrementGeneration advances the evolution generation for an agent.
func (c *ReMemController) IncrementGeneration(agentID string) {
	c.genMu.Lock()
	defer c.genMu.Unlock()
	c.generationCounter[agentID]++
}

// GetStats returns memory statistics.
func (c *ReMemController) GetStats() *MemoryStats {
	stats := c.retriever.GetStats()

	c.breakthroughMu.RLock()
	stats.BreakthroughCount = int64(len(c.breakthroughs))
	c.breakthroughMu.RUnlock()

	return stats
}

// GetImpasseStats returns impasse detection statistics.
func (c *ReMemController) GetImpasseStats() *ImpasseStats {
	return c.impasseDetector.GetStats()
}

// GetActiveImpasses returns all currently active impasses.
func (c *ReMemController) GetActiveImpasses() []*Impasse {
	return c.impasseDetector.GetActive()
}

// TriggerConsolidation manually triggers memory consolidation.
func (c *ReMemController) TriggerConsolidation() (*ConsolidationResult, error) {
	return c.consolidator.Consolidate()
}

// GetConsolidationStats returns memory consolidation statistics.
func (c *ReMemController) GetConsolidationStats() *ConsolidationStats {
	return c.consolidator.GetStats()
}

// GetConsolidatedMemories returns all consolidated memories/schemas.
func (c *ReMemController) GetConsolidatedMemories() map[string]*ConsolidatedMemory {
	return c.consolidator.GetConsolidated()
}

// GetRetriever returns the underlying retriever for direct access if needed.
func (c *ReMemController) GetRetriever() *SubLinearRetriever {
	return c.retriever
}

// ============================================================================
// Default Embedding Service (placeholder - should be replaced with real service)
// ============================================================================

// NoOpEmbeddingService is a placeholder that doesn't compute real embeddings.
type NoOpEmbeddingService struct {
	dimension int
}

// NewNoOpEmbeddingService creates a no-op embedding service.
func NewNoOpEmbeddingService(dimension int) *NoOpEmbeddingService {
	return &NoOpEmbeddingService{dimension: dimension}
}

// Embed returns a zero vector (placeholder implementation).
func (s *NoOpEmbeddingService) Embed(text string) ([]float32, error) {
	// In production, this should call a real embedding service
	// For now, return a simple hash-based pseudo-embedding
	embedding := make([]float32, s.dimension)

	// Create a deterministic pseudo-embedding from the text hash
	hash := hashString(text)
	for i := 0; i < s.dimension; i++ {
		// Use different bits of the hash for each dimension
		embedding[i] = float32((hash>>(uint(i)%64))&0xFF) / 255.0
		hash = hash*6364136223846793005 + 1442695040888963407 // LCG
	}

	return embedding, nil
}
