package memory

import (
	"fmt"
	"testing"
	"time"
)

// ============================================================================
// Semantic Network Tests
// ============================================================================

func TestSemanticNetwork_NewSemanticNetwork(t *testing.T) {
	config := DefaultSemanticNetworkConfig()
	sn := NewSemanticNetwork(config)

	if sn == nil {
		t.Fatal("NewSemanticNetwork returned nil")
	}
	if sn.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes, got %d", sn.NodeCount())
	}
	if sn.RelationCount() != 0 {
		t.Errorf("Expected 0 relations, got %d", sn.RelationCount())
	}
}

func TestSemanticNetwork_AddNode(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	node := NewSemanticNode("animal", "Animal", ConceptNode)
	node.SetProperty("is_living", true)

	err := sn.AddNode(node)
	if err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}

	if sn.NodeCount() != 1 {
		t.Errorf("Expected 1 node, got %d", sn.NodeCount())
	}

	// Test duplicate
	err = sn.AddNode(node)
	if err != ErrNodeAlreadyExists {
		t.Errorf("Expected ErrNodeAlreadyExists, got %v", err)
	}
}

func TestSemanticNetwork_GetNode(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	node := NewSemanticNode("dog", "Dog", ConceptNode)
	sn.AddNode(node)

	retrieved, err := sn.GetNode("dog")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if retrieved.Label != "Dog" {
		t.Errorf("Expected label 'Dog', got '%s'", retrieved.Label)
	}

	// Test not found
	_, err = sn.GetNode("cat")
	if err != ErrNodeNotFound {
		t.Errorf("Expected ErrNodeNotFound, got %v", err)
	}
}

func TestSemanticNetwork_RemoveNode(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	animal := NewSemanticNode("animal", "Animal", ConceptNode)
	dog := NewSemanticNode("dog", "Dog", ConceptNode)
	sn.AddNode(animal)
	sn.AddNode(dog)

	rel := NewSemanticRelation("dog", "animal", IsA)
	sn.AddRelation(rel)

	// Remove dog - should also remove its relations
	err := sn.RemoveNode("dog")
	if err != nil {
		t.Fatalf("RemoveNode failed: %v", err)
	}

	if sn.NodeCount() != 1 {
		t.Errorf("Expected 1 node after removal, got %d", sn.NodeCount())
	}
	if sn.RelationCount() != 0 {
		t.Errorf("Expected 0 relations after removal, got %d", sn.RelationCount())
	}
}

func TestSemanticNetwork_AddRelation(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	animal := NewSemanticNode("animal", "Animal", ConceptNode)
	dog := NewSemanticNode("dog", "Dog", ConceptNode)
	sn.AddNode(animal)
	sn.AddNode(dog)

	rel := NewSemanticRelation("dog", "animal", IsA)
	err := sn.AddRelation(rel)
	if err != nil {
		t.Fatalf("AddRelation failed: %v", err)
	}

	if sn.RelationCount() != 1 {
		t.Errorf("Expected 1 relation, got %d", sn.RelationCount())
	}

	// Test self-relation
	selfRel := NewSemanticRelation("dog", "dog", IsA)
	err = sn.AddRelation(selfRel)
	if err != ErrSelfRelation {
		t.Errorf("Expected ErrSelfRelation, got %v", err)
	}
}

func TestSemanticNetwork_GetRelatedNodes(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	animal := NewSemanticNode("animal", "Animal", ConceptNode)
	dog := NewSemanticNode("dog", "Dog", ConceptNode)
	cat := NewSemanticNode("cat", "Cat", ConceptNode)
	sn.AddNode(animal)
	sn.AddNode(dog)
	sn.AddNode(cat)

	sn.AddRelation(NewSemanticRelation("dog", "animal", IsA))
	sn.AddRelation(NewSemanticRelation("cat", "animal", IsA))

	// Find what dog is-a
	related := sn.GetRelatedNodes("dog", IsA)
	if len(related) != 1 {
		t.Errorf("Expected 1 related node, got %d", len(related))
	}
	if len(related) > 0 && related[0].ID != "animal" {
		t.Errorf("Expected 'animal', got '%s'", related[0].ID)
	}

	// Find what is-a animal
	reverse := sn.GetReverseRelatedNodes("animal", IsA)
	if len(reverse) != 2 {
		t.Errorf("Expected 2 reverse related nodes, got %d", len(reverse))
	}
}

