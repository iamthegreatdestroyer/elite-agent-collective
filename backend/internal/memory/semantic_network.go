// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements Semantic Memory from @NEURAL's Cognitive Architecture Analysis.
//
// Semantic Network (Knowledge Representation):
// - Semantic Nodes: Concepts, entities, and their attributes
// - Semantic Relations: IS-A, HAS-A, PART-OF, and custom relationships
// - Spreading Activation: Associative retrieval through link traversal
// - Property Inheritance: Derive properties through IS-A hierarchies
// - Emergent Concepts: Learn new concepts from experience patterns
//
// This forms the declarative knowledge component of the cognitive architecture,
// complementing episodic memory (experiences) and procedural memory (productions).

package memory

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Error Definitions
// ============================================================================

var (
	// ErrNodeNotFound indicates the semantic node doesn't exist
	ErrNodeNotFound = errors.New("semantic node not found")
	// ErrNodeAlreadyExists indicates a node with this ID already exists
	ErrNodeAlreadyExists = errors.New("semantic node already exists")
	// ErrRelationNotFound indicates the relation doesn't exist
	ErrRelationNotFound = errors.New("semantic relation not found")
	// ErrRelationAlreadyExists indicates this relation already exists
	ErrRelationAlreadyExists = errors.New("semantic relation already exists")
	// ErrInvalidRelationType indicates an unknown relation type
	ErrInvalidRelationType = errors.New("invalid relation type")
	// ErrCyclicHierarchy indicates creating this relation would cause a cycle
	ErrCyclicHierarchy = errors.New("cyclic hierarchy detected")
	// ErrSelfRelation indicates a node cannot relate to itself
	ErrSelfRelation = errors.New("node cannot relate to itself")
)

// ============================================================================
// Semantic Node Types
// ============================================================================

// NodeType represents the type of semantic node.
type NodeType int

const (
	// ConceptNode represents an abstract concept (e.g., "Algorithm", "Security")
	ConceptNode NodeType = iota
	// InstanceNode represents a specific instance (e.g., "QuickSort", "TLS 1.3")
	InstanceNode
	// AttributeNode represents a property or characteristic
	AttributeNode
	// ActionNode represents an action or operation
	ActionNode
	// AgentNode represents an agent in the collective
	AgentNode
	// DomainNode represents a knowledge domain or tier
	DomainNode
)

// String returns the string representation of a NodeType.
func (t NodeType) String() string {
	switch t {
	case ConceptNode:
		return "concept"
	case InstanceNode:
		return "instance"
	case AttributeNode:
		return "attribute"
	case ActionNode:
		return "action"
	case AgentNode:
		return "agent"
	case DomainNode:
		return "domain"
	default:
		return "unknown"
	}
}

// ============================================================================
// Semantic Relation Types
// ============================================================================

// RelationType represents the type of semantic relationship.
type RelationType int

const (
	// IsA represents taxonomic hierarchy (Dog IS-A Animal)
	IsA RelationType = iota
	// HasA represents possession/attribute (Car HAS-A Engine)
	HasA
	// PartOf represents mereological relation (Wheel PART-OF Car)
	PartOf
	// CanDo represents capability (Agent CAN-DO Task)
	CanDo
	// UsedFor represents purpose (Hammer USED-FOR Nailing)
	UsedFor
	// RelatedTo represents generic association
	RelatedTo
	// Requires represents dependency (Security REQUIRES Encryption)
	Requires
	// Produces represents output (Algorithm PRODUCES Result)
	Produces
	// SimilarTo represents similarity relation
	SimilarTo
	// OppositeOf represents antonym/opposite relation
	OppositeOf
	// InstanceOf represents instance-class relation
	InstanceOf
	// BelongsTo represents membership (Agent BELONGS-TO Tier)
	BelongsTo
)

// String returns the string representation of a RelationType.
func (r RelationType) String() string {
	switch r {
	case IsA:
		return "is-a"
	case HasA:
		return "has-a"
	case PartOf:
		return "part-of"
	case CanDo:
		return "can-do"
	case UsedFor:
		return "used-for"
	case RelatedTo:
		return "related-to"
	case Requires:
		return "requires"
	case Produces:
		return "produces"
	case SimilarTo:
		return "similar-to"
	case OppositeOf:
		return "opposite-of"
	case InstanceOf:
		return "instance-of"
	case BelongsTo:
		return "belongs-to"
	default:
		return "unknown"
	}
}

// IsHierarchical returns true if this relation type creates a hierarchy.
func (r RelationType) IsHierarchical() bool {
	return r == IsA || r == PartOf || r == InstanceOf || r == BelongsTo
}

// IsInheritable returns true if properties propagate through this relation.
func (r RelationType) IsInheritable() bool {
	return r == IsA || r == InstanceOf
}

// ============================================================================
// Semantic Node
// ============================================================================

// SemanticNode represents a node in the semantic network.
type SemanticNode struct {
	// ID is the unique identifier for this node
	ID string
	// Label is the human-readable name
	Label string
	// Type indicates the node category
	Type NodeType
	// Activation is the current activation level (0.0 to 1.0)
	Activation float64
	// BaseActivation is the resting activation level
	BaseActivation float64
	// Properties are key-value attributes of this node
	Properties map[string]interface{}
	// Embedding is the vector representation for similarity
	Embedding []float32
	// CreatedAt is when this node was created
	CreatedAt time.Time
	// LastAccessed is when this node was last retrieved
	LastAccessed time.Time
	// AccessCount tracks how often this node is accessed
	AccessCount int64
	// Confidence represents certainty about this concept (0.0 to 1.0)
	Confidence float64
	// Source indicates where this knowledge came from
	Source string
}

// NewSemanticNode creates a new semantic node.
func NewSemanticNode(id, label string, nodeType NodeType) *SemanticNode {
	now := time.Now()
	return &SemanticNode{
		ID:             id,
		Label:          label,
		Type:           nodeType,
		Activation:     0.0,
		BaseActivation: 0.3,
		Properties:     make(map[string]interface{}),
		CreatedAt:      now,
		LastAccessed:   now,
		AccessCount:    0,
		Confidence:     1.0,
		Source:         "manual",
	}
}

// SetProperty sets a property on the node.
func (n *SemanticNode) SetProperty(key string, value interface{}) {
	n.Properties[key] = value
}

// GetProperty gets a property from the node.
func (n *SemanticNode) GetProperty(key string) (interface{}, bool) {
	val, ok := n.Properties[key]
	return val, ok
}

