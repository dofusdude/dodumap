package dodumap

import "testing"

func TestInterpretedCompoundConditions(t *testing.T) {
	langs := map[string]LangDictUnity{"en": {}}
	data := &JSONGameDataUnity{ConditionReferences: map[string]map[int]ConditionReference{
		"PO": {21172: {AnkamaID: 21172, Name: map[string]string{"en": "Pollenectar Net"}}},
		"PJ": {41: {AnkamaID: 41, Name: map[string]string{"en": "Hunter"}}},
	}}
	age := ParseConditionUnity("DD<30,21172", &langs, data).Value
	if age.Kind != "item_received_age" || age.Unit != "days" || age.Value != 30 || age.Reference == nil || age.Reference.AnkamaID != 21172 || age.ServerExpression != "" {
		t.Fatalf("item age not interpreted: %#v", age)
	}
	job := ParseConditionUnity("Pj>41,9", &langs, data).Value
	if job.Kind != "profession_level" || job.Value != 9 || job.Reference == nil || job.Reference.AnkamaID != 41 {
		t.Fatalf("profession not interpreted: %#v", job)
	}
	equipped := ParseConditionUnity("POE21172", &langs, data).Value
	if equipped.Kind != "item_equipped" || equipped.Operator != "=" || equipped.Quantity != nil {
		t.Fatalf("equipment not interpreted: %#v", equipped)
	}
}

func TestUnityConditionsKeepBranchesAndCompoundValues(t *testing.T) {
	langs := map[string]LangDictUnity{"en": {}}
	data := &JSONGameDataUnity{}
	tree := ParseConditionUnity("PL>9&PL<31|ZZ<30,21172&XY=944", &langs, data)
	if tree.Relation == nil || *tree.Relation != "or" || len(tree.Children) != 2 {
		t.Fatal("OR lost")
	}
	left, right := tree.Children[0], tree.Children[1]
	if *left.Relation != "and" || *right.Relation != "and" {
		t.Fatal("precedence lost")
	}
	if left.Children[0].Value.Element != "PL" {
		t.Fatal("level not decoded")
	}
	if right.Children[0].Value.ServerExpression != "ZZ<30,21172" || right.Children[1].Value.ServerExpression != "XY=944" {
		t.Fatal("server rules lost")
	}
}

func TestUnityConditionsResolveInventoryAndKeepCase(t *testing.T) {
	langs := map[string]LangDictUnity{"en": {}}
	data := &JSONGameDataUnity{ConditionReferences: map[string]map[int]ConditionReference{"PO": {42: {AnkamaID: 42, Name: map[string]string{"en": "Item"}}}}}
	c := ParseConditionUnity("PO<42,10", &langs, data).Value
	if c.Reference == nil || c.Reference.AnkamaID != 42 || c.Quantity == nil || *c.Quantity != 10 {
		t.Fatalf("inventory lost: %#v", c)
	}
	if ParseConditionUnity("Po=42", &langs, data).Value.Kind != "area_location" {
		t.Fatal("case conflated")
	}
	if ParseConditionUnity("ZZ=944", &langs, data) == nil {
		t.Fatal("unclassified single equality discarded")
	}
}

func TestUnityConditionsPreserveParentheses(t *testing.T) {
	langs := map[string]LangDictUnity{"en": {}}
	tree := ParseConditionUnity("(PL>9|PL<5)&(ZZ=944|DD<30,21172)", &langs, &JSONGameDataUnity{})
	if *tree.Relation != "and" || len(tree.Children) != 2 || *tree.Children[0].Relation != "or" || *tree.Children[1].Relation != "or" {
		t.Fatal("nested logic lost")
	}
}

func TestUnityClientConditionScopeExcludesConfirmedServerRules(t *testing.T) {
	langs := map[string]LangDictUnity{"en": {}}
	for _, expression := range []string{"Sc=944", "Sv=946,5", "RV>7,Raid_Score,9999", "PX=A", "ha!52", "Pc=11", "SD=5"} {
		if ParseConditionUnity(expression, &langs, &JSONGameDataUnity{}) != nil {
			t.Fatalf("server condition emitted: %s", expression)
		}
	}
	tree := ParseConditionUnity("PL>9&Sc=944&Sv=946,5", &langs, &JSONGameDataUnity{})
	if tree == nil || !tree.IsOperand || tree.Value.Kind != "character_level" {
		t.Fatal("client condition projection failed")
	}
	if ParseConditionUnity("ZZ=12", &langs, &JSONGameDataUnity{}) == nil {
		t.Fatal("unclassified code silently treated as server-only")
	}
	if ParseConditionUnity("PL>9|Sc=944", &langs, &JSONGameDataUnity{}) != nil {
		t.Fatal("static true OR incorrectly became a restriction")
	}
	tree = ParseConditionUnity("(PL>9|Sc=944)&PL<31", &langs, &JSONGameDataUnity{})
	if tree == nil || !tree.IsOperand || tree.Value.Value != 31 {
		t.Fatal("nested static truth incorrectly simplified")
	}
}