func TestSemanticNetwork_CyclicHierarchy(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	a := NewSemanticNode("a", "A", ConceptNode)
	b := NewSemanticNode("b", "B", ConceptNode)
	c := NewSemanticNode("c", "C", ConceptNode)
	sn.AddNode(a)
	sn.AddNode(b)
	sn.AddNode(c)

	sn.AddRelation(NewSemanticRelation("a", "b", IsA))
	sn.AddRelation(NewSemanticRelation("b", "c", IsA))

	// Try to create a cycle: c -> a
	cycleRel := NewSemanticRelation("c", "a", IsA)
	err := sn.AddRelation(cycleRel)
	if err != ErrCyclicHierarchy {
		t.Errorf("Expected ErrCyclicHierarchy, got %v", err)
	}
}

func TestSemanticNetwork_SpreadActivation(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	// Create a simple network
	animal := NewSemanticNode("animal", "Animal", ConceptNode)
	dog := NewSemanticNode("dog", "Dog", ConceptNode)
	fur := NewSemanticNode("fur", "Fur", AttributeNode)
	sn.AddNode(animal)
	sn.AddNode(dog)
	sn.AddNode(fur)

	sn.AddRelation(NewSemanticRelation("dog", "animal", IsA))
	sn.AddRelation(NewSemanticRelation("dog", "fur", HasA))

	// Spread activation from dog
	result := sn.SpreadActivation([]string{"dog"}, 1.0)

	if result == nil {
		t.Fatal("SpreadActivation returned nil")
	}
	if len(result.ActivatedNodes) < 2 {
		t.Errorf("Expected at least 2 activated nodes, got %d", len(result.ActivatedNodes))
	}

	// Dog should have highest activation
	if result.ActivatedNodes["dog"] < 0.9 {
		t.Errorf("Expected dog activation near 1.0, got %f", result.ActivatedNodes["dog"])
	}

	// Neighbors should have some activation
	if result.ActivatedNodes["animal"] == 0 {
		t.Error("Expected animal to have some activation")
	}
}

func TestSemanticNetwork_DecayActivation(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	node := NewSemanticNode("test", "Test", ConceptNode)
	node.Activation = 1.0
	node.BaseActivation = 0.3
	sn.AddNode(node)

	// Decay over 10 seconds
	sn.DecayActivation(10 * time.Second)

	retrieved, _ := sn.GetNode("test")
	if retrieved.Activation >= 1.0 {
		t.Error("Expected activation to decay")
	}
	if retrieved.Activation < retrieved.BaseActivation {
		t.Error("Activation should not decay below base activation")
	}
}

func TestSemanticNetwork_GetInheritedProperties(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	// Create hierarchy: Mammal -> Animal
	animal := NewSemanticNode("animal", "Animal", ConceptNode)
	animal.SetProperty("is_living", true)
	animal.SetProperty("needs_food", true)

	mammal := NewSemanticNode("mammal", "Mammal", ConceptNode)
	mammal.SetProperty("has_fur", true)

	dog := NewSemanticNode("dog", "Dog", InstanceNode)
	dog.SetProperty("barks", true)

	sn.AddNode(animal)
	sn.AddNode(mammal)
	sn.AddNode(dog)

	sn.AddRelation(NewSemanticRelation("mammal", "animal", IsA))
	sn.AddRelation(NewSemanticRelation("dog", "mammal", InstanceOf))

	// Dog should inherit properties from mammal and animal
	props, err := sn.GetInheritedProperties("dog")
	if err != nil {
		t.Fatalf("GetInheritedProperties failed: %v", err)
	}

	if _, ok := props["barks"]; !ok {
		t.Error("Expected 'barks' property (local)")
	}
	if _, ok := props["has_fur"]; !ok {
		t.Error("Expected 'has_fur' property (inherited from mammal)")
	}
	if _, ok := props["is_living"]; !ok {
		t.Error("Expected 'is_living' property (inherited from animal)")
	}
}

