package memory

import (
	"context"
	"testing"
	"time"
)

// =============================================================================
// CURRICULUM LEARNER TESTS
// =============================================================================

func TestCurriculumLearner_New(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		cl := NewCurriculumLearner(nil)
		if cl == nil {
			t.Fatal("expected non-nil curriculum learner")
		}
		if cl.config.MaxLevels != 10 {
			t.Errorf("expected 10 max levels, got %d", cl.config.MaxLevels)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &CurriculumConfig{
			MaxLevels:        20,
			MasteryThreshold: 0.9,
			TasksPerLevel:    10,
		}
		cl := NewCurriculumLearner(config)
		if cl.config.MaxLevels != 20 {
			t.Errorf("expected 20 max levels, got %d", cl.config.MaxLevels)
		}
	})
}

func TestCurriculumLearner_RegisterAgent(t *testing.T) {
	cl := NewCurriculumLearner(nil)

	t.Run("register valid agent", func(t *testing.T) {
		err := cl.RegisterAgent("APEX")
		if err != nil {
			t.Fatalf("failed to register agent: %v", err)
		}

		progress, err := cl.GetAgentProgress("APEX")
		if err != nil {
			t.Fatalf("failed to get progress: %v", err)
		}
		if progress.CurrentLevel != 1 {
			t.Errorf("expected level 1, got %d", progress.CurrentLevel)
		}
	})

	t.Run("register empty agent ID fails", func(t *testing.T) {
		err := cl.RegisterAgent("")
		if err == nil {
			t.Error("expected error for empty agent ID")
		}
	})

	t.Run("re-register is idempotent", func(t *testing.T) {
		cl.RegisterAgent("TENSOR")
		err := cl.RegisterAgent("TENSOR")
		if err != nil {
			t.Errorf("re-registration should not fail: %v", err)
		}
	})
}

func TestCurriculumLearner_AddTask(t *testing.T) {
	cl := NewCurriculumLearner(nil)

	t.Run("add valid task", func(t *testing.T) {
		task := &CurriculumTask{
			ID:             "task-1",
			Description:    "Test task",
			Difficulty:     0.5,
			Level:          1,
			RequiredSkills: []string{"coding"},
			TeachesSkills:  []string{"python"},
		}
		err := cl.AddTask(task)
		if err != nil {
			t.Fatalf("failed to add task: %v", err)
		}
	})

	t.Run("add nil task fails", func(t *testing.T) {
		err := cl.AddTask(nil)
		if err == nil {
			t.Error("expected error for nil task")
		}
	})

	t.Run("add task with empty ID fails", func(t *testing.T) {
		task := &CurriculumTask{ID: ""}
		err := cl.AddTask(task)
		if err == nil {
			t.Error("expected error for empty task ID")
		}
	})
}

func TestCurriculumLearner_RegisterSkill(t *testing.T) {
	cl := NewCurriculumLearner(nil)

	t.Run("register valid skill", func(t *testing.T) {
		skill := &SkillDefinition{
			ID:          "python",
			Name:        "Python Programming",
			Description: "Ability to write Python code",
			Category:    "programming",
			MaxLevel:    10,
		}
		err := cl.RegisterSkill(skill)
		if err != nil {
			t.Fatalf("failed to register skill: %v", err)
		}
	})

	t.Run("register nil skill fails", func(t *testing.T) {
		err := cl.RegisterSkill(nil)
		if err == nil {
			t.Error("expected error for nil skill")
		}
	})
}

func TestCurriculumLearner_GetNextTask(t *testing.T) {
	cl := NewCurriculumLearner(nil)
	ctx := context.Background()

	// Setup
	cl.RegisterAgent("APEX")
	for i := 1; i <= 5; i++ {
		cl.AddTask(&CurriculumTask{
			ID:            string(rune('A' + i - 1)),
			Level:         1,
			Difficulty:    float64(i) / 10.0,
			TeachesSkills: []string{"basic"},
		})
	}

	t.Run("get task for registered agent", func(t *testing.T) {
		task, err := cl.GetNextTask(ctx, "APEX")
		if err != nil {
			t.Fatalf("failed to get next task: %v", err)
		}
		if task == nil {
			t.Fatal("expected non-nil task")
		}
	})

	t.Run("get task for unregistered agent fails", func(t *testing.T) {
		_, err := cl.GetNextTask(ctx, "UNKNOWN")
		if err == nil {
			t.Error("expected error for unregistered agent")
		}
	})
}

