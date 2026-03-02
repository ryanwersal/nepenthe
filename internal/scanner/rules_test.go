package scanner

import "testing"

func TestBuildSentinelRulesReturnsIndependentCopy(t *testing.T) {
	rules1 := BuildSentinelRules()
	rules2 := BuildSentinelRules()

	if len(rules1) == 0 {
		t.Fatal("expected non-empty rules")
	}
	if len(rules1) != len(rules2) {
		t.Fatal("expected same length")
	}

	// Mutating one copy should not affect the other
	rules1[0].Ecosystem = "MUTATED"
	if rules2[0].Ecosystem == "MUTATED" {
		t.Error("BuildSentinelRules should return independent copies")
	}
}

func TestFixedPathCategories(t *testing.T) {
	cats, err := FixedPathCategories()
	if err != nil {
		t.Fatal(err)
	}

	expectedCategories := []Category{CategoryDevCaches, CategoryContainers, CategoryVMs, CategoryOptional}
	for _, cat := range expectedCategories {
		if _, ok := cats[cat]; !ok {
			t.Errorf("expected category %q in FixedPathCategories", cat)
		}
	}

	// Should not contain dependencies or custom
	if _, ok := cats[CategoryDependencies]; ok {
		t.Error("FixedPathCategories should not contain dependencies")
	}
	if _, ok := cats[CategoryCustom]; ok {
		t.Error("FixedPathCategories should not contain custom")
	}
}