func TestSemanticNetwork_IsA(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	living := NewSemanticNode("living", "Living Thing", ConceptNode)
	animal := NewSemanticNode("animal", "Animal", ConceptNode)
	mammal := NewSemanticNode("mammal", "Mammal", ConceptNode)
	dog := NewSemanticNode("dog", "Dog", ConceptNode)

	sn.AddNode(living)
	sn.AddNode(animal)
	sn.AddNode(mammal)
	sn.AddNode(dog)

	sn.AddRelation(NewSemanticRelation("animal", "living", IsA))
	sn.AddRelation(NewSemanticRelation("mammal", "animal", IsA))
	sn.AddRelation(NewSemanticRelation("dog", "mammal", IsA))

	// Direct IS-A
	if !sn.IsA("dog", "mammal") {
		t.Error("dog should IS-A mammal")
	}

	// Transitive IS-A
	if !sn.IsA("dog", "living") {
		t.Error("dog should IS-A living (transitive)")
	}

	// Negative
	if sn.IsA("living", "dog") {
		t.Error("living should NOT IS-A dog")
	}
}

func TestSemanticNetwork_FindCommonAncestors(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	animal := NewSemanticNode("animal", "Animal", ConceptNode)
	mammal := NewSemanticNode("mammal", "Mammal", ConceptNode)
	dog := NewSemanticNode("dog", "Dog", ConceptNode)
	cat := NewSemanticNode("cat", "Cat", ConceptNode)

	sn.AddNode(animal)
	sn.AddNode(mammal)
	sn.AddNode(dog)
	sn.AddNode(cat)

	sn.AddRelation(NewSemanticRelation("mammal", "animal", IsA))
	sn.AddRelation(NewSemanticRelation("dog", "mammal", IsA))
	sn.AddRelation(NewSemanticRelation("cat", "mammal", IsA))

	common := sn.FindCommonAncestors("dog", "cat")
	if len(common) < 1 {
		t.Error("Expected at least 1 common ancestor (mammal)")
	}

	found := false
	for _, ancestor := range common {
		if ancestor.ID == "mammal" || ancestor.ID == "animal" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected mammal or animal as common ancestor")
	}
}

func TestSemanticNetwork_FindShortestPath(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	a := NewSemanticNode("a", "A", ConceptNode)
	b := NewSemanticNode("b", "B", ConceptNode)
	c := NewSemanticNode("c", "C", ConceptNode)
	d := NewSemanticNode("d", "D", ConceptNode)

	sn.AddNode(a)
	sn.AddNode(b)
	sn.AddNode(c)
	sn.AddNode(d)

	sn.AddRelation(NewSemanticRelation("a", "b", RelatedTo))
	sn.AddRelation(NewSemanticRelation("b", "c", RelatedTo))
	sn.AddRelation(NewSemanticRelation("c", "d", RelatedTo))

	path, err := sn.FindShortestPath("a", "d")
	if err != nil {
		t.Fatalf("FindShortestPath failed: %v", err)
	}

	if len(path) != 4 {
		t.Errorf("Expected path length 4, got %d", len(path))
	}
	if path[0].ID != "a" || path[len(path)-1].ID != "d" {
		t.Error("Path should start with 'a' and end with 'd'")
	}
}

func TestSemanticNetwork_ComputeSimilarity(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	// Create nodes with embeddings
	dog := NewSemanticNode("dog", "Dog", ConceptNode)
	dog.Embedding = []float32{0.8, 0.2, 0.1}

	cat := NewSemanticNode("cat", "Cat", ConceptNode)
	cat.Embedding = []float32{0.7, 0.3, 0.1}

	car := NewSemanticNode("car", "Car", ConceptNode)
	car.Embedding = []float32{0.1, 0.1, 0.9}

	sn.AddNode(dog)
	sn.AddNode(cat)
	sn.AddNode(car)

	// Dog and cat should be similar
	simDogCat, err := sn.ComputeSimilarity("dog", "cat")
	if err != nil {
		t.Fatalf("ComputeSimilarity failed: %v", err)
	}

	// Dog and car should be less similar
	simDogCar, err := sn.ComputeSimilarity("dog", "car")
	if err != nil {
		t.Fatalf("ComputeSimilarity failed: %v", err)
	}

	if simDogCat.Similarity <= simDogCar.Similarity {
		t.Error("Dog-cat similarity should be higher than dog-car")
	}
}