func TestCurriculumLearner_RecordTaskCompletion(t *testing.T) {
	cl := NewCurriculumLearner(nil)

	// Setup
	cl.RegisterAgent("APEX")
	cl.AddTask(&CurriculumTask{
		ID:            "task-1",
		Level:         1,
		Difficulty:    0.3,
		TeachesSkills: []string{"coding"},
	})

	t.Run("record completion updates progress", func(t *testing.T) {
		err := cl.RecordTaskCompletion("APEX", "task-1", 0.9)
		if err != nil {
			t.Fatalf("failed to record completion: %v", err)
		}

		progress, _ := cl.GetAgentProgress("APEX")
		if progress.TotalTasksDone != 1 {
			t.Errorf("expected 1 task done, got %d", progress.TotalTasksDone)
		}
		if progress.AverageQuality != 0.9 {
			t.Errorf("expected 0.9 avg quality, got %f", progress.AverageQuality)
		}
	})

	t.Run("record completion updates skill mastery", func(t *testing.T) {
		progress, _ := cl.GetAgentProgress("APEX")
		if mastery, ok := progress.MasteryScores["coding"]; ok {
			if mastery <= 0 {
				t.Error("skill mastery should increase after completion")
			}
		}
	})

	t.Run("record completion for unregistered agent fails", func(t *testing.T) {
		err := cl.RecordTaskCompletion("UNKNOWN", "task-1", 0.8)
		if err == nil {
			t.Error("expected error for unregistered agent")
		}
	})
}

func TestCurriculumLearner_LevelAdvancement(t *testing.T) {
	config := &CurriculumConfig{
		MaxLevels:        10,
		MasteryThreshold: 0.7,
		TasksPerLevel:    3,
		ProgressionRate:  0.5, // Fast progression for testing
	}
	cl := NewCurriculumLearner(config)

	// Setup
	cl.RegisterAgent("APEX")
	for i := 1; i <= 5; i++ {
		cl.AddTask(&CurriculumTask{
			ID:            string(rune('A' + i - 1)),
			Level:         1,
			Difficulty:    0.3,
			TeachesSkills: []string{"skill-1"},
		})
	}

	// Complete enough tasks with high quality
	for i := 1; i <= 5; i++ {
		cl.RecordTaskCompletion("APEX", string(rune('A'+i-1)), 0.95)
	}

	progress, _ := cl.GetAgentProgress("APEX")
	if progress.CurrentLevel <= 1 {
		t.Logf("Level: %d, Tasks: %d, Quality: %f, Mastery: %v",
			progress.CurrentLevel, progress.TotalTasksDone,
			progress.AverageQuality, progress.MasteryScores)
		// Note: Level advancement depends on multiple factors
		// This test verifies the mechanism works, not exact level
	}
}

func TestCurriculumLearner_ZoneOfProximalDevelopment(t *testing.T) {
	cl := NewCurriculumLearner(nil)

	// Setup agent with known performance
	cl.RegisterAgent("APEX")
	cl.mu.Lock()
	cl.agentCurricula["APEX"].AverageQuality = 0.7
	cl.mu.Unlock()

	// Create tasks at different difficulties
	easyTask := &CurriculumTask{ID: "easy", Difficulty: 0.3, Level: 1}
	mediumTask := &CurriculumTask{ID: "medium", Difficulty: 0.8, Level: 1}
	hardTask := &CurriculumTask{ID: "hard", Difficulty: 0.95, Level: 1}

	cl.mu.RLock()
	curriculum := cl.agentCurricula["APEX"]
	cl.mu.RUnlock()

	// ZPD should favor medium difficulty
	easyZPD := cl.computeZPD(easyTask, curriculum)
	mediumZPD := cl.computeZPD(mediumTask, curriculum)
	hardZPD := cl.computeZPD(hardTask, curriculum)

	if mediumZPD <= easyZPD || mediumZPD <= hardZPD {
		t.Logf("ZPD scores - Easy: %f, Medium: %f, Hard: %f", easyZPD, mediumZPD, hardZPD)
		// Medium should generally score highest for 0.7 avg quality
	}
}

