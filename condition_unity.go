package dodumap

import (
	"regexp"
	"strconv"
	"strings"
)

var conditionAtomUnity = regexp.MustCompile(`^([A-Za-z]{2})([=!<>EX~])(.*)$`)

// ConditionTreeFullyInterpreted is deliberately all-or-nothing: removing an
// unknown AND/OR branch changes eligibility and must never produce public data.
func ConditionTreeFullyInterpreted(root *ConditionTreeNodeMapped) bool {
	if root == nil {
		return false
	}
	if root.IsOperand {
		return root.Value != nil && root.Value.ServerExpression == "" && root.Value.Element != "server_requirement" && len(root.Children) == 0
	}
	if root.Value != nil || root.Relation == nil || (*root.Relation != "and" && *root.Relation != "or") || len(root.Children) < 2 {
		return false
	}
	for _, child := range root.Children {
		if !ConditionTreeFullyInterpreted(child) {
			return false
		}
	}
	return true
}

// parseConditionExpression preserves compound values and case-sensitive codes.
// AND binds more tightly than OR. Invalid/unverified atoms are retained whole.
func parseConditionExpression(expression string) *ConditionTreeNode {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil
	}
	if expression == "null" {
		return nil
	}
	for strings.HasPrefix(expression, "(") && strings.HasSuffix(expression, ")") {
		depth, wraps := 0, true
		for i, r := range expression {
			if r == '(' {
				depth++
			}
			if r == ')' {
				depth--
			}
			if depth == 0 && i < len(expression)-1 {
				wraps = false
				break
			}
		}
		if !wraps {
			break
		}
		expression = expression[1 : len(expression)-1]
	}
	for _, operator := range []rune{'|', '&'} {
		depth, start := 0, 0
		parts := []string{}
		for i, r := range expression {
			if r == '(' {
				depth++
			}
			if r == ')' {
				depth--
			}
			if r == operator && depth == 0 {
				parts = append(parts, expression[start:i])
				start = i + 1
			}
		}
		if len(parts) > 0 {
			parts = append(parts, expression[start:])
			out := parseConditionExpression(parts[0])
			for _, part := range parts[1:] {
				out = &ConditionTreeNode{Type: Operator, Value: string(operator), Children: []*ConditionTreeNode{out, parseConditionExpression(part)}}
			}
			return out
		}
	}
	return &ConditionTreeNode{Type: Operand, Value: expression}
}

func mapConditionAtomUnity(expression string, langs *map[string]LangDictUnity, data *JSONGameDataUnity) MappedMultilangCondition {
	if mapped, ok := interpretCondition(expression, langs, data); ok {
		if (mapped.Element == "PL" || mapped.Element == "PZ" || strings.HasPrefix(mapped.Element, "C") || strings.HasPrefix(mapped.Element, "c")) && PersistedElements.Entries != nil {
			if ok, previous := atomicConditionUnity(expression, langs, data); ok {
				mapped.ElementId = previous.ElementId
				if strings.HasPrefix(mapped.Element, "C") {
					mapped.Templated = previous.Templated
				}
			}
		}
		if labels, ok := additionalCharacteristicLabels[mapped.Element]; ok {
			for lang, label := range labels {
				if _, present := (*langs)[lang]; present {
					mapped.Templated[lang] = label
				}
			}
		}
		return mapped
	}
	out := MappedMultilangCondition{Element: "server_requirement", ElementId: -1, ServerExpression: expression, Templated: map[string]string{}}
	for language := range *langs {
		out.Templated[language] = "Requirement evaluated by the game server"
	}
	parts := conditionAtomUnity.FindStringSubmatch(expression)
	if parts == nil {
		return out
	}
	code, op := parts[1], parts[2]
	args := strings.Split(parts[3], ",")
	value, err := strconv.Atoi(args[0])
	if err != nil {
		return out
	}
	names := map[string]string{"PL": "Character level", "PO": "Inventory item", "Qa": "Quest active", "Qf": "Quest completed", "Qc": "Quest startable", "HS": "Combat state", "OA": "Achievement completed", "PT": "Spell learned", "PZ": "Subscription active", "Sd": "Day of month", "SG": "Month"}
	name, supported := names[code]
	if supported {
		if len(args) > 2 || (len(args) > 1 && code != "PO") {
			return out
		}
		switch code {
		case "PO", "Qa", "Qf", "Qc", "HS", "OA", "PT":
			refs := data.ConditionReferences[code]
			ref, found := refs[value]
			if !found && code == "PO" {
				if item, ok := data.Items[value]; ok {
					ref = ConditionReference{AnkamaID: value, Name: map[string]string{}}
					for language, dictionary := range *langs {
						ref.Name[language] = dictionary.Texts[item.NameId]
					}
					found = true
				}
			}
			if !found || len(ref.Name) == 0 {
				return out
			}
			out.Reference = &ref
		}
		if len(args) == 2 {
			quantity, err := strconv.Atoi(args[1])
			if err != nil {
				return out
			}
			out.Quantity = &quantity
		}
		out.Element, out.Value, out.Operator, out.ServerExpression = code, value, op, ""
		// Keep already published IDs for the established scalar conditions.
		if (code == "PL" || code == "PZ") && PersistedElements.Entries != nil {
			if ok, previous := atomicConditionUnity(expression, langs, data); ok {
				out.ElementId = previous.ElementId
			}
		}
		for language := range *langs {
			out.Templated[language] = name
		}
		return out
	}
	// Existing scalar item characteristics retain their persisted element IDs.
	// Only exact supported spellings enter the legacy formatter.
	legacy := map[string]bool{"cs": true, "ci": true, "cv": true, "ca": true, "cc": true, "cw": true, "CS": true, "Cs": true, "CI": true, "Ci": true, "CV": true, "Cv": true, "CA": true, "Ca": true, "CC": true, "Cc": true, "CW": true, "Cw": true, "PK": true, "Pk": true, "CM": true, "CP": true, "Pa": true, "OF": true, "PF": true}
	if legacy[code] && len(args) == 1 {
		if ok, mapped := atomicConditionUnity(expression, langs, data); ok {
			return mapped
		}
	}
	return out
}