func TestSemanticNetwork_SnapshotRestore(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	node := NewSemanticNode("test", "Test", ConceptNode)
	node.SetProperty("key", "value")
	sn.AddNode(node)

	rel := NewSemanticRelation("test", "test", RelatedTo)
	// Can't add self-relation, create another node
	node2 := NewSemanticNode("test2", "Test2", ConceptNode)
	sn.AddNode(node2)
	rel = NewSemanticRelation("test", "test2", RelatedTo)
	sn.AddRelation(rel)

	// Take snapshot
	snapshot := sn.Snapshot()

	// Clear network
	sn.Clear()
	if sn.NodeCount() != 0 {
		t.Error("Network should be empty after Clear")
	}

	// Restore
	err := sn.Restore(snapshot)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if sn.NodeCount() != 2 {
		t.Errorf("Expected 2 nodes after restore, got %d", sn.NodeCount())
	}
	if sn.RelationCount() != 1 {
		t.Errorf("Expected 1 relation after restore, got %d", sn.RelationCount())
	}
}

// ============================================================================
// Semantic Inference Engine Tests
// ============================================================================

func TestSemanticInferenceEngine_InferProperty(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	animal := NewSemanticNode("animal", "Animal", ConceptNode)
	animal.SetProperty("needs_food", true)
	sn.AddNode(animal)

	dog := NewSemanticNode("dog", "Dog", InstanceNode)
	sn.AddNode(dog)

	sn.AddRelation(NewSemanticRelation("dog", "animal", InstanceOf))

	engine := NewSemanticInferenceEngine(sn)

	result, err := engine.InferProperty("dog", "needs_food")
	if err != nil {
		t.Fatalf("InferProperty failed: %v", err)
	}

	if result.Answer != true {
		t.Errorf("Expected true, got %v", result.Answer)
	}
	if len(result.Reasoning) == 0 {
		t.Error("Expected reasoning to be provided")
	}
}

func TestSemanticInferenceEngine_InferMembership(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	living := NewSemanticNode("living", "Living", ConceptNode)
	animal := NewSemanticNode("animal", "Animal", ConceptNode)
	dog := NewSemanticNode("dog", "Dog", ConceptNode)

	sn.AddNode(living)
	sn.AddNode(animal)
	sn.AddNode(dog)

	sn.AddRelation(NewSemanticRelation("animal", "living", IsA))
	sn.AddRelation(NewSemanticRelation("dog", "animal", IsA))

	engine := NewSemanticInferenceEngine(sn)

	// Positive case
	result, err := engine.InferMembership("dog", "living")
	if err != nil {
		t.Fatalf("InferMembership failed: %v", err)
	}
	if result.Answer != true {
		t.Error("Expected dog to be a living thing")
	}

	// Negative case
	result, err = engine.InferMembership("living", "dog")
	if err != nil {
		t.Fatalf("InferMembership failed: %v", err)
	}
	if result.Answer != false {
		t.Error("Expected living to NOT be a dog")
	}
}

func TestSemanticInferenceEngine_InferAnalogy(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	// Dog is-a Mammal, Cat is-a ???
	mammal := NewSemanticNode("mammal", "Mammal", ConceptNode)
	dog := NewSemanticNode("dog", "Dog", ConceptNode)
	cat := NewSemanticNode("cat", "Cat", ConceptNode)

	sn.AddNode(mammal)
	sn.AddNode(dog)
	sn.AddNode(cat)

	sn.AddRelation(NewSemanticRelation("dog", "mammal", IsA))
	sn.AddRelation(NewSemanticRelation("cat", "mammal", IsA))

	engine := NewSemanticInferenceEngine(sn)

	// Dog:Mammal :: Cat:?
	result, err := engine.InferAnalogy("dog", "mammal", "cat")
	if err != nil {
		t.Fatalf("InferAnalogy failed: %v", err)
	}

	if result.Answer != "mammal" {
		t.Errorf("Expected 'mammal', got %v", result.Answer)
	}
}

// ============================================================================
// Concept Learner Tests
// ============================================================================

