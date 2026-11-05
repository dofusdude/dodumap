package dodumap

import (
	"strconv"
	"strings"
)

// Meanings checked against DofusDB /criterion and the client criterion factory.
// References: https://api.dofusdb.fr/criterion/DD%3C30%2C21172?lang=en
// https://github.com/bot4dofus/Datafus/blob/master/data/A/DofusInvoker/scripts/com/ankamagames/dofus/datacenter/items/criterion/ItemCriterionFactory.as
type conditionMeaning struct{ kind, name, reference, unit string }

var additionalCharacteristicLabels = map[string]map[string]string{
	"cs": {"en": "Additional Strength", "de": "Zusätzliche Stärke", "fr": "Force additionnelle", "es": "Fuerza adicional", "pt": "Força adicional"},
	"ci": {"en": "Additional Intelligence", "de": "Zusätzliche Intelligenz", "fr": "Intelligence additionnelle", "es": "Inteligencia adicional", "pt": "Inteligência adicional"},
	"cv": {"en": "Additional Vitality", "de": "Zusätzliche Vitalität", "fr": "Vitalité additionnelle", "es": "Vitalidad adicional", "pt": "Vitalidade adicional"},
	"ca": {"en": "Additional Agility", "de": "Zusätzliche Flinkheit", "fr": "Agilité additionnelle", "es": "Agilidad adicional", "pt": "Agilidade adicional"},
	"cc": {"en": "Additional Chance", "de": "Zusätzliches Glück", "fr": "Chance additionnelle", "es": "Suerte adicional", "pt": "Sorte adicional"},
	"cw": {"en": "Additional Wisdom", "de": "Zusätzliche Weisheit", "fr": "Sagesse additionnelle", "es": "Sabiduría adicional", "pt": "Sabedoria adicional"},
}

