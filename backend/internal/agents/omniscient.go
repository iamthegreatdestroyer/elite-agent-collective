// Package agents provides the agent registry and HTTP handlers.
package agents

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/memory"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/metrics"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// omniscientMaxAgents caps how many specialists OMNISCIENT will orchestrate for
// a single request, protecting the CPU-only box from over-splitting.
const omniscientMaxAgents = 4

// OmniscientAgent is the meta-orchestrator. It decomposes a request into
// subtasks, drives the StrategicPlanner to build (and feasibility-check) a
// plan, maps each subtask to a specialist via the semantic router, and
// orchestrates the chosen agents through the existing pipeline. Fail-open: any
// planning or orchestration failure degrades to a single APEX answer.
type OmniscientAgent struct {
	info     models.Agent
	upstream *copilot.UpstreamClient
	registry *Registry
	crews    *CrewRegistry
	planner  *memory.StrategicPlanner

	once     sync.Once
	pipeline *AgentPipeline
	router   *SemanticRouter
}

// NewOmniscientAgent builds the meta-orchestrator over the given registry.
func NewOmniscientAgent(info models.Agent, uc *copilot.UpstreamClient, reg *Registry) *OmniscientAgent {
	return &OmniscientAgent{
		info:     info,
		upstream: uc,
		registry: reg,
		planner:  memory.NewStrategicPlanner(memory.PlanningConfig{PlanCachingEnabled: true, MaxLookaheadDepth: 3}),
	}
}

// GetInfo returns OMNISCIENT's metadata.
func (o *OmniscientAgent) GetInfo() models.Agent { return o.info }

// SetCrews attaches the crew registry (via Registry.InjectCrews) so OMNISCIENT
// can fall back to a configured crew when routing yields no specialist.
func (o *OmniscientAgent) SetCrews(c *CrewRegistry) { o.crews = c }

// SetOllama and SetMemory are accepted for injection parity but unused:
// OMNISCIENT delegates real work to the specialists it orchestrates.
func (o *OmniscientAgent) SetOllama(*copilot.OllamaClient) {}
func (o *OmniscientAgent) SetMemory(*memory.Store)         {}

func (o *OmniscientAgent) lazyInit() {
	o.once.Do(func() {
		o.pipeline = NewAgentPipeline(o.registry)
		o.router = NewSemanticRouter(o.registry)
	})
}

// Handle decomposes, plans, and orchestrates a request.
func (o *OmniscientAgent) Handle(ctx context.Context, req *models.CopilotRequest) (*models.CopilotResponse, error) {
	metrics.IncAgent("OMNISCIENT")
	o.lazyInit()

	userMsg := copilot.GetLastUserMessage(req)
	if strings.TrimSpace(userMsg) == "" {
		return copilot.NewResponse("[OMNISCIENT] No task provided."), nil
	}

	subtasks := omniDecompose(userMsg)
	agentFor := func(st string) string {
		if r := o.router.Route(st); len(r) > 0 {
			return r[0]
		}
		return ""
	}

	goal := &memory.Goal{
		ID:          omniGoalID(userMsg),
		Name:        omniTruncate(userMsg, 80),
		Description: userMsg,
		Priority:    memory.PriorityHigh,
		Status:      memory.GoalActive,
		CreatedAt:   time.Now(),
		Context:     map[string]interface{}{"source": "OMNISCIENT"},
	}

	// Genuinely drive the planner: it builds actions per subtask, checks
	// feasibility, expands a lookahead tree, and caches — see
	// StrategicPlanner.CreatePlanForSubtasks.
	plan, perr := o.planner.CreatePlanForSubtasks(ctx, goal, subtasks, agentFor)

	// Derive an ordered, unique specialist set (never OMNISCIENT itself, to
	// avoid self-recursion). Prefer the plan's agent assignments; fall back to
	// routing each subtask, then a configured crew, then APEX.
	var codenames []string
	seen := map[string]bool{}
	addCode := func(c string) {
		c = strings.ToUpper(strings.TrimSpace(c))
		if c == "" || c == "OMNISCIENT" || seen[c] || len(codenames) >= omniscientMaxAgents {
			return
		}
		seen[c] = true
		codenames = append(codenames, c)
	}
	if perr == nil && plan != nil {
		for _, a := range plan.Actions {
			addCode(a.AgentRequired)
		}
	}
	if len(codenames) == 0 {
		for _, st := range subtasks {
			addCode(agentFor(st))
		}
	}
	if len(codenames) == 0 && o.crews != nil {
		for _, name := range o.crews.List() {
			if crew, ok := o.crews.Get(name); ok && len(crew.Agents) > 0 {
				for _, c := range crew.Agents {
					addCode(c)
				}
				break
			}
		}
	}
	if len(codenames) == 0 {
		codenames = []string{"APEX"}
	}

	// Orchestrate through the existing machinery.
	var body string
	if len(codenames) == 1 {
		ag, err := o.registry.Get(codenames[0])
		if err != nil {
			return o.fallbackAPEX(ctx, req)
		}
		inner, herr := ag.Handle(ctx, req)
		if herr != nil || inner == nil || len(inner.Choices) == 0 {
			return o.fallbackAPEX(ctx, req)
		}
		body = inner.Choices[0].Message.Content
	} else {
		presp, perr2 := o.pipeline.Execute(ctx, req, codenames)
		if perr2 != nil || presp == nil || len(presp.Choices) == 0 {
			return o.fallbackAPEX(ctx, req)
		}
		body = presp.Choices[0].Message.Content
	}

	return copilot.NewResponse(o.planHeader(codenames, plan) + body), nil
}

func (o *OmniscientAgent) fallbackAPEX(ctx context.Context, req *models.CopilotRequest) (*models.CopilotResponse, error) {
	ag, err := o.registry.Get("APEX")
	if err != nil {
		return copilot.NewResponse("[OMNISCIENT] No agents available."), nil
	}
	return ag.Handle(ctx, req)
}

// planHeader summarizes the orchestration decision and the planner's own
// evaluation of the plan.
func (o *OmniscientAgent) planHeader(codenames []string, plan *memory.Plan) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[OMNISCIENT] Orchestrated %d specialist(s): %s\n", len(codenames), strings.Join(codenames, " -> "))
	if plan != nil {
		score := 0.0
		if best := o.planner.GetBestStrategy(o.planner.GetLookaheadTree()); best != nil {
			score = best.Score
		}
		fmt.Fprintf(&sb, "[OMNISCIENT] Plan: %d action(s), feasible=%t, cost=%.2f, lookahead-score=%.2f\n\n",
			len(plan.Actions), plan.Feasible, plan.TotalCost, score)
	} else {
		sb.WriteString("\n")
	}
	return sb.String()
}

// omniDecompose splits a request into subtasks on sentence/clause boundaries,
// capped at omniscientMaxAgents.
func omniDecompose(msg string) []string {
	repl := strings.NewReplacer("\n", "|", ".", "|", "?", "|", "!", "|", ";", "|", " and ", "|", " then ", "|")
	parts := strings.Split(repl.Replace(msg), "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) < 3 {
			continue
		}
		out = append(out, p)
		if len(out) >= omniscientMaxAgents {
			break
		}
	}
	if len(out) == 0 {
		out = []string{strings.TrimSpace(msg)}
	}
	return out
}

func omniTruncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n]
	}
	return s
}

func omniGoalID(msg string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(msg))
	return fmt.Sprintf("goal-%x", h.Sum64())
}