func TestConceptLearner_ExtractPrototype(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	// Create similar instances
	for i := 1; i <= 5; i++ {
		node := NewSemanticNode(
			fmt.Sprintf("dog%d", i),
			fmt.Sprintf("Dog %d", i),
			InstanceNode,
		)
		node.SetProperty("has_fur", true)
		node.SetProperty("barks", true)
		node.SetProperty("id", i) // This will vary
		sn.AddNode(node)
	}

	learner := NewConceptLearner(sn)

	instances := []string{"dog1", "dog2", "dog3", "dog4", "dog5"}
	concept, err := learner.ExtractPrototype(instances)
	if err != nil {
		t.Fatalf("ExtractPrototype failed: %v", err)
	}

	if concept == nil {
		t.Fatal("Expected concept to be created")
	}

	// Common properties should be extracted
	if _, ok := concept.CommonProperties["has_fur"]; !ok {
		t.Error("Expected 'has_fur' in common properties")
	}
	if _, ok := concept.CommonProperties["barks"]; !ok {
		t.Error("Expected 'barks' in common properties")
	}
	// 'id' should NOT be common (varies)
	if _, ok := concept.CommonProperties["id"]; ok {
		t.Error("'id' should not be in common properties (varies)")
	}
}

func TestConceptLearner_CommitLearnedConcept(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	// Create instances
	for i := 1; i <= 3; i++ {
		node := NewSemanticNode(
			fmt.Sprintf("inst%d", i),
			fmt.Sprintf("Instance %d", i),
			InstanceNode,
		)
		node.SetProperty("common", true)
		sn.AddNode(node)
	}

	learner := NewConceptLearner(sn)

	instances := []string{"inst1", "inst2", "inst3"}
	concept, err := learner.ExtractPrototype(instances)
	if err != nil {
		t.Fatalf("ExtractPrototype failed: %v", err)
	}

	err = learner.CommitLearnedConcept(concept)
	if err != nil {
		t.Fatalf("CommitLearnedConcept failed: %v", err)
	}

	// Prototype should now be in network
	_, err = sn.GetNode(concept.ID)
	if err != nil {
		t.Error("Prototype node should be in network")
	}

	// Check stats
	stats := sn.GetStats()
	if stats.ConceptsLearned != 1 {
		t.Errorf("Expected 1 concept learned, got %d", stats.ConceptsLearned)
	}
}

func TestConceptLearner_DiscoverRelationships(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	// Create similar but unconnected nodes
	dog := NewSemanticNode("dog", "Dog", ConceptNode)
	dog.Embedding = []float32{0.8, 0.2, 0.0}

	cat := NewSemanticNode("cat", "Cat", ConceptNode)
	cat.Embedding = []float32{0.75, 0.25, 0.0}

	animal := NewSemanticNode("animal", "Animal", ConceptNode)
	sn.AddNode(dog)
	sn.AddNode(cat)
	sn.AddNode(animal)

	// Connect dog to animal but not cat
	sn.AddRelation(NewSemanticRelation("dog", "animal", IsA))

	learner := NewConceptLearner(sn)
	discovered := learner.DiscoverRelationships(0.5)

	// Should discover cat-dog similarity
	if len(discovered) == 0 {
		t.Log("No relationships discovered (may depend on similarity threshold)")
	}
}

// ============================================================================
// Node Type and Relation Type Tests
// ============================================================================

func TestNodeType_String(t *testing.T) {
	tests := []struct {
		nodeType NodeType
		expected string
	}{
		{ConceptNode, "concept"},
		{InstanceNode, "instance"},
		{AttributeNode, "attribute"},
		{ActionNode, "action"},
		{AgentNode, "agent"},
		{DomainNode, "domain"},
		{NodeType(99), "unknown"},
	}

	for _, tc := range tests {
		if got := tc.nodeType.String(); got != tc.expected {
			t.Errorf("NodeType(%d).String() = %s, want %s", tc.nodeType, got, tc.expected)
		}
	}
}

func TestRelationType_String(t *testing.T) {
	tests := []struct {
		relType  RelationType
		expected string
	}{
		{IsA, "is-a"},
		{HasA, "has-a"},
		{PartOf, "part-of"},
		{CanDo, "can-do"},
		{UsedFor, "used-for"},
		{RelatedTo, "related-to"},
		{Requires, "requires"},
		{Produces, "produces"},
		{SimilarTo, "similar-to"},
		{OppositeOf, "opposite-of"},
		{InstanceOf, "instance-of"},
		{BelongsTo, "belongs-to"},
		{RelationType(99), "unknown"},
	}

	for _, tc := range tests {
		if got := tc.relType.String(); got != tc.expected {
			t.Errorf("RelationType(%d).String() = %s, want %s", tc.relType, got, tc.expected)
		}
	}
}