// Clone creates a deep copy of the node.
func (n *SemanticNode) Clone() *SemanticNode {
	clone := &SemanticNode{
		ID:             n.ID,
		Label:          n.Label,
		Type:           n.Type,
		Activation:     n.Activation,
		BaseActivation: n.BaseActivation,
		Properties:     make(map[string]interface{}),
		CreatedAt:      n.CreatedAt,
		LastAccessed:   n.LastAccessed,
		AccessCount:    n.AccessCount,
		Confidence:     n.Confidence,
		Source:         n.Source,
	}
	for k, v := range n.Properties {
		clone.Properties[k] = v
	}
	if n.Embedding != nil {
		clone.Embedding = make([]float32, len(n.Embedding))
		copy(clone.Embedding, n.Embedding)
	}
	return clone
}

// ============================================================================
// Semantic Relation
// ============================================================================

// SemanticRelation represents an edge in the semantic network.
type SemanticRelation struct {
	// ID is the unique identifier for this relation
	ID string
	// SourceID is the ID of the source node
	SourceID string
	// TargetID is the ID of the target node
	TargetID string
	// Type is the type of relationship
	Type RelationType
	// Weight is the strength of this relation (0.0 to 1.0)
	Weight float64
	// Properties are key-value attributes of this relation
	Properties map[string]interface{}
	// CreatedAt is when this relation was created
	CreatedAt time.Time
	// Confidence represents certainty about this relation
	Confidence float64
	// Source indicates where this knowledge came from
	Source string
}

// NewSemanticRelation creates a new semantic relation.
func NewSemanticRelation(sourceID, targetID string, relType RelationType) *SemanticRelation {
	return &SemanticRelation{
		ID:         fmt.Sprintf("%s-%s-%s", sourceID, relType.String(), targetID),
		SourceID:   sourceID,
		TargetID:   targetID,
		Type:       relType,
		Weight:     1.0,
		Properties: make(map[string]interface{}),
		CreatedAt:  time.Now(),
		Confidence: 1.0,
		Source:     "manual",
	}
}

// ============================================================================
// Semantic Network Configuration
// ============================================================================

// SemanticNetworkConfig holds configuration for the semantic network.
type SemanticNetworkConfig struct {
	// MaxNodes is the maximum number of nodes allowed
	MaxNodes int
	// MaxRelationsPerNode is the maximum relations per node
	MaxRelationsPerNode int
	// ActivationDecayRate controls how fast activation fades
	ActivationDecayRate float64
	// ActivationThreshold below which nodes are considered inactive
	ActivationThreshold float64
	// SpreadingFactor controls activation propagation strength
	SpreadingFactor float64
	// MaxSpreadingDepth limits how far activation spreads
	MaxSpreadingDepth int
	// InheritanceDepth limits property inheritance chain
	InheritanceDepth int
	// MinConfidenceThreshold below which facts are considered unreliable
	MinConfidenceThreshold float64
}

// DefaultSemanticNetworkConfig returns sensible defaults.
func DefaultSemanticNetworkConfig() SemanticNetworkConfig {
	return SemanticNetworkConfig{
		MaxNodes:               100000,
		MaxRelationsPerNode:    100,
		ActivationDecayRate:    0.1,
		ActivationThreshold:    0.1,
		SpreadingFactor:        0.5,
		MaxSpreadingDepth:      3,
		InheritanceDepth:       5,
		MinConfidenceThreshold: 0.3,
	}
}

// ============================================================================
// Semantic Network
// ============================================================================

// SemanticNetwork manages semantic memory with spreading activation.
type SemanticNetwork struct {
	mu sync.RWMutex

	// nodes maps node IDs to nodes
	nodes map[string]*SemanticNode
	// relations maps relation IDs to relations
	relations map[string]*SemanticRelation

	// outgoing maps source node ID to its outgoing relations
	outgoing map[string][]*SemanticRelation
	// incoming maps target node ID to its incoming relations
	incoming map[string][]*SemanticRelation

	// config holds network configuration
	config SemanticNetworkConfig

	// stats tracks network statistics
	stats *SemanticNetworkStats
}

// SemanticNetworkStats tracks network performance.
type SemanticNetworkStats struct {
	NodesCreated       int64
	RelationsCreated   int64
	ActivationQueries  int64
	InheritanceQueries int64
	SpreadingCycles    int64
	ConceptsLearned    int64
	LastUpdated        time.Time
}

// NewSemanticNetwork creates a new semantic network.
func NewSemanticNetwork(config SemanticNetworkConfig) *SemanticNetwork {
	return &SemanticNetwork{
		nodes:     make(map[string]*SemanticNode),
		relations: make(map[string]*SemanticRelation),
		outgoing:  make(map[string][]*SemanticRelation),
		incoming:  make(map[string][]*SemanticRelation),
		config:    config,
		stats: &SemanticNetworkStats{
			LastUpdated: time.Now(),
		},
	}
}

// ============================================================================
// Node Management
// ============================================================================

// AddNode adds a new node to the network.
func (sn *SemanticNetwork) AddNode(node *SemanticNode) error {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	if _, exists := sn.nodes[node.ID]; exists {
		return ErrNodeAlreadyExists
	}

	if len(sn.nodes) >= sn.config.MaxNodes {
		// Evict least recently accessed node
		sn.evictLRUNode()
	}

	sn.nodes[node.ID] = node
	sn.outgoing[node.ID] = make([]*SemanticRelation, 0)
	sn.incoming[node.ID] = make([]*SemanticRelation, 0)
	sn.stats.NodesCreated++
	sn.stats.LastUpdated = time.Now()

	return nil
}

// GetNode retrieves a node by ID.
func (sn *SemanticNetwork) GetNode(id string) (*SemanticNode, error) {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	node, exists := sn.nodes[id]
	if !exists {
		return nil, ErrNodeNotFound
	}

	node.LastAccessed = time.Now()
	node.AccessCount++

	return node, nil
}

// RemoveNode removes a node and all its relations.
func (sn *SemanticNetwork) RemoveNode(id string) error {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	if _, exists := sn.nodes[id]; !exists {
		return ErrNodeNotFound
	}

	// Remove all relations involving this node
	for _, rel := range sn.outgoing[id] {
		delete(sn.relations, rel.ID)
		sn.removeFromIncoming(rel.TargetID, rel.ID)
	}
	for _, rel := range sn.incoming[id] {
		delete(sn.relations, rel.ID)
		sn.removeFromOutgoing(rel.SourceID, rel.ID)
	}

	delete(sn.nodes, id)
	delete(sn.outgoing, id)
	delete(sn.incoming, id)

	return nil
}

// UpdateNode updates an existing node.
func (sn *SemanticNetwork) UpdateNode(node *SemanticNode) error {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	if _, exists := sn.nodes[node.ID]; !exists {
		return ErrNodeNotFound
	}

	sn.nodes[node.ID] = node
	sn.stats.LastUpdated = time.Now()

	return nil
}

