// GetStats returns a pointer to a copy of the current statistics.
// Returns a pointer to avoid copying the embedded sync.RWMutex.
func (s *MemoryStats) GetStats() *MemoryStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create copies of the maps
	agentCopy := make(map[string]int64)
	for k, v := range s.ExperiencesByAgent {
		agentCopy[k] = v
	}
	tierCopy := make(map[int]int64)
	for k, v := range s.ExperiencesByTier {
		tierCopy[k] = v
	}

	return &MemoryStats{
		TotalExperiences:      s.TotalExperiences,
		ExperiencesByAgent:    agentCopy,
		ExperiencesByTier:     tierCopy,
		TotalRetrievals:       s.TotalRetrievals,
		AvgRetrievalLatencyNs: s.AvgRetrievalLatencyNs,
		CacheHitRate:          s.CacheHitRate,
		BreakthroughCount:     s.BreakthroughCount,
		EvolutionGeneration:   s.EvolutionGeneration,
		LastEvolutionTime:     s.LastEvolutionTime,
	}
}