var conditionMeanings = map[string]conditionMeaning{
	"CS": {"strength", "Strength", "", ""}, "CI": {"intelligence", "Intelligence", "", ""},
	"CV": {"vitality", "Vitality", "", ""}, "CA": {"agility", "Agility", "", ""},
	"CC": {"chance", "Chance", "", ""}, "CW": {"wisdom", "Wisdom", "", ""},
	"CP": {"action_points", "Action points", "", "AP"}, "CM": {"movement_points", "Movement points", "", "MP"},
	"cs": {"additional_strength", "Additional Strength", "", ""}, "ci": {"additional_intelligence", "Additional Intelligence", "", ""},
	"cv": {"additional_vitality", "Additional Vitality", "", ""}, "ca": {"additional_agility", "Additional Agility", "", ""},
	"cc": {"additional_chance", "Additional Chance", "", ""}, "cw": {"additional_wisdom", "Additional Wisdom", "", ""},
	"PL": {"character_level", "Character level", "", ""},
	"PO": {"inventory_quantity", "Inventory quantity", "PO", "items"},
	"DD": {"item_received_age", "Time since receiving item", "PO", "days"},
	"DH": {"item_received_age", "Time since receiving item", "PO", "hours"},
	"Pj": {"profession_level", "Profession level", "PJ", "levels"},
	"PJ": {"profession_learned", "Profession learned", "PJ", ""},
	"CU": {"hunting_weapon_equipped", "Hunting weapon equipped", "", ""},
	"Pm": {"map_location", "Current map", "Pm", ""},
	"Po": {"area_location", "Current area", "Po", ""},
	"PB": {"subarea_location", "Current subarea", "PB", ""},
	"Qo": {"quest_objective", "Quest objective", "Qo", ""},
	"Qs": {"quest_step", "Quest step", "Qs", ""},
	"Qa": {"quest_active", "Quest active", "Qa", ""},
	"Qf": {"quest_completed", "Quest completed", "Qf", ""},
	"Qc": {"quest_available", "Quest available", "Qc", ""},
	"HS": {"combat_state", "Combat state", "HS", ""},
	"HA": {"alteration_active", "Alteration active", "HA", ""},
	"ST": {"server_season", "Active server season", "", ""},
	"SC": {"server_game_type", "Server game type", "SC", ""},
	"SI": {"server", "Server", "", ""},
	"Az": {"area_anomaly", "Area affected by an anomaly", "", ""},
	"Ms": {"arena_map", "Arena map", "", ""},
	"Mt": {"dungeon_map", "Dungeon map", "", ""},
	"PZ": {"subscription_active", "Subscription active", "", ""},
	"Pz": {"restricted_account_mode", "Restricted account mode", "", ""},
	"OA": {"achievement_completed", "Achievement completed", "OA", ""},
	"OO": {"achievement_objective_completed", "Achievement objective completed", "OO", ""},
	"PT": {"spell_learned", "Spell learned", "PT", ""},
	"PE": {"emote_learned", "Emote learned", "PE", ""},
	"Ot": {"title_unlocked", "Title unlocked", "Ot", ""},
	"Os": {"smiley_pack_unlocked", "Smiley pack unlocked", "Os", ""},
	"OH": {"havenbag_theme_unlocked", "Haven bag theme unlocked", "OH", ""},
	"OS": {"only_equipped_set", "Only this item set equipped", "OS", ""},
	"PG": {"character_class", "Character class", "PG", ""},
	"Ps": {"alignment", "Character alignment", "Ps", ""},
	"Pa": {"alignment_level", "Alignment level", "", "levels"},
	"SG": {"month", "Month (zero-based game calendar)", "", ""},
	"Sd": {"day_of_month", "Day of month", "", ""},
	"BI": {"item_unequippable", "Item cannot be equipped", "", ""},
	"Ga": {"character_alive", "Character alive", "", ""},
	"Oa": {"achievement_points", "Achievement points", "", "points"},
	"OV": {"subscription_duration", "Accumulated subscription duration", "", "days"},
	"Pw": {"guild_member", "Member of a guild", "", ""},
	"GM": {"guild_master", "Guild master", "", ""},
	"AM": {"alliance_general", "Alliance general", "", ""},
	"Ow": {"alliance_member", "Member of an alliance", "", ""},
	"OL": {"missing_health", "Missing health", "", "HP"},
	"PR": {"married", "Married", "", ""},
	"PS": {"character_sex", "Character sex (0 male, 1 female)", "", ""},
	"PW": {"carried_weight", "Carried weight", "", "pods"},
	"PK": {"currency", "Kamas", "", "kamas"},
	"SL": {"territory_level", "Territory level", "", "levels"},
	"PC": {"map_cell", "Current map cell", "", ""},
	"Ox": {"alliance_right", "Alliance right", "Ox", ""},
	"EM": {"monster_kill_count", "Monster kills", "EM", "kills"},
	"Pn": {"equipment_slot", "Equipment slot", "", ""},
	"Mp": {"map_prism_state", "Prism state (3 vulnerable)", "", ""},
	"pk": {"equipped_set_bonus_count", "Equipped set bonus count", "", "bonuses"},
	"Ss": {"subarea_monster_present", "Monster present in subarea", "PB", ""},
	"WE": {"world_event_active", "World event active", "WE", ""},
	"Ad": {"almanax_calendar_day", "Almanax calendar day", "Ad", ""},
}

