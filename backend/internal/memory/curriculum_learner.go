// Package memory implements the MNEMONIC memory system for the Elite Agent Collective.
// This file implements Curriculum Learning for progressive agent skill development.
package memory

import (
	"container/heap"
	"context"
	"errors"
	"math"
	"sort"
	"sync"
	"time"
)

// =============================================================================
// CURRICULUM LEARNING SYSTEM
// =============================================================================

// CurriculumLearner manages progressive task difficulty for agent skill development
type CurriculumLearner struct {
	mu sync.RWMutex

	// Curriculum state per agent
	agentCurricula map[string]*AgentCurriculum

	// Task pool organized by difficulty
	taskPool *TaskPool

	// Skill tracking
	skillTracker *SkillTracker

	// Configuration
	config *CurriculumConfig

	// Metrics
	metrics *CurriculumMetrics
}

// AgentCurriculum tracks an agent's progress through the curriculum
type AgentCurriculum struct {
	AgentID          string
	CurrentLevel     int
	MasteryScores    map[string]float64 // Skill -> mastery level [0,1]
	ProgressHistory  []ProgressPoint
	CompletedTasks   []string
	CurrentObjective *LearningObjective
	LastActivity     time.Time
	TotalTasksDone   int
	AverageQuality   float64
}

// ProgressPoint tracks learning progress over time
type ProgressPoint struct {
	Timestamp   time.Time
	Level       int
	SkillScores map[string]float64
	TaskID      string
	Quality     float64
}

// LearningObjective defines a specific learning goal
type LearningObjective struct {
	ID             string
	SkillTarget    string
	TargetMastery  float64
	CurrentMastery float64
	RequiredTasks  int
	CompletedTasks int
	Deadline       time.Time
	Priority       float64
}

// TaskPool manages tasks organized by difficulty
type TaskPool struct {
	mu sync.RWMutex

	// Tasks by difficulty level
	levelTasks map[int][]*CurriculumTask

	// Task metadata
	taskMeta map[string]*TaskMetadata

	// Skill requirements index
	skillIndex map[string][]string // Skill -> TaskIDs
}

// CurriculumTask represents a task in the curriculum
type CurriculumTask struct {
	ID              string
	Description     string
	Difficulty      float64 // [0,1] difficulty score
	Level           int     // Discrete level
	RequiredSkills  []string
	TeachesSkills   []string
	Prerequisites   []string // Task IDs that should be completed first
	Embedding       []float64
	EstimatedTime   time.Duration
	SuccessRate     float64 // Historical success rate
	CompletionCount int
}

// TaskMetadata stores additional task information
type TaskMetadata struct {
	TaskID           string
	AverageQuality   float64
	AttemptCount     int
	SuccessCount     int
	LastAttempted    time.Time
	AgentPerformance map[string]float64 // AgentID -> avg quality
}

// SkillTracker monitors skill development
type SkillTracker struct {
	mu sync.RWMutex

	// Skill definitions
	skills map[string]*SkillDefinition

	// Skill dependencies (prerequisite graph)
	dependencies map[string][]string // Skill -> prerequisite skills

	// Agent skill matrices
	agentSkills map[string]map[string]*SkillLevel
}

// SkillDefinition describes a learnable skill
type SkillDefinition struct {
	ID            string
	Name          string
	Description   string
	Category      string
	MaxLevel      int
	Difficulty    float64
	Prerequisites []string
}

// SkillLevel tracks an agent's level in a skill
type SkillLevel struct {
	SkillID       string
	Level         int
	Mastery       float64 // [0,1] within current level
	Experience    int     // Total practice count
	LastPracticed time.Time
	LearningRate  float64 // How fast agent learns this skill
}

// CurriculumConfig holds curriculum learning configuration
type CurriculumConfig struct {
	MaxLevels            int
	MasteryThreshold     float64 // Required to advance level
	DecayRate            float64 // Skill decay over time
	ProgressionRate      float64 // Base rate for level advancement
	DifficultyScaling    float64 // How fast difficulty increases
	TasksPerLevel        int     // Minimum tasks before level up
	EnableAdaptive       bool    // Enable adaptive curriculum
	EnableSpacedPractice bool    // Enable spaced repetition
}

