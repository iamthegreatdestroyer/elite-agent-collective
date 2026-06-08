package agents

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// CrewDefinition defines a named, ordered sequence of agents for a specific workflow.
type CrewDefinition struct {
	Description string   `yaml:"description"`
	Agents      []string `yaml:"agents"`
}

// crewsFile is the top-level structure for crews.yaml.
type crewsFile struct {
	Crews map[string]*CrewDefinition `yaml:"crews"`
}

// CrewRegistry stores crew definitions indexed by lowercase name.
type CrewRegistry struct {
	crews map[string]*CrewDefinition
}

// NewCrewRegistry loads crew definitions from path.
// Returns an empty registry (not an error) if the file doesn't exist.
func NewCrewRegistry(path string) (*CrewRegistry, error) {
	r := &CrewRegistry{crews: make(map[string]*CrewDefinition)}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return r, nil
	}
	if err != nil {
		return nil, err
	}

	var cf crewsFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return nil, err
	}
	for name, def := range cf.Crews {
		r.crews[strings.ToLower(name)] = def
	}
	return r, nil
}

// Get returns a crew by name (case-insensitive). Returns nil, false if not found.
func (r *CrewRegistry) Get(name string) (*CrewDefinition, bool) {
	def, ok := r.crews[strings.ToLower(name)]
	return def, ok
}

// List returns all crew names.
func (r *CrewRegistry) List() []string {
	names := make([]string, 0, len(r.crews))
	for name := range r.crews {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered crews.
func (r *CrewRegistry) Count() int {
	return len(r.crews)
}