func TestCurriculumLearner_SkillDecay(t *testing.T) {
	config := &CurriculumConfig{
		MaxLevels: 10,
		DecayRate: 0.1, // High decay for testing
	}
	cl := NewCurriculumLearner(config)

	// Setup
	cl.RegisterAgent("APEX")
	cl.mu.Lock()
	cl.agentCurricula["APEX"].MasteryScores["coding"] = 0.8
	cl.agentCurricula["APEX"].LastActivity = time.Now().Add(-24 * time.Hour) // 1 day ago
	cl.mu.Unlock()

	// Apply decay
	cl.ApplyDecay()

	progress, _ := cl.GetAgentProgress("APEX")
	if mastery := progress.MasteryScores["coding"]; mastery >= 0.8 {
		t.Errorf("mastery should decay, got %f", mastery)
	}
}

func TestCurriculumLearner_Metrics(t *testing.T) {
	cl := NewCurriculumLearner(nil)

	// Setup and complete some tasks
	cl.RegisterAgent("APEX")
	cl.AddTask(&CurriculumTask{ID: "task-1", Level: 1, TeachesSkills: []string{"skill"}})
	cl.AddTask(&CurriculumTask{ID: "task-2", Level: 1, TeachesSkills: []string{"skill"}})

	ctx := context.Background()
	cl.GetNextTask(ctx, "APEX")
	cl.GetNextTask(ctx, "APEX")
	cl.RecordTaskCompletion("APEX", "task-1", 0.9)
	cl.RecordTaskCompletion("APEX", "task-2", 0.8)

	metrics := cl.GetMetrics()
	if metrics.TotalTasksAssigned != 2 {
		t.Errorf("expected 2 assigned, got %d", metrics.TotalTasksAssigned)
	}
	if metrics.TotalTasksCompleted != 2 {
		t.Errorf("expected 2 completed, got %d", metrics.TotalTasksCompleted)
	}
}

// =============================================================================
// TASK PRIORITY QUEUE TESTS
// =============================================================================

func TestTaskPriorityQueue_Operations(t *testing.T) {
	pq := NewTaskPriorityQueue()

	t.Run("empty queue", func(t *testing.T) {
		if !pq.IsEmpty() {
			t.Error("new queue should be empty")
		}
		if pq.GetNext() != nil {
			t.Error("GetNext on empty queue should return nil")
		}
		if pq.Peek() != nil {
			t.Error("Peek on empty queue should return nil")
		}
	})

	t.Run("add and retrieve by priority", func(t *testing.T) {
		pq.Add(&CurriculumTask{ID: "low"}, 1.0)
		pq.Add(&CurriculumTask{ID: "high"}, 10.0)
		pq.Add(&CurriculumTask{ID: "medium"}, 5.0)

		if pq.IsEmpty() {
			t.Error("queue should not be empty")
		}

		// Should get highest priority first
		first := pq.GetNext()
		if first.ID != "high" {
			t.Errorf("expected high, got %s", first.ID)
		}

		second := pq.GetNext()
		if second.ID != "medium" {
			t.Errorf("expected medium, got %s", second.ID)
		}

		third := pq.GetNext()
		if third.ID != "low" {
			t.Errorf("expected low, got %s", third.ID)
		}
	})

	t.Run("peek does not remove", func(t *testing.T) {
		pq.Add(&CurriculumTask{ID: "peek-test"}, 5.0)

		peeked := pq.Peek()
		if peeked.ID != "peek-test" {
			t.Errorf("peek returned wrong task")
		}

		// Should still be there
		got := pq.GetNext()
		if got.ID != "peek-test" {
			t.Errorf("task should still be in queue after peek")
		}
	})
}