func interpretCondition(expression string, langs *map[string]LangDictUnity, data *JSONGameDataUnity) (MappedMultilangCondition, bool) {
	parts := conditionAtomUnity.FindStringSubmatch(expression)
	if parts == nil {
		return MappedMultilangCondition{}, false
	}
	code, op := parts[1], parts[2]
	args := strings.Split(parts[3], ",")
	if code == "Oc" && parts[3] == "R" && op == "=" {
		out := MappedMultilangCondition{Kind: "subarea_captured_by_alliance", Element: "Oc", ElementId: -1, Operator: "=", Value: 0, ValueLabel: "not captured", Templated: map[string]string{}}
		for lang := range *langs {
			out.Templated[lang] = "Subarea captured by an alliance"
		}
		return out, true
	}
	if code == "PN" {
		out := MappedMultilangCondition{Kind: "character_name", Element: "PN", ElementId: -1, Operator: op, StringValue: parts[3], Templated: map[string]string{}}
		for lang := range *langs {
			out.Templated[lang] = "Character name"
		}
		return out, true
	}
	meaning, ok := conditionMeanings[code]
	if !ok {
		return MappedMultilangCondition{}, false
	}
	value, err := strconv.Atoi(args[0])
	if err != nil {
		return MappedMultilangCondition{}, false
	}
	out := MappedMultilangCondition{Kind: meaning.kind, Element: code, ElementId: -1, Operator: op, Value: value, Unit: meaning.unit, Templated: map[string]string{}}
	for lang := range *langs {
		out.Templated[lang] = meaning.name
	}
	refID := value
	if code == "PO" && (op == "E" || op == "X") {
		out.Kind = "item_equipped"
		for lang := range *langs {
			out.Templated[lang] = "Item equipped"
		}
		out.Operator = "="
		if op == "X" {
			out.Operator = "!"
		}
	}
	switch code {
	case "Ss":
		if len(args) != 2 {
			return out, false
		}
		monsterID, e := strconv.Atoi(args[1])
		if e != nil {
			return out, false
		}
		monster, found := data.ConditionReferences["EM"][monsterID]
		if !found {
			return out, false
		}
		out.SecondaryReference = &monster
	case "DD", "DH":
		if len(args) != 2 {
			return out, false
		}
		refID, err = strconv.Atoi(args[1])
		if err != nil {
			return out, false
		}
	case "EM":
		if len(args) != 2 {
			return out, false
		}
		out.Value, err = strconv.Atoi(args[1])
		if err != nil {
			return out, false
		}
	case "PO":
		if len(args) > 2 {
			return out, false
		}
		quantity := 1
		if len(args) == 2 {
			quantity, err = strconv.Atoi(args[1])
			if err != nil {
				return out, false
			}
		}
		if op != "E" && op != "X" {
			out.Quantity = &quantity
			out.Value = quantity
		}
	case "PJ", "Pj":
		if len(args) > 2 {
			return out, false
		}
		if len(args) == 2 {
			out.Value, err = strconv.Atoi(args[1])
			if err != nil {
				return out, false
			}
		} else if op == ">" || op == "<" {
			out.Value = 0
		}
		if op == ">" || op == "<" || len(args) == 2 {
			out.Kind = "profession_level"
			for lang := range *langs {
				out.Templated[lang] = "Profession level"
			}
		} else {
			out.Kind = "profession_learned"
			for lang := range *langs {
				out.Templated[lang] = "Profession learned"
			}
		}
	default:
		if len(args) != 1 {
			return out, false
		}
	}
	if meaning.reference != "" {
		ref, found := data.ConditionReferences[meaning.reference][refID]
		if !found && meaning.reference == "PO" {
			if item, exists := data.Items[refID]; exists {
				ref = ConditionReference{AnkamaID: refID, Name: map[string]string{}}
				for lang, dict := range *langs {
					ref.Name[lang] = dict.Texts[item.NameId]
				}
				found = true
			}
		}
		// Missing catalog text does not change the meaning of a verified
		// predicate. Keep the typed ID; never invent an entity name.
		if !found {
			ref = ConditionReference{AnkamaID: refID}
		}
		out.Reference = &ref
	}
	if code == "Qo" {
		out.Kind = "quest_objective_active"
		name := "Quest objective active"
		if op == ">" || op == "<" {
			out.Kind = "quest_objective_completed"
			name = "Quest objective completed"
			if op == ">" {
				out.Operator = "="
			} else {
				out.Operator = "!"
			}
		}
		for lang := range *langs {
			out.Templated[lang] = name
		}
	}
	if code == "PS" {
		if value == 0 {
			out.ValueLabel = "male"
		} else if value == 1 {
			out.ValueLabel = "female"
		}
	}
	if code == "Mp" && value == 3 {
		out.ValueLabel = "vulnerable"
	}
	if code == "Pn" && value == -1 {
		out.ValueLabel = "unequipped"
	}
	return out, true
}