func mapConditionTreeUnity(root *ConditionTreeNode, langs *map[string]LangDictUnity, data *JSONGameDataUnity) *ConditionTreeNodeMapped {
	if root == nil {
		return nil
	}
	if staticClientTautology(root) {
		return nil
	}
	if root.Type == Operand {
		if serverOnlyCondition(root.Value) {
			return nil
		}
		value := mapConditionAtomUnity(root.Value, langs, data)
		return &ConditionTreeNodeMapped{IsOperand: true, Value: &value}
	}
	relation := "and"
	if root.Value == "|" {
		relation = "or"
	}
	out := &ConditionTreeNodeMapped{Relation: &relation, Children: []*ConditionTreeNodeMapped{}}
	for _, child := range root.Children {
		if mapped := mapConditionTreeUnity(child, langs, data); mapped != nil {
			out.Children = append(out.Children, mapped)
		}
	}
	if len(out.Children) == 0 {
		return nil
	}
	if len(out.Children) == 1 {
		return out.Children[0]
	}
	return out
}

// Sc is the client's StaticCriterionCriterion, whose local evaluation always
// succeeds. Preserve that truth through Boolean groups before projecting out
// server-only rules; in particular A OR Sc must not become A.
func staticClientTautology(root *ConditionTreeNode) bool {
	if root == nil {
		return false
	}
	if root.Type == Operand {
		return strings.HasPrefix(root.Value, "Sc")
	}
	if root.Value == "|" {
		for _, child := range root.Children {
			if staticClientTautology(child) {
				return true
			}
		}
		return false
	}
	if len(root.Children) == 0 {
		return false
	}
	for _, child := range root.Children {
		if !staticClientTautology(child) {
			return false
		}
	}
	return true
}

// These are deliberately outside the public client-condition model. Sc is
// StaticCriterionCriterion (always accepts locally); the other server-owned
// predicates have no handler in the current Unity criterion factory. PX is
// privileged account access, not a normal player gameplay constraint.
// Evidence: Jondo.Unity.WebClient Decompiled/dump.cs and mw$$ood/bwf factories
// in Decompiled/decompiled_all.c, inspected 2026-11-02 (current 3.7 client).
func serverOnlyCondition(expression string) bool {
	if len(expression) < 2 {
		return false
	}
	switch expression[:2] {
	case "Sc", "Sv", "RV", "Pc", "ha", "SD", "PX":
		return true
	}
	return false
}