// DefaultCurriculumConfig returns sensible defaults
func DefaultCurriculumConfig() *CurriculumConfig {
	return &CurriculumConfig{
		MaxLevels:            10,
		MasteryThreshold:     0.8,
		DecayRate:            0.01,
		ProgressionRate:      0.1,
		DifficultyScaling:    1.2,
		TasksPerLevel:        5,
		EnableAdaptive:       true,
		EnableSpacedPractice: true,
	}
}

// CurriculumMetrics tracks learning system performance
type CurriculumMetrics struct {
	mu sync.RWMutex

	TotalTasksAssigned  int
	TotalTasksCompleted int
	AverageCompletion   float64
	LevelUpEvents       int
	SkillMasteryEvents  int
	AdaptationCount     int
	AverageLearningTime time.Duration
}

// NewCurriculumLearner creates a new curriculum learning system
func NewCurriculumLearner(config *CurriculumConfig) *CurriculumLearner {
	if config == nil {
		config = DefaultCurriculumConfig()
	}

	return &CurriculumLearner{
		agentCurricula: make(map[string]*AgentCurriculum),
		taskPool:       newTaskPool(),
		skillTracker:   newSkillTracker(),
		config:         config,
		metrics:        &CurriculumMetrics{},
	}
}

func newTaskPool() *TaskPool {
	return &TaskPool{
		levelTasks: make(map[int][]*CurriculumTask),
		taskMeta:   make(map[string]*TaskMetadata),
		skillIndex: make(map[string][]string),
	}
}

func newSkillTracker() *SkillTracker {
	return &SkillTracker{
		skills:       make(map[string]*SkillDefinition),
		dependencies: make(map[string][]string),
		agentSkills:  make(map[string]map[string]*SkillLevel),
	}
}