// GetAllNodes returns all nodes in the network.
func (sn *SemanticNetwork) GetAllNodes() []*SemanticNode {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	nodes := make([]*SemanticNode, 0, len(sn.nodes))
	for _, node := range sn.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodesByType returns all nodes of a specific type.
func (sn *SemanticNetwork) GetNodesByType(nodeType NodeType) []*SemanticNode {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	nodes := make([]*SemanticNode, 0)
	for _, node := range sn.nodes {
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// FindNodesByLabel finds nodes whose labels contain the query string.
func (sn *SemanticNetwork) FindNodesByLabel(query string) []*SemanticNode {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	nodes := make([]*SemanticNode, 0)
	for _, node := range sn.nodes {
		if containsIgnoreCase(node.Label, query) {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// evictLRUNode removes the least recently used node.
func (sn *SemanticNetwork) evictLRUNode() {
	var oldest *SemanticNode
	for _, node := range sn.nodes {
		if oldest == nil || node.LastAccessed.Before(oldest.LastAccessed) {
			oldest = node
		}
	}
	if oldest != nil {
		// Remove relations first
		for _, rel := range sn.outgoing[oldest.ID] {
			delete(sn.relations, rel.ID)
		}
		for _, rel := range sn.incoming[oldest.ID] {
			delete(sn.relations, rel.ID)
		}
		delete(sn.nodes, oldest.ID)
		delete(sn.outgoing, oldest.ID)
		delete(sn.incoming, oldest.ID)
	}
}

// ============================================================================
// Relation Management
// ============================================================================

// AddRelation adds a new relation between nodes.
func (sn *SemanticNetwork) AddRelation(rel *SemanticRelation) error {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	// Validate source and target exist
	if _, exists := sn.nodes[rel.SourceID]; !exists {
		return fmt.Errorf("%w: source %s", ErrNodeNotFound, rel.SourceID)
	}
	if _, exists := sn.nodes[rel.TargetID]; !exists {
		return fmt.Errorf("%w: target %s", ErrNodeNotFound, rel.TargetID)
	}

	// Check for self-relation
	if rel.SourceID == rel.TargetID {
		return ErrSelfRelation
	}

	// Check for duplicate
	if _, exists := sn.relations[rel.ID]; exists {
		return ErrRelationAlreadyExists
	}

	// Check for cycles in hierarchical relations
	if rel.Type.IsHierarchical() {
		if sn.wouldCreateCycle(rel.SourceID, rel.TargetID, rel.Type) {
			return ErrCyclicHierarchy
		}
	}

	// Check max relations per node
	if len(sn.outgoing[rel.SourceID]) >= sn.config.MaxRelationsPerNode {
		return fmt.Errorf("max relations exceeded for node %s", rel.SourceID)
	}

	sn.relations[rel.ID] = rel
	sn.outgoing[rel.SourceID] = append(sn.outgoing[rel.SourceID], rel)
	sn.incoming[rel.TargetID] = append(sn.incoming[rel.TargetID], rel)
	sn.stats.RelationsCreated++
	sn.stats.LastUpdated = time.Now()

	return nil
}

// GetRelation retrieves a relation by ID.
func (sn *SemanticNetwork) GetRelation(id string) (*SemanticRelation, error) {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	rel, exists := sn.relations[id]
	if !exists {
		return nil, ErrRelationNotFound
	}
	return rel, nil
}

// RemoveRelation removes a relation.
func (sn *SemanticNetwork) RemoveRelation(id string) error {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	rel, exists := sn.relations[id]
	if !exists {
		return ErrRelationNotFound
	}

	sn.removeFromOutgoing(rel.SourceID, id)
	sn.removeFromIncoming(rel.TargetID, id)
	delete(sn.relations, id)

	return nil
}

// GetOutgoingRelations returns all relations from a node.
func (sn *SemanticNetwork) GetOutgoingRelations(nodeID string) []*SemanticRelation {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	return sn.outgoing[nodeID]
}

// GetIncomingRelations returns all relations to a node.
func (sn *SemanticNetwork) GetIncomingRelations(nodeID string) []*SemanticRelation {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	return sn.incoming[nodeID]
}

// GetRelatedNodes returns nodes related to a source node by a specific relation type.
func (sn *SemanticNetwork) GetRelatedNodes(nodeID string, relType RelationType) []*SemanticNode {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	related := make([]*SemanticNode, 0)
	for _, rel := range sn.outgoing[nodeID] {
		if rel.Type == relType {
			if node, exists := sn.nodes[rel.TargetID]; exists {
				related = append(related, node)
			}
		}
	}
	return related
}

// GetReverseRelatedNodes returns nodes that relate TO this node by a specific type.
func (sn *SemanticNetwork) GetReverseRelatedNodes(nodeID string, relType RelationType) []*SemanticNode {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	related := make([]*SemanticNode, 0)
	for _, rel := range sn.incoming[nodeID] {
		if rel.Type == relType {
			if node, exists := sn.nodes[rel.SourceID]; exists {
				related = append(related, node)
			}
		}
	}
	return related
}

// removeFromOutgoing removes a relation from the outgoing list.
func (sn *SemanticNetwork) removeFromOutgoing(nodeID, relID string) {
	rels := sn.outgoing[nodeID]
	for i, rel := range rels {
		if rel.ID == relID {
			sn.outgoing[nodeID] = append(rels[:i], rels[i+1:]...)
			return
		}
	}
}

// removeFromIncoming removes a relation from the incoming list.
func (sn *SemanticNetwork) removeFromIncoming(nodeID, relID string) {
	rels := sn.incoming[nodeID]
	for i, rel := range rels {
		if rel.ID == relID {
			sn.incoming[nodeID] = append(rels[:i], rels[i+1:]...)
			return
		}
	}
}

// wouldCreateCycle checks if adding this relation would create a cycle.
func (sn *SemanticNetwork) wouldCreateCycle(sourceID, targetID string, relType RelationType) bool {
	// Check if there's already a path from target to source
	visited := make(map[string]bool)
	return sn.hasPath(targetID, sourceID, relType, visited)
}

// hasPath checks if there's a path from source to target.
func (sn *SemanticNetwork) hasPath(from, to string, relType RelationType, visited map[string]bool) bool {
	if from == to {
		return true
	}
	if visited[from] {
		return false
	}
	visited[from] = true

	for _, rel := range sn.outgoing[from] {
		if rel.Type == relType {
			if sn.hasPath(rel.TargetID, to, relType, visited) {
				return true
			}
		}
	}
	return false
}

// ============================================================================
// Spreading Activation
// ============================================================================

// ActivationResult holds the result of spreading activation.
type ActivationResult struct {
	// ActivatedNodes maps node IDs to their activation levels
	ActivatedNodes map[string]float64
	// SpreadPath tracks how activation spread
	SpreadPath []string
	// Iterations is how many spreading cycles occurred
	Iterations int
}

// SpreadActivation performs spreading activation from source nodes.
func (sn *SemanticNetwork) SpreadActivation(sourceIDs []string, initialActivation float64) *ActivationResult {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	sn.stats.SpreadingCycles++

	result := &ActivationResult{
		ActivatedNodes: make(map[string]float64),
		SpreadPath:     make([]string, 0),
	}

	// Initialize source nodes
	for _, id := range sourceIDs {
		if node, exists := sn.nodes[id]; exists {
			node.Activation = initialActivation
			result.ActivatedNodes[id] = initialActivation
			result.SpreadPath = append(result.SpreadPath, id)
		}
	}

	// Spreading activation loop
	for depth := 0; depth < sn.config.MaxSpreadingDepth; depth++ {
		result.Iterations++
		newActivations := make(map[string]float64)

		for nodeID, activation := range result.ActivatedNodes {
			// Spread to connected nodes
			for _, rel := range sn.outgoing[nodeID] {
				spreadAmount := activation * sn.config.SpreadingFactor * rel.Weight
				if spreadAmount > sn.config.ActivationThreshold {
					targetNode := sn.nodes[rel.TargetID]
					if targetNode != nil {
						newAct := targetNode.Activation + spreadAmount
						if newAct > 1.0 {
							newAct = 1.0
						}
						newActivations[rel.TargetID] = newAct
					}
				}
			}
		}

		// Apply new activations
		for nodeID, newAct := range newActivations {
			if node, exists := sn.nodes[nodeID]; exists {
				node.Activation = newAct
				if _, already := result.ActivatedNodes[nodeID]; !already {
					result.SpreadPath = append(result.SpreadPath, nodeID)
				}
				result.ActivatedNodes[nodeID] = newAct
			}
		}

		// Stop if no new activations
		if len(newActivations) == 0 {
			break
		}
	}

	return result
}

// DecayActivation reduces activation levels over time.
func (sn *SemanticNetwork) DecayActivation(elapsed time.Duration) {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	decayFactor := math.Exp(-sn.config.ActivationDecayRate * elapsed.Seconds())

	for _, node := range sn.nodes {
		// Decay towards base activation
		node.Activation = node.BaseActivation + (node.Activation-node.BaseActivation)*decayFactor
	}
}

// ResetActivation resets all nodes to base activation.
func (sn *SemanticNetwork) ResetActivation() {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	for _, node := range sn.nodes {
		node.Activation = node.BaseActivation
	}
}

// GetMostActivated returns the N most activated nodes.
func (sn *SemanticNetwork) GetMostActivated(n int) []*SemanticNode {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	nodes := make([]*SemanticNode, 0, len(sn.nodes))
	for _, node := range sn.nodes {
		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Activation > nodes[j].Activation
	})

	if n > len(nodes) {
		n = len(nodes)
	}
	return nodes[:n]
}

// ============================================================================
// Property Inheritance
// ============================================================================

// InheritedProperty represents a property inherited through the hierarchy.
type InheritedProperty struct {
	Key          string
	Value        interface{}
	SourceNodeID string
	Distance     int
	Confidence   float64
}

// GetInheritedProperties returns all properties including inherited ones.
func (sn *SemanticNetwork) GetInheritedProperties(nodeID string) (map[string]*InheritedProperty, error) {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	sn.stats.InheritanceQueries++

	node, exists := sn.nodes[nodeID]
	if !exists {
		return nil, ErrNodeNotFound
	}

	properties := make(map[string]*InheritedProperty)

	// Start with local properties
	for k, v := range node.Properties {
		properties[k] = &InheritedProperty{
			Key:          k,
			Value:        v,
			SourceNodeID: nodeID,
			Distance:     0,
			Confidence:   node.Confidence,
		}
	}

	// Traverse inheritance hierarchy
	visited := make(map[string]bool)
	sn.collectInheritedProperties(nodeID, properties, visited, 0)

	return properties, nil
}

// collectInheritedProperties recursively collects inherited properties.
func (sn *SemanticNetwork) collectInheritedProperties(
	nodeID string,
	properties map[string]*InheritedProperty,
	visited map[string]bool,
	depth int,
) {
	if depth >= sn.config.InheritanceDepth {
		return
	}
	if visited[nodeID] {
		return
	}
	visited[nodeID] = true

	// Follow inheritable relations (IS-A, INSTANCE-OF)
	for _, rel := range sn.outgoing[nodeID] {
		if rel.Type.IsInheritable() {
			parent := sn.nodes[rel.TargetID]
			if parent == nil {
				continue
			}

			// Inherit properties from parent
			for k, v := range parent.Properties {
				// Only inherit if not already defined closer
				if existing, ok := properties[k]; !ok || existing.Distance > depth+1 {
					properties[k] = &InheritedProperty{
						Key:          k,
						Value:        v,
						SourceNodeID: parent.ID,
						Distance:     depth + 1,
						Confidence:   parent.Confidence * rel.Confidence,
					}
				}
			}

			// Recurse to parent's parents
			sn.collectInheritedProperties(parent.ID, properties, visited, depth+1)
		}
	}
}

// ============================================================================
// Semantic Queries
// ============================================================================

// QueryResult holds the result of a semantic query.
type QueryResult struct {
	Nodes      []*SemanticNode
	Relations  []*SemanticRelation
	Confidence float64
}

// IsA checks if nodeA is-a nodeB (directly or transitively).
func (sn *SemanticNetwork) IsA(nodeA, nodeB string) bool {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	visited := make(map[string]bool)
	return sn.hasPath(nodeA, nodeB, IsA, visited)
}

// HasProperty checks if a node has a property (directly or inherited).
func (sn *SemanticNetwork) HasProperty(nodeID, propertyKey string) (interface{}, bool) {
	props, err := sn.GetInheritedProperties(nodeID)
	if err != nil {
		return nil, false
	}

	if prop, ok := props[propertyKey]; ok {
		return prop.Value, true
	}
	return nil, false
}

// FindCommonAncestors finds common ancestors of two nodes.
func (sn *SemanticNetwork) FindCommonAncestors(nodeA, nodeB string) []*SemanticNode {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	ancestorsA := sn.getAncestors(nodeA)
	ancestorsB := sn.getAncestors(nodeB)

	common := make([]*SemanticNode, 0)
	for id := range ancestorsA {
		if _, ok := ancestorsB[id]; ok {
			if node, exists := sn.nodes[id]; exists {
				common = append(common, node)
			}
		}
	}
	return common
}

// getAncestors returns all ancestors of a node through IS-A relations.
func (sn *SemanticNetwork) getAncestors(nodeID string) map[string]int {
	ancestors := make(map[string]int)
	visited := make(map[string]bool)
	sn.collectAncestors(nodeID, ancestors, visited, 0)
	return ancestors
}

// collectAncestors recursively collects ancestors.
func (sn *SemanticNetwork) collectAncestors(nodeID string, ancestors map[string]int, visited map[string]bool, depth int) {
	if visited[nodeID] {
		return
	}
	visited[nodeID] = true

	for _, rel := range sn.outgoing[nodeID] {
		if rel.Type == IsA || rel.Type == InstanceOf {
			if _, ok := ancestors[rel.TargetID]; !ok {
				ancestors[rel.TargetID] = depth + 1
			}
			sn.collectAncestors(rel.TargetID, ancestors, visited, depth+1)
		}
	}
}

// FindShortestPath finds the shortest path between two nodes.
func (sn *SemanticNetwork) FindShortestPath(fromID, toID string) ([]*SemanticNode, error) {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	if _, exists := sn.nodes[fromID]; !exists {
		return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, fromID)
	}
	if _, exists := sn.nodes[toID]; !exists {
		return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, toID)
	}

	// BFS for shortest path
	type pathNode struct {
		id   string
		path []string
	}

	queue := []pathNode{{id: fromID, path: []string{fromID}}}
	visited := make(map[string]bool)
	visited[fromID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.id == toID {
			// Convert IDs to nodes
			result := make([]*SemanticNode, len(current.path))
			for i, id := range current.path {
				result[i] = sn.nodes[id]
			}
			return result, nil
		}

		// Explore neighbors (both directions)
		for _, rel := range sn.outgoing[current.id] {
			if !visited[rel.TargetID] {
				visited[rel.TargetID] = true
				newPath := make([]string, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = rel.TargetID
				queue = append(queue, pathNode{id: rel.TargetID, path: newPath})
			}
		}
		for _, rel := range sn.incoming[current.id] {
			if !visited[rel.SourceID] {
				visited[rel.SourceID] = true
				newPath := make([]string, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = rel.SourceID
				queue = append(queue, pathNode{id: rel.SourceID, path: newPath})
			}
		}
	}

	return nil, fmt.Errorf("no path found between %s and %s", fromID, toID)
}

// ============================================================================
// Concept Similarity
// ============================================================================

// SimilarityResult holds similarity computation result.
type SimilarityResult struct {
	NodeA      string
	NodeB      string
	Similarity float64
	Method     string
}

// ComputeSimilarity computes semantic similarity between two nodes.
func (sn *SemanticNetwork) ComputeSimilarity(nodeA, nodeB string) (*SimilarityResult, error) {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	a, existsA := sn.nodes[nodeA]
	b, existsB := sn.nodes[nodeB]

	if !existsA {
		return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, nodeA)
	}
	if !existsB {
		return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, nodeB)
	}

	result := &SimilarityResult{
		NodeA: nodeA,
		NodeB: nodeB,
	}

	// Use embedding similarity if available
	if len(a.Embedding) > 0 && len(b.Embedding) > 0 {
		result.Similarity = cosineSimilarityFloat32(a.Embedding, b.Embedding)
		result.Method = "embedding"
		return result, nil
	}

	// Fall back to structure-based similarity
	// Wu-Palmer similarity based on common ancestors
	ancestorsA := sn.getAncestors(nodeA)
	ancestorsB := sn.getAncestors(nodeB)

	// Find lowest common ancestor (LCA)
	var lcaDepth int
	for id, depthA := range ancestorsA {
		if depthB, ok := ancestorsB[id]; ok {
			combinedDepth := depthA + depthB
			if combinedDepth > lcaDepth {
				lcaDepth = combinedDepth
			}
		}
	}

	depthA := sn.getNodeDepth(nodeA)
	depthB := sn.getNodeDepth(nodeB)

	if depthA+depthB > 0 {
		result.Similarity = float64(2*lcaDepth) / float64(depthA+depthB+2*lcaDepth)
	}
	result.Method = "wu-palmer"

	return result, nil
}

// getNodeDepth returns the depth of a node in the IS-A hierarchy.
func (sn *SemanticNetwork) getNodeDepth(nodeID string) int {
	maxDepth := 0
	for _, rel := range sn.outgoing[nodeID] {
		if rel.Type == IsA || rel.Type == InstanceOf {
			depth := 1 + sn.getNodeDepth(rel.TargetID)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}
	return maxDepth
}

// cosineSimilarityFloat32 computes cosine similarity between two vectors.
func cosineSimilarityFloat32(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ============================================================================
// Statistics and Introspection
// ============================================================================

// GetStats returns network statistics.
func (sn *SemanticNetwork) GetStats() *SemanticNetworkStats {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	// Return a copy to avoid data races
	return &SemanticNetworkStats{
		NodesCreated:       sn.stats.NodesCreated,
		RelationsCreated:   sn.stats.RelationsCreated,
		ActivationQueries:  sn.stats.ActivationQueries,
		InheritanceQueries: sn.stats.InheritanceQueries,
		SpreadingCycles:    sn.stats.SpreadingCycles,
		ConceptsLearned:    sn.stats.ConceptsLearned,
		LastUpdated:        sn.stats.LastUpdated,
	}
}

// NodeCount returns the number of nodes.
func (sn *SemanticNetwork) NodeCount() int {
	sn.mu.RLock()
	defer sn.mu.RUnlock()
	return len(sn.nodes)
}

// RelationCount returns the number of relations.
func (sn *SemanticNetwork) RelationCount() int {
	sn.mu.RLock()
	defer sn.mu.RUnlock()
	return len(sn.relations)
}

// ============================================================================
// Snapshot and Restore
// ============================================================================

// SemanticNetworkSnapshot holds a serializable snapshot of the network.
type SemanticNetworkSnapshot struct {
	Nodes     []*SemanticNode
	Relations []*SemanticRelation
	Stats     *SemanticNetworkStats
	Timestamp time.Time
}

// Snapshot creates a snapshot of the current network state.
func (sn *SemanticNetwork) Snapshot() *SemanticNetworkSnapshot {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	snapshot := &SemanticNetworkSnapshot{
		Nodes:     make([]*SemanticNode, 0, len(sn.nodes)),
		Relations: make([]*SemanticRelation, 0, len(sn.relations)),
		Stats:     sn.GetStats(),
		Timestamp: time.Now(),
	}

	for _, node := range sn.nodes {
		snapshot.Nodes = append(snapshot.Nodes, node.Clone())
	}

	for _, rel := range sn.relations {
		relCopy := *rel
		relCopy.Properties = make(map[string]interface{})
		for k, v := range rel.Properties {
			relCopy.Properties[k] = v
		}
		snapshot.Relations = append(snapshot.Relations, &relCopy)
	}

	return snapshot
}

// Restore restores the network from a snapshot.
func (sn *SemanticNetwork) Restore(snapshot *SemanticNetworkSnapshot) error {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	// Clear current state
	sn.nodes = make(map[string]*SemanticNode)
	sn.relations = make(map[string]*SemanticRelation)
	sn.outgoing = make(map[string][]*SemanticRelation)
	sn.incoming = make(map[string][]*SemanticRelation)

	// Restore nodes
	for _, node := range snapshot.Nodes {
		sn.nodes[node.ID] = node.Clone()
		sn.outgoing[node.ID] = make([]*SemanticRelation, 0)
		sn.incoming[node.ID] = make([]*SemanticRelation, 0)
	}

	// Restore relations
	for _, rel := range snapshot.Relations {
		relCopy := *rel
		relCopy.Properties = make(map[string]interface{})
		for k, v := range rel.Properties {
			relCopy.Properties[k] = v
		}
		sn.relations[rel.ID] = &relCopy
		sn.outgoing[rel.SourceID] = append(sn.outgoing[rel.SourceID], &relCopy)
		sn.incoming[rel.TargetID] = append(sn.incoming[rel.TargetID], &relCopy)
	}

	// Restore stats
	if snapshot.Stats != nil {
		sn.stats = &SemanticNetworkStats{
			NodesCreated:       snapshot.Stats.NodesCreated,
			RelationsCreated:   snapshot.Stats.RelationsCreated,
			ActivationQueries:  snapshot.Stats.ActivationQueries,
			InheritanceQueries: snapshot.Stats.InheritanceQueries,
			SpreadingCycles:    snapshot.Stats.SpreadingCycles,
			ConceptsLearned:    snapshot.Stats.ConceptsLearned,
			LastUpdated:        time.Now(),
		}
	}

	return nil
}

// Clear removes all nodes and relations.
func (sn *SemanticNetwork) Clear() {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	sn.nodes = make(map[string]*SemanticNode)
	sn.relations = make(map[string]*SemanticRelation)
	sn.outgoing = make(map[string][]*SemanticRelation)
	sn.incoming = make(map[string][]*SemanticRelation)
}

// ============================================================================
// Helper Functions
// ============================================================================

// containsIgnoreCase checks if s contains substr (case insensitive).
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsIgnoreCaseImpl(s, substr)))
}

func containsIgnoreCaseImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			pc := substr[j]
			if sc >= 'A' && sc <= 'Z' {
				sc += 'a' - 'A'
			}
			if pc >= 'A' && pc <= 'Z' {
				pc += 'a' - 'A'
			}
			if sc != pc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// ============================================================================
// Semantic Inference Engine
// ============================================================================

// InferenceType specifies the type of semantic inference.
type InferenceType int

const (
	// InferenceInheritance derives properties through IS-A hierarchy
	InferenceInheritance InferenceType = iota
	// InferenceMembership determines category membership
	InferenceMembership
	// InferenceAnalogy finds analogous relationships
	InferenceAnalogy
	// InferenceCompletion predicts missing relationships
	InferenceCompletion
	// InferenceDefault applies default reasoning
	InferenceDefault
)

// InferenceResult holds the result of semantic inference.
type InferenceResult struct {
	Type       InferenceType
	Query      string
	Answer     interface{}
	Confidence float64
	Reasoning  []string
	SourceIDs  []string
}

// SemanticInferenceEngine performs reasoning over the semantic network.
type SemanticInferenceEngine struct {
	network *SemanticNetwork
}

// NewSemanticInferenceEngine creates a new inference engine.
func NewSemanticInferenceEngine(network *SemanticNetwork) *SemanticInferenceEngine {
	return &SemanticInferenceEngine{network: network}
}

// InferProperty uses inheritance to determine a property value.
func (e *SemanticInferenceEngine) InferProperty(nodeID, propertyKey string) (*InferenceResult, error) {
	props, err := e.network.GetInheritedProperties(nodeID)
	if err != nil {
		return nil, err
	}

	result := &InferenceResult{
		Type:      InferenceInheritance,
		Query:     fmt.Sprintf("What is %s of %s?", propertyKey, nodeID),
		Reasoning: make([]string, 0),
		SourceIDs: make([]string, 0),
	}

	if prop, ok := props[propertyKey]; ok {
		result.Answer = prop.Value
		result.Confidence = prop.Confidence
		result.SourceIDs = append(result.SourceIDs, prop.SourceNodeID)

		if prop.Distance == 0 {
			result.Reasoning = append(result.Reasoning,
				fmt.Sprintf("%s has direct property %s = %v", nodeID, propertyKey, prop.Value))
		} else {
			result.Reasoning = append(result.Reasoning,
				fmt.Sprintf("%s inherits %s = %v from %s (distance: %d)",
					nodeID, propertyKey, prop.Value, prop.SourceNodeID, prop.Distance))
		}
		return result, nil
	}

	return nil, fmt.Errorf("property %s not found for node %s", propertyKey, nodeID)
}

// InferMembership determines if a node belongs to a category.
func (e *SemanticInferenceEngine) InferMembership(instanceID, categoryID string) (*InferenceResult, error) {
	result := &InferenceResult{
		Type:      InferenceMembership,
		Query:     fmt.Sprintf("Is %s a %s?", instanceID, categoryID),
		Reasoning: make([]string, 0),
		SourceIDs: make([]string, 0),
	}

	// Check direct and transitive IS-A relationships
	if e.network.IsA(instanceID, categoryID) {
		result.Answer = true
		result.Confidence = 1.0

		// Build reasoning chain
		path, err := e.network.FindShortestPath(instanceID, categoryID)
		if err == nil {
			for i := 0; i < len(path)-1; i++ {
				result.Reasoning = append(result.Reasoning,
					fmt.Sprintf("%s is-a %s", path[i].Label, path[i+1].Label))
				result.SourceIDs = append(result.SourceIDs, path[i].ID)
			}
			result.SourceIDs = append(result.SourceIDs, path[len(path)-1].ID)
		}
		return result, nil
	}

	result.Answer = false
	result.Confidence = 1.0
	result.Reasoning = append(result.Reasoning,
		fmt.Sprintf("No IS-A path found from %s to %s", instanceID, categoryID))
	return result, nil
}

// InferAnalogy finds analogous relationships between concepts.
// Given A:B, find X such that C:X has the same relationship.
func (e *SemanticInferenceEngine) InferAnalogy(nodeA, nodeB, nodeC string) (*InferenceResult, error) {
	result := &InferenceResult{
		Type:      InferenceAnalogy,
		Query:     fmt.Sprintf("%s is to %s as %s is to ?", nodeA, nodeB, nodeC),
		Reasoning: make([]string, 0),
		SourceIDs: []string{nodeA, nodeB, nodeC},
	}

	// Find the relation from A to B
	e.network.mu.RLock()
	defer e.network.mu.RUnlock()

	var abRelation *SemanticRelation
	for _, rel := range e.network.outgoing[nodeA] {
		if rel.TargetID == nodeB {
			abRelation = rel
			break
		}
	}

	if abRelation == nil {
		return nil, fmt.Errorf("no relation found from %s to %s", nodeA, nodeB)
	}

	result.Reasoning = append(result.Reasoning,
		fmt.Sprintf("Found relation: %s -%s-> %s", nodeA, abRelation.Type, nodeB))

	// Find analogous relation from C
	candidates := make([]*SemanticNode, 0)
	for _, rel := range e.network.outgoing[nodeC] {
		if rel.Type == abRelation.Type {
			if target, exists := e.network.nodes[rel.TargetID]; exists {
				candidates = append(candidates, target)
				result.Reasoning = append(result.Reasoning,
					fmt.Sprintf("Candidate: %s -%s-> %s", nodeC, rel.Type, target.Label))
			}
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no analogous relationship found for %s", nodeC)
	}

	// Return the best candidate (could use similarity to B to rank)
	result.Answer = candidates[0].ID
	result.Confidence = 1.0 / float64(len(candidates)) // Lower confidence if multiple candidates

	return result, nil
}

// InferCompletion predicts missing relationships for a node.
func (e *SemanticInferenceEngine) InferCompletion(nodeID string) (*InferenceResult, error) {
	result := &InferenceResult{
		Type:      InferenceCompletion,
		Query:     fmt.Sprintf("What relationships might %s have?", nodeID),
		Reasoning: make([]string, 0),
		SourceIDs: []string{nodeID},
	}

	e.network.mu.RLock()
	defer e.network.mu.RUnlock()

	node, exists := e.network.nodes[nodeID]
	if !exists {
		return nil, ErrNodeNotFound
	}

	// Find similar nodes based on shared properties/relationships
	similarNodes := e.findSimilarNodes(nodeID)

	// Collect relationships that similar nodes have but this node doesn't
	existingRels := make(map[string]bool)
	for _, rel := range e.network.outgoing[nodeID] {
		existingRels[rel.Type.String()+":"+rel.TargetID] = true
	}

	predictions := make([]map[string]interface{}, 0)
	for _, similar := range similarNodes {
		for _, rel := range e.network.outgoing[similar.ID] {
			key := rel.Type.String() + ":" + rel.TargetID
			if !existingRels[key] {
				predictions = append(predictions, map[string]interface{}{
					"type":       rel.Type.String(),
					"target":     rel.TargetID,
					"confidence": similar.Activation * rel.Weight,
					"source":     similar.ID,
				})
				result.Reasoning = append(result.Reasoning,
					fmt.Sprintf("Similar node %s has %s relation to %s",
						similar.Label, rel.Type, rel.TargetID))
			}
		}
	}

	result.Answer = predictions
	if len(predictions) > 0 {
		result.Confidence = predictions[0]["confidence"].(float64)
	}

	_ = node // Use node variable
	return result, nil
}

// findSimilarNodes finds nodes similar to the given node.
func (e *SemanticInferenceEngine) findSimilarNodes(nodeID string) []*SemanticNode {
	node := e.network.nodes[nodeID]
	if node == nil {
		return nil
	}

	similar := make([]*SemanticNode, 0)

	// Find nodes with same type
	for _, other := range e.network.nodes {
		if other.ID == nodeID {
			continue
		}
		if other.Type == node.Type {
			// Check for shared relationships
			sharedCount := e.countSharedRelationships(nodeID, other.ID)
			if sharedCount > 0 {
				similar = append(similar, other)
			}
		}
	}

	// Sort by similarity
	sort.Slice(similar, func(i, j int) bool {
		return e.countSharedRelationships(nodeID, similar[i].ID) >
			e.countSharedRelationships(nodeID, similar[j].ID)
	})

	// Return top 5
	if len(similar) > 5 {
		similar = similar[:5]
	}

	return similar
}

// countSharedRelationships counts shared relationships between two nodes.
func (e *SemanticInferenceEngine) countSharedRelationships(nodeA, nodeB string) int {
	targetsA := make(map[string]bool)
	for _, rel := range e.network.outgoing[nodeA] {
		targetsA[rel.TargetID] = true
	}

	count := 0
	for _, rel := range e.network.outgoing[nodeB] {
		if targetsA[rel.TargetID] {
			count++
		}
	}
	return count
}

// ============================================================================
// Emergent Concept Formation
// ============================================================================

// ConceptLearner learns new concepts from experience patterns.
type ConceptLearner struct {
	network               *SemanticNetwork
	minExamplesForConcept int
	similarityThreshold   float64
}

// NewConceptLearner creates a new concept learner.
func NewConceptLearner(network *SemanticNetwork) *ConceptLearner {
	return &ConceptLearner{
		network:               network,
		minExamplesForConcept: 3,
		similarityThreshold:   0.7,
	}
}

// LearnedConcept represents a newly learned concept.
type LearnedConcept struct {
	ID               string
	Label            string
	PrototypeNode    *SemanticNode
	Instances        []string
	CommonProperties map[string]interface{}
	Confidence       float64
	LearnedAt        time.Time
}

// ExtractPrototype creates a prototype from a set of instances.
func (cl *ConceptLearner) ExtractPrototype(instanceIDs []string) (*LearnedConcept, error) {
	if len(instanceIDs) < cl.minExamplesForConcept {
		return nil, fmt.Errorf("need at least %d examples, got %d",
			cl.minExamplesForConcept, len(instanceIDs))
	}

	cl.network.mu.RLock()
	defer cl.network.mu.RUnlock()

	// Collect instances
	instances := make([]*SemanticNode, 0, len(instanceIDs))
	for _, id := range instanceIDs {
		if node, exists := cl.network.nodes[id]; exists {
			instances = append(instances, node)
		}
	}

	if len(instances) < cl.minExamplesForConcept {
		return nil, fmt.Errorf("not enough valid instances found")
	}

	// Find common properties
	commonProps := cl.findCommonProperties(instances)

	// Create prototype node
	prototypeID := fmt.Sprintf("proto_%d", time.Now().UnixNano())
	prototypeLabel := cl.generatePrototypeLabel(instances)

	prototype := NewSemanticNode(prototypeID, prototypeLabel, ConceptNode)
	prototype.Source = "learned"
	prototype.Confidence = float64(len(commonProps)) / float64(len(instances[0].Properties))

	for k, v := range commonProps {
		prototype.SetProperty(k, v)
	}

	// Average embedding if available
	prototype.Embedding = cl.averageEmbeddings(instances)

	learned := &LearnedConcept{
		ID:               prototypeID,
		Label:            prototypeLabel,
		PrototypeNode:    prototype,
		Instances:        instanceIDs,
		CommonProperties: commonProps,
		Confidence:       prototype.Confidence,
		LearnedAt:        time.Now(),
	}

	return learned, nil
}

// findCommonProperties finds properties shared by all instances.
func (cl *ConceptLearner) findCommonProperties(instances []*SemanticNode) map[string]interface{} {
	if len(instances) == 0 {
		return nil
	}

	// Start with first instance's properties
	common := make(map[string]interface{})
	for k, v := range instances[0].Properties {
		common[k] = v
	}

	// Keep only properties that appear in all instances
	for _, inst := range instances[1:] {
		for k, v := range common {
			if instVal, ok := inst.Properties[k]; !ok || !valuesEqual(v, instVal) {
				delete(common, k)
			}
		}
	}

	return common
}

// generatePrototypeLabel creates a label for the prototype.
func (cl *ConceptLearner) generatePrototypeLabel(instances []*SemanticNode) string {
	// Find most common words in labels
	// For simplicity, use "Category of" + first instance type
	if len(instances) > 0 {
		return fmt.Sprintf("Learned Category (%d instances)", len(instances))
	}
	return "Unknown Category"
}

// averageEmbeddings computes the centroid of instance embeddings.
func (cl *ConceptLearner) averageEmbeddings(instances []*SemanticNode) []float32 {
	if len(instances) == 0 {
		return nil
	}

	// Check if instances have embeddings
	dim := 0
	for _, inst := range instances {
		if len(inst.Embedding) > 0 {
			dim = len(inst.Embedding)
			break
		}
	}
	if dim == 0 {
		return nil
	}

	avg := make([]float32, dim)
	count := 0

	for _, inst := range instances {
		if len(inst.Embedding) == dim {
			for i, v := range inst.Embedding {
				avg[i] += v
			}
			count++
		}
	}

	if count > 0 {
		for i := range avg {
			avg[i] /= float32(count)
		}
	}

	return avg
}

// DiscoverRelationships discovers potential relationships between unconnected nodes.
func (cl *ConceptLearner) DiscoverRelationships(minConfidence float64) []*SemanticRelation {
	cl.network.mu.RLock()
	defer cl.network.mu.RUnlock()

	discovered := make([]*SemanticRelation, 0)

	// Find pairs of similar unconnected nodes
	nodes := make([]*SemanticNode, 0, len(cl.network.nodes))
	for _, node := range cl.network.nodes {
		nodes = append(nodes, node)
	}

	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			nodeA := nodes[i]
			nodeB := nodes[j]

			// Skip if already connected
			if cl.areConnected(nodeA.ID, nodeB.ID) {
				continue
			}

			// Compute similarity
			sim, err := cl.network.ComputeSimilarity(nodeA.ID, nodeB.ID)
			if err != nil || sim.Similarity < cl.similarityThreshold {
				continue
			}

			// Infer relationship type from node types
			relType := cl.inferRelationType(nodeA, nodeB)
			if relType == RelatedTo && sim.Similarity < 0.8 {
				continue // Only create generic relations for very similar nodes
			}

			rel := NewSemanticRelation(nodeA.ID, nodeB.ID, relType)
			rel.Weight = sim.Similarity
			rel.Confidence = sim.Similarity * minConfidence
			rel.Source = "discovered"

			discovered = append(discovered, rel)
		}
	}

	return discovered
}

// areConnected checks if two nodes are directly connected.
func (cl *ConceptLearner) areConnected(nodeA, nodeB string) bool {
	for _, rel := range cl.network.outgoing[nodeA] {
		if rel.TargetID == nodeB {
			return true
		}
	}
	for _, rel := range cl.network.incoming[nodeA] {
		if rel.SourceID == nodeB {
			return true
		}
	}
	return false
}

// inferRelationType infers the most likely relationship type between nodes.
func (cl *ConceptLearner) inferRelationType(nodeA, nodeB *SemanticNode) RelationType {
	// Agent to Domain -> BelongsTo
	if nodeA.Type == AgentNode && nodeB.Type == DomainNode {
		return BelongsTo
	}
	// Instance to Concept -> InstanceOf
	if nodeA.Type == InstanceNode && nodeB.Type == ConceptNode {
		return InstanceOf
	}
	// Concept to Concept with same parent -> SimilarTo
	if nodeA.Type == ConceptNode && nodeB.Type == ConceptNode {
		return SimilarTo
	}
	// Action to Concept -> CanDo or UsedFor
	if nodeA.Type == ActionNode && nodeB.Type == ConceptNode {
		return UsedFor
	}
	// Default
	return RelatedTo
}

// CommitLearnedConcept adds a learned concept to the network.
func (cl *ConceptLearner) CommitLearnedConcept(concept *LearnedConcept) error {
	// Add the prototype node
	if err := cl.network.AddNode(concept.PrototypeNode); err != nil {
		return err
	}

	// Create INSTANCE-OF relations from instances to prototype
	for _, instID := range concept.Instances {
		rel := NewSemanticRelation(instID, concept.ID, InstanceOf)
		rel.Confidence = concept.Confidence
		rel.Source = "learned"
		if err := cl.network.AddRelation(rel); err != nil {
			// Continue even if some relations fail
			continue
		}
	}

	cl.network.mu.Lock()
	cl.network.stats.ConceptsLearned++
	cl.network.mu.Unlock()

	return nil
}

// LearnFromExperience creates concepts from experience tuples.
func (cl *ConceptLearner) LearnFromExperience(experiences []*ExperienceTuple) ([]*LearnedConcept, error) {
	if len(experiences) == 0 {
		return nil, nil
	}

	learned := make([]*LearnedConcept, 0)

	// Group experiences by agent
	byAgent := make(map[string][]*ExperienceTuple)
	for _, exp := range experiences {
		byAgent[exp.AgentID] = append(byAgent[exp.AgentID], exp)
	}

	// For each agent with enough experiences, extract a concept
	for agentID, agentExps := range byAgent {
		if len(agentExps) < cl.minExamplesForConcept {
			continue
		}

		// Create nodes from experiences if they don't exist
		instanceIDs := make([]string, 0)
		for _, exp := range agentExps {
			expNodeID := fmt.Sprintf("exp_%s", exp.TaskSignature)

			if _, err := cl.network.GetNode(expNodeID); err != nil {
				// Create a node for this experience
				expNode := NewSemanticNode(expNodeID, exp.Strategy, InstanceNode)
				expNode.SetProperty("agent", agentID)
				expNode.SetProperty("success", exp.Success)
				expNode.SetProperty("fitness", exp.FitnessScore)
				expNode.Embedding = exp.Embedding
				expNode.Source = "experience"

				if err := cl.network.AddNode(expNode); err == nil {
					instanceIDs = append(instanceIDs, expNodeID)
				}
			} else {
				instanceIDs = append(instanceIDs, expNodeID)
			}
		}

		// Extract prototype if we have enough instances
		if len(instanceIDs) >= cl.minExamplesForConcept {
			concept, err := cl.ExtractPrototype(instanceIDs)
			if err == nil {
				concept.Label = fmt.Sprintf("%s Strategy Pattern", agentID)
				learned = append(learned, concept)
			}
		}
	}

	return learned, nil
}

// valuesEqual compares two interface values for equality.
func valuesEqual(a, b interface{}) bool {
	// Simple comparison - could be extended for deep equality
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