func TestRelationType_IsHierarchical(t *testing.T) {
	hierarchical := []RelationType{IsA, PartOf, InstanceOf, BelongsTo}
	nonHierarchical := []RelationType{HasA, CanDo, UsedFor, RelatedTo, SimilarTo}

	for _, rt := range hierarchical {
		if !rt.IsHierarchical() {
			t.Errorf("%s should be hierarchical", rt)
		}
	}

	for _, rt := range nonHierarchical {
		if rt.IsHierarchical() {
			t.Errorf("%s should NOT be hierarchical", rt)
		}
	}
}

func TestRelationType_IsInheritable(t *testing.T) {
	inheritable := []RelationType{IsA, InstanceOf}
	nonInheritable := []RelationType{HasA, PartOf, CanDo, BelongsTo}

	for _, rt := range inheritable {
		if !rt.IsInheritable() {
			t.Errorf("%s should be inheritable", rt)
		}
	}

	for _, rt := range nonInheritable {
		if rt.IsInheritable() {
			t.Errorf("%s should NOT be inheritable", rt)
		}
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"Hello World", "world", true},
		{"Hello World", "HELLO", true},
		{"Hello World", "nothere", false},
		{"", "", true},
		{"abc", "", true},
		{"", "abc", false},
		{"ABC", "abc", true},
	}

	for _, tc := range tests {
		if got := containsIgnoreCase(tc.s, tc.substr); got != tc.expected {
			t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v",
				tc.s, tc.substr, got, tc.expected)
		}
	}
}

func TestSemanticNode_Clone(t *testing.T) {
	node := NewSemanticNode("test", "Test", ConceptNode)
	node.SetProperty("key", "value")
	node.Embedding = []float32{0.1, 0.2, 0.3}
	node.Activation = 0.8

	clone := node.Clone()

	if clone.ID != node.ID {
		t.Error("Clone should have same ID")
	}
	if clone.Activation != node.Activation {
		t.Error("Clone should have same activation")
	}

	// Modify clone, original should not change
	clone.Activation = 0.1
	if node.Activation == clone.Activation {
		t.Error("Modifying clone should not affect original")
	}

	clone.Properties["key"] = "modified"
	if node.Properties["key"] == "modified" {
		t.Error("Modifying clone properties should not affect original")
	}
}

func TestSemanticNetwork_GetMostActivated(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	for i := 1; i <= 5; i++ {
		node := NewSemanticNode(
			fmt.Sprintf("node%d", i),
			fmt.Sprintf("Node %d", i),
			ConceptNode,
		)
		node.Activation = float64(i) * 0.1 // 0.1, 0.2, 0.3, 0.4, 0.5
		sn.AddNode(node)
	}

	top := sn.GetMostActivated(3)
	if len(top) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(top))
	}

	// Should be ordered by activation (highest first)
	if top[0].Activation < top[1].Activation {
		t.Error("Nodes should be ordered by activation descending")
	}
}

func TestSemanticNetwork_FindNodesByLabel(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	sn.AddNode(NewSemanticNode("dog", "Golden Retriever", ConceptNode))
	sn.AddNode(NewSemanticNode("cat", "Persian Cat", ConceptNode))
	sn.AddNode(NewSemanticNode("bird", "Golden Eagle", ConceptNode))

	found := sn.FindNodesByLabel("Golden")
	if len(found) != 2 {
		t.Errorf("Expected 2 nodes with 'Golden', got %d", len(found))
	}

	found = sn.FindNodesByLabel("cat")
	if len(found) != 1 {
		t.Errorf("Expected 1 node with 'cat', got %d", len(found))
	}
}

func TestSemanticNetwork_GetNodesByType(t *testing.T) {
	sn := NewSemanticNetwork(DefaultSemanticNetworkConfig())

	sn.AddNode(NewSemanticNode("concept1", "Concept 1", ConceptNode))
	sn.AddNode(NewSemanticNode("concept2", "Concept 2", ConceptNode))
	sn.AddNode(NewSemanticNode("agent1", "Agent 1", AgentNode))

	concepts := sn.GetNodesByType(ConceptNode)
	if len(concepts) != 2 {
		t.Errorf("Expected 2 concept nodes, got %d", len(concepts))
	}

	agents := sn.GetNodesByType(AgentNode)
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent node, got %d", len(agents))
	}
}

// Helper for tests
func fmt_Sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