// RegisterAgent enrolls an agent in the curriculum
func (c *CurriculumLearner) RegisterAgent(agentID string) error {
	if agentID == "" {
		return errors.New("agent ID required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.agentCurricula[agentID]; exists {
		return nil // Already registered
	}

	c.agentCurricula[agentID] = &AgentCurriculum{
		AgentID:         agentID,
		CurrentLevel:    1,
		MasteryScores:   make(map[string]float64),
		ProgressHistory: make([]ProgressPoint, 0),
		CompletedTasks:  make([]string, 0),
		LastActivity:    time.Now(),
	}

	c.skillTracker.mu.Lock()
	c.skillTracker.agentSkills[agentID] = make(map[string]*SkillLevel)
	c.skillTracker.mu.Unlock()

	return nil
}

// AddTask adds a task to the curriculum pool
func (c *CurriculumLearner) AddTask(task *CurriculumTask) error {
	if task == nil || task.ID == "" {
		return errors.New("valid task required")
	}

	c.taskPool.mu.Lock()
	defer c.taskPool.mu.Unlock()

	// Add to level-based index
	c.taskPool.levelTasks[task.Level] = append(c.taskPool.levelTasks[task.Level], task)

	// Add metadata
	c.taskPool.taskMeta[task.ID] = &TaskMetadata{
		TaskID:           task.ID,
		AgentPerformance: make(map[string]float64),
	}

	// Index by skills
	for _, skill := range task.TeachesSkills {
		c.taskPool.skillIndex[skill] = append(c.taskPool.skillIndex[skill], task.ID)
	}

	return nil
}

// RegisterSkill adds a skill to the tracker
func (c *CurriculumLearner) RegisterSkill(skill *SkillDefinition) error {
	if skill == nil || skill.ID == "" {
		return errors.New("valid skill definition required")
	}

	c.skillTracker.mu.Lock()
	defer c.skillTracker.mu.Unlock()

	c.skillTracker.skills[skill.ID] = skill
	c.skillTracker.dependencies[skill.ID] = skill.Prerequisites

	return nil
}

// GetNextTask returns the next recommended task for an agent
func (c *CurriculumLearner) GetNextTask(ctx context.Context, agentID string) (*CurriculumTask, error) {
	c.mu.RLock()
	curriculum, exists := c.agentCurricula[agentID]
	c.mu.RUnlock()

	if !exists {
		return nil, errors.New("agent not registered")
	}

	// Get candidate tasks for current and adjacent levels
	candidates := c.getCandidateTasks(curriculum)

	if len(candidates) == 0 {
		return nil, errors.New("no suitable tasks available")
	}

	// Score and rank tasks
	scored := c.scoreTasksForAgent(candidates, curriculum)

	// Select best task
	if len(scored) == 0 {
		return nil, errors.New("no tasks passed scoring")
	}

	best := scored[0]

	// Update metrics
	c.metrics.mu.Lock()
	c.metrics.TotalTasksAssigned++
	c.metrics.mu.Unlock()

	return best, nil
}

// getCandidateTasks returns tasks appropriate for the agent's level
func (c *CurriculumLearner) getCandidateTasks(curriculum *AgentCurriculum) []*CurriculumTask {
	c.taskPool.mu.RLock()
	defer c.taskPool.mu.RUnlock()

	candidates := make([]*CurriculumTask, 0)
	level := curriculum.CurrentLevel

	// Get tasks from current level, one below, and one above
	for l := level - 1; l <= level+1; l++ {
		if l < 1 {
			continue
		}
		if tasks, ok := c.taskPool.levelTasks[l]; ok {
			for _, task := range tasks {
				// Check if not already completed or if spaced practice
				if !c.isCompleted(curriculum, task.ID) || c.shouldRepeat(curriculum, task) {
					candidates = append(candidates, task)
				}
			}
		}
	}

	return candidates
}

// isCompleted checks if agent has completed a task
func (c *CurriculumLearner) isCompleted(curriculum *AgentCurriculum, taskID string) bool {
	for _, completed := range curriculum.CompletedTasks {
		if completed == taskID {
			return true
		}
	}
	return false
}

// shouldRepeat determines if a task should be repeated (spaced practice)
func (c *CurriculumLearner) shouldRepeat(curriculum *AgentCurriculum, task *CurriculumTask) bool {
	if !c.config.EnableSpacedPractice {
		return false
	}

	c.taskPool.mu.RLock()
	meta, exists := c.taskPool.taskMeta[task.ID]
	c.taskPool.mu.RUnlock()

	if !exists {
		return false
	}

	// Repeat if agent performance was poor
	if perf, ok := meta.AgentPerformance[curriculum.AgentID]; ok {
		if perf < c.config.MasteryThreshold {
			// Check time since last attempt (spaced repetition)
			daysSince := time.Since(meta.LastAttempted).Hours() / 24
			return daysSince >= float64(curriculum.CurrentLevel) // Longer spacing at higher levels
		}
	}

	return false
}

// ScoredTask pairs a task with its score
type ScoredTask struct {
	Task  *CurriculumTask
	Score float64
}

// scoreTasksForAgent scores and sorts tasks by suitability
func (c *CurriculumLearner) scoreTasksForAgent(tasks []*CurriculumTask, curriculum *AgentCurriculum) []*CurriculumTask {
	scored := make([]ScoredTask, 0, len(tasks))

	for _, task := range tasks {
		score := c.computeTaskScore(task, curriculum)
		if score > 0 {
			scored = append(scored, ScoredTask{Task: task, Score: score})
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Extract tasks
	result := make([]*CurriculumTask, len(scored))
	for i, s := range scored {
		result[i] = s.Task
	}

	return result
}

// computeTaskScore calculates how suitable a task is for an agent
func (c *CurriculumLearner) computeTaskScore(task *CurriculumTask, curriculum *AgentCurriculum) float64 {
	score := 1.0

	// Factor 1: Level appropriateness (higher score for matching level)
	levelDiff := math.Abs(float64(task.Level - curriculum.CurrentLevel))
	score *= math.Exp(-levelDiff * 0.5)

	// Factor 2: Skill gap (prioritize tasks that teach weak skills)
	skillGapBonus := 0.0
	for _, skill := range task.TeachesSkills {
		if mastery, ok := curriculum.MasteryScores[skill]; ok {
			// Bonus for skills below threshold
			if mastery < c.config.MasteryThreshold {
				skillGapBonus += (c.config.MasteryThreshold - mastery)
			}
		} else {
			// New skill - high bonus
			skillGapBonus += 0.5
		}
	}
	score *= (1.0 + skillGapBonus)

	// Factor 3: Prerequisites met
	prereqsMet := c.checkPrerequisites(task, curriculum)
	if !prereqsMet {
		score *= 0.1 // Strong penalty but not zero
	}

	// Factor 4: Zone of proximal development
	if c.config.EnableAdaptive {
		zpd := c.computeZPD(task, curriculum)
		score *= zpd
	}

	// Factor 5: Historical success rate (prefer doable tasks)
	if task.SuccessRate > 0 {
		// Ideal success rate is around 70-80% (challenging but achievable)
		optimal := 0.75
		deviation := math.Abs(task.SuccessRate - optimal)
		score *= (1.0 - deviation*0.5)
	}

	return score
}

// checkPrerequisites verifies task prerequisites are met
func (c *CurriculumLearner) checkPrerequisites(task *CurriculumTask, curriculum *AgentCurriculum) bool {
	for _, prereq := range task.Prerequisites {
		if !c.isCompleted(curriculum, prereq) {
			return false
		}
	}

	// Check skill prerequisites
	for _, skill := range task.RequiredSkills {
		if mastery, ok := curriculum.MasteryScores[skill]; ok {
			if mastery < 0.3 { // Minimum required mastery
				return false
			}
		}
	}

	return true
}

// computeZPD calculates zone of proximal development score
func (c *CurriculumLearner) computeZPD(task *CurriculumTask, curriculum *AgentCurriculum) float64 {
	// ZPD: sweet spot between too easy and too hard
	// Based on agent's average quality and task difficulty

	avgQuality := curriculum.AverageQuality
	if avgQuality == 0 {
		avgQuality = 0.5
	}

	// Target difficulty is slightly above current performance
	targetDiff := avgQuality + 0.1
	if targetDiff > 1.0 {
		targetDiff = 0.95
	}

	// Score based on how close task difficulty is to target
	diff := math.Abs(task.Difficulty - targetDiff)
	return math.Exp(-diff * 3)
}

// RecordTaskCompletion records that an agent completed a task
func (c *CurriculumLearner) RecordTaskCompletion(agentID, taskID string, quality float64) error {
	c.mu.Lock()
	curriculum, exists := c.agentCurricula[agentID]
	if !exists {
		c.mu.Unlock()
		return errors.New("agent not registered")
	}

	// Update completion list
	curriculum.CompletedTasks = append(curriculum.CompletedTasks, taskID)
	curriculum.TotalTasksDone++
	curriculum.LastActivity = time.Now()

	// Update average quality
	n := float64(curriculum.TotalTasksDone)
	curriculum.AverageQuality = ((n-1)*curriculum.AverageQuality + quality) / n

	// Record progress point
	curriculum.ProgressHistory = append(curriculum.ProgressHistory, ProgressPoint{
		Timestamp:   time.Now(),
		Level:       curriculum.CurrentLevel,
		SkillScores: copySkillScores(curriculum.MasteryScores),
		TaskID:      taskID,
		Quality:     quality,
	})

	c.mu.Unlock()

	// Update task metadata
	c.taskPool.mu.Lock()
	if meta, ok := c.taskPool.taskMeta[taskID]; ok {
		meta.AttemptCount++
		if quality >= c.config.MasteryThreshold {
			meta.SuccessCount++
		}
		meta.AverageQuality = ((float64(meta.AttemptCount-1) * meta.AverageQuality) + quality) / float64(meta.AttemptCount)
		meta.LastAttempted = time.Now()
		meta.AgentPerformance[agentID] = quality
	}
	c.taskPool.mu.Unlock()

	// Update skill mastery
	c.updateSkillMastery(agentID, taskID, quality)

	// Check for level advancement
	c.checkLevelAdvancement(agentID)

	// Update metrics
	c.metrics.mu.Lock()
	c.metrics.TotalTasksCompleted++
	c.metrics.AverageCompletion = ((float64(c.metrics.TotalTasksCompleted-1) * c.metrics.AverageCompletion) + quality) / float64(c.metrics.TotalTasksCompleted)
	c.metrics.mu.Unlock()

	return nil
}

func copySkillScores(m map[string]float64) map[string]float64 {
	result := make(map[string]float64)
	for k, v := range m {
		result[k] = v
	}
	return result
}

// updateSkillMastery updates skill levels based on task completion
func (c *CurriculumLearner) updateSkillMastery(agentID, taskID string, quality float64) {
	// Find task to get skills
	c.taskPool.mu.RLock()
	var task *CurriculumTask
	for _, tasks := range c.taskPool.levelTasks {
		for _, t := range tasks {
			if t.ID == taskID {
				task = t
				break
			}
		}
		if task != nil {
			break
		}
	}
	c.taskPool.mu.RUnlock()

	if task == nil {
		return
	}

	c.mu.Lock()
	curriculum := c.agentCurricula[agentID]

	// Update mastery for each skill the task teaches
	for _, skillID := range task.TeachesSkills {
		current := curriculum.MasteryScores[skillID]
		// Learning curve: diminishing returns as mastery increases
		learningGain := quality * c.config.ProgressionRate * (1.0 - current)
		newMastery := current + learningGain
		if newMastery > 1.0 {
			newMastery = 1.0
		}
		curriculum.MasteryScores[skillID] = newMastery

		// Check for mastery event
		if current < c.config.MasteryThreshold && newMastery >= c.config.MasteryThreshold {
			c.metrics.mu.Lock()
			c.metrics.SkillMasteryEvents++
			c.metrics.mu.Unlock()
		}
	}
	c.mu.Unlock()

	// Update skill tracker
	c.skillTracker.mu.Lock()
	if agentSkills, ok := c.skillTracker.agentSkills[agentID]; ok {
		for _, skillID := range task.TeachesSkills {
			if sl, ok := agentSkills[skillID]; ok {
				sl.Experience++
				sl.LastPracticed = time.Now()
				sl.Mastery += quality * 0.1 * (1 - sl.Mastery)
				if sl.Mastery > 1.0 {
					sl.Mastery = 1.0
				}
			} else {
				agentSkills[skillID] = &SkillLevel{
					SkillID:       skillID,
					Level:         1,
					Mastery:       quality * 0.1,
					Experience:    1,
					LastPracticed: time.Now(),
					LearningRate:  1.0,
				}
			}
		}
	}
	c.skillTracker.mu.Unlock()
}

// checkLevelAdvancement checks if agent should advance a level
func (c *CurriculumLearner) checkLevelAdvancement(agentID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	curriculum, exists := c.agentCurricula[agentID]
	if !exists {
		return
	}

	// Conditions for advancement:
	// 1. Minimum tasks at current level
	tasksAtLevel := c.countTasksAtLevel(curriculum, curriculum.CurrentLevel)
	if tasksAtLevel < c.config.TasksPerLevel {
		return
	}

	// 2. Average quality above threshold
	if curriculum.AverageQuality < c.config.MasteryThreshold {
		return
	}

	// 3. Key skills above threshold
	skillsMastered := 0
	totalSkills := len(curriculum.MasteryScores)
	for _, mastery := range curriculum.MasteryScores {
		if mastery >= c.config.MasteryThreshold {
			skillsMastered++
		}
	}

	if totalSkills > 0 && float64(skillsMastered)/float64(totalSkills) < 0.6 {
		return
	}

	// Advance level
	if curriculum.CurrentLevel < c.config.MaxLevels {
		curriculum.CurrentLevel++
		c.metrics.mu.Lock()
		c.metrics.LevelUpEvents++
		c.metrics.mu.Unlock()
	}
}

// countTasksAtLevel counts completed tasks at a specific level
func (c *CurriculumLearner) countTasksAtLevel(curriculum *AgentCurriculum, level int) int {
	count := 0

	c.taskPool.mu.RLock()
	defer c.taskPool.mu.RUnlock()

	levelTasks := make(map[string]bool)
	if tasks, ok := c.taskPool.levelTasks[level]; ok {
		for _, t := range tasks {
			levelTasks[t.ID] = true
		}
	}

	for _, completed := range curriculum.CompletedTasks {
		if levelTasks[completed] {
			count++
		}
	}

	return count
}

// GetAgentProgress returns an agent's curriculum progress
func (c *CurriculumLearner) GetAgentProgress(agentID string) (*AgentCurriculum, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	curriculum, exists := c.agentCurricula[agentID]
	if !exists {
		return nil, errors.New("agent not registered")
	}

	// Return a copy
	return &AgentCurriculum{
		AgentID:        curriculum.AgentID,
		CurrentLevel:   curriculum.CurrentLevel,
		MasteryScores:  copySkillScores(curriculum.MasteryScores),
		CompletedTasks: append([]string{}, curriculum.CompletedTasks...),
		TotalTasksDone: curriculum.TotalTasksDone,
		AverageQuality: curriculum.AverageQuality,
		LastActivity:   curriculum.LastActivity,
	}, nil
}

// GetMetrics returns curriculum learning metrics
func (c *CurriculumLearner) GetMetrics() *CurriculumMetrics {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	return &CurriculumMetrics{
		TotalTasksAssigned:  c.metrics.TotalTasksAssigned,
		TotalTasksCompleted: c.metrics.TotalTasksCompleted,
		AverageCompletion:   c.metrics.AverageCompletion,
		LevelUpEvents:       c.metrics.LevelUpEvents,
		SkillMasteryEvents:  c.metrics.SkillMasteryEvents,
		AdaptationCount:     c.metrics.AdaptationCount,
	}
}

// ApplyDecay applies skill decay to all agents (call periodically)
func (c *CurriculumLearner) ApplyDecay() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, curriculum := range c.agentCurricula {
		daysSinceActivity := time.Since(curriculum.LastActivity).Hours() / 24
		decayFactor := math.Exp(-c.config.DecayRate * daysSinceActivity)

		for skill, mastery := range curriculum.MasteryScores {
			curriculum.MasteryScores[skill] = mastery * decayFactor
		}
	}
}

// =============================================================================
// TASK PRIORITY QUEUE (for task scheduling)
// =============================================================================

// TaskPriorityQueue implements a priority queue for tasks
type TaskPriorityQueue struct {
	items []*priorityItem
}

type priorityItem struct {
	task     *CurriculumTask
	priority float64
	index    int
}

func (pq TaskPriorityQueue) Len() int { return len(pq.items) }

func (pq TaskPriorityQueue) Less(i, j int) bool {
	return pq.items[i].priority > pq.items[j].priority // Higher priority first
}

func (pq TaskPriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

func (pq *TaskPriorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*priorityItem)
	item.index = n
	pq.items = append(pq.items, item)
}

func (pq *TaskPriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	pq.items = old[0 : n-1]
	return item
}

// NewTaskPriorityQueue creates a new priority queue
func NewTaskPriorityQueue() *TaskPriorityQueue {
	pq := &TaskPriorityQueue{items: make([]*priorityItem, 0)}
	heap.Init(pq)
	return pq
}

// Add adds a task with priority
func (pq *TaskPriorityQueue) Add(task *CurriculumTask, priority float64) {
	heap.Push(pq, &priorityItem{task: task, priority: priority})
}

// GetNext returns and removes the highest priority task
func (pq *TaskPriorityQueue) GetNext() *CurriculumTask {
	if len(pq.items) == 0 {
		return nil
	}
	item := heap.Pop(pq).(*priorityItem)
	return item.task
}

// Peek returns the highest priority task without removing it
func (pq *TaskPriorityQueue) Peek() *CurriculumTask {
	if len(pq.items) == 0 {
		return nil
	}
	return pq.items[0].task
}

// IsEmpty returns true if queue is empty
func (pq *TaskPriorityQueue) IsEmpty() bool {
	return len(pq.items) == 0
}