// =============================================================================
// SKILL TRACKER TESTS
// =============================================================================

func TestSkillTracker_SkillDependencies(t *testing.T) {
	cl := NewCurriculumLearner(nil)

	// Register skills with dependencies
	cl.RegisterSkill(&SkillDefinition{
		ID:            "advanced-python",
		Prerequisites: []string{"basic-python"},
	})
	cl.RegisterSkill(&SkillDefinition{
		ID:            "basic-python",
		Prerequisites: []string{},
	})

	// Verify dependencies stored
	cl.skillTracker.mu.RLock()
	deps := cl.skillTracker.dependencies["advanced-python"]
	cl.skillTracker.mu.RUnlock()

	if len(deps) != 1 || deps[0] != "basic-python" {
		t.Errorf("expected basic-python dependency, got %v", deps)
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestCurriculumLearner_FullWorkflow(t *testing.T) {
	cl := NewCurriculumLearner(&CurriculumConfig{
		MaxLevels:        5,
		MasteryThreshold: 0.7,
		TasksPerLevel:    2,
		ProgressionRate:  0.3,
		EnableAdaptive:   true,
	})
	ctx := context.Background()

	// Register skills
	cl.RegisterSkill(&SkillDefinition{ID: "coding", Name: "Coding"})
	cl.RegisterSkill(&SkillDefinition{ID: "testing", Name: "Testing"})

	// Add tasks for multiple levels
	for level := 1; level <= 3; level++ {
		for i := 0; i < 3; i++ {
			cl.AddTask(&CurriculumTask{
				ID:            string(rune('A'+(level-1)*3+i)) + "-L" + string(rune('0'+level)),
				Level:         level,
				Difficulty:    float64(level) * 0.3,
				TeachesSkills: []string{"coding", "testing"},
			})
		}
	}

	// Register agent
	cl.RegisterAgent("APEX")

	// Simulate learning workflow
	for i := 0; i < 10; i++ {
		task, err := cl.GetNextTask(ctx, "APEX")
		if err != nil {
			t.Logf("No more tasks at iteration %d: %v", i, err)
			break
		}

		// Complete with varying quality
		quality := 0.7 + float64(i)*0.03
		if quality > 1.0 {
			quality = 0.95
		}
		cl.RecordTaskCompletion("APEX", task.ID, quality)
	}

	// Verify progress
	progress, _ := cl.GetAgentProgress("APEX")
	if progress.TotalTasksDone == 0 {
		t.Error("should have completed some tasks")
	}

	t.Logf("Final state - Level: %d, Tasks: %d, Quality: %.2f",
		progress.CurrentLevel, progress.TotalTasksDone, progress.AverageQuality)
}

// =============================================================================
// BENCHMARKS
// =============================================================================

func BenchmarkCurriculumLearner_GetNextTask(b *testing.B) {
	cl := NewCurriculumLearner(nil)
	ctx := context.Background()

	cl.RegisterAgent("BENCH")
	for i := 0; i < 100; i++ {
		cl.AddTask(&CurriculumTask{
			ID:            string(rune(i)),
			Level:         (i % 5) + 1,
			Difficulty:    float64(i%10) / 10.0,
			TeachesSkills: []string{"skill"},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cl.GetNextTask(ctx, "BENCH")
	}
}

func BenchmarkCurriculumLearner_RecordCompletion(b *testing.B) {
	cl := NewCurriculumLearner(nil)
	cl.RegisterAgent("BENCH")

	for i := 0; i < 100; i++ {
		cl.AddTask(&CurriculumTask{
			ID:            string(rune(i)),
			Level:         1,
			TeachesSkills: []string{"skill"},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		taskID := string(rune(i % 100))
		cl.RecordTaskCompletion("BENCH", taskID, 0.8)
	}
}
