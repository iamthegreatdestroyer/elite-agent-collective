// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file defines error types for the memory package.

package memory

import "errors"

var (
	// ErrInvalidExperience is returned when an experience is invalid or missing required fields.
	ErrInvalidExperience = errors.New("invalid experience: missing required fields")

	// ErrExperienceNotFound is returned when an experience is not found in the store.
	ErrExperienceNotFound = errors.New("experience not found")

	// ErrInvalidQuery is returned when a query context is invalid.
	ErrInvalidQuery = errors.New("invalid query context")

	// ErrInvalidEmbedding is returned when an embedding has incorrect dimensions.
	ErrInvalidEmbedding = errors.New("invalid embedding dimensions")

	// ErrAgentNotFound is returned when an agent is not found.
	ErrAgentNotFound = errors.New("agent not found")

	// ErrMemoryFull is returned when the memory system has reached capacity.
	ErrMemoryFull = errors.New("memory system at capacity")

	// ErrEvolutionInProgress is returned when an evolution cycle is already running.
	ErrEvolutionInProgress = errors.New("evolution cycle already in progress")

	// ErrPersistenceFailed is returned when memory persistence fails.
	ErrPersistenceFailed = errors.New("failed to persist memory")

	// ErrLoadFailed is returned when memory loading fails.
	ErrLoadFailed = errors.New("failed to load memory")
)
