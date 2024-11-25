package dodumap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

func ParseRawDataPartUnityMulti[T HasId, A any](fileSource string, dir string, selfType string, otherType *string) JsonGameUnityRefLookup[T, A] {
	file, err := os.ReadFile(filepath.Join(dir, fileSource))
	if err != nil {
		log.Fatal(err)
	}
	fileStr := CleanJSON(string(file))

	pattern := `"rid": (-?\d+)`
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal("Error compiling regex:", err)
	}
	replaceFunc := func(match string) string {
		return re.ReplaceAllString(match, `"rid": "$1"`)
	}
	fileStr = re.ReplaceAllStringFunc(fileStr, replaceFunc)

	var fileJson struct {
		References struct {
			Version int `json:"version"`
			RefIds  []struct {
				Rid  string `json:"rid"`
				Type struct {
					Class string `json:"class"`
				}
				Data interface{} // depends on class
			} `json:"RefIds"`
		} `json:"references"`
	}
	err = json.Unmarshal([]byte(fileStr), &fileJson)
	if err != nil {
		log.Fatal(err)
	}

	itemsAnkamaIdLookup := make(map[int]T)
	itemsRefIdLookup := make(map[int64]A)
	for _, entry := range fileJson.References.RefIds {
		rid, err := strconv.ParseInt(entry.Rid, 10, 64)
		if err != nil {
			log.Fatalf("Error parsing rid: %v", err)
		}
		entryType := entry.Type.Class

		if otherType != nil && entryType == *otherType {
			var item A
			payloadBytes, err := json.Marshal(entry.Data)
			if err != nil {
				log.Fatalf("Error marshaling to bytes: %v", err)
			}
			err = json.Unmarshal(payloadBytes, &item)
			if err != nil {
				continue
			}

			itemsRefIdLookup[rid] = item
			continue
		}

		if selfType == "*" || entryType == selfType {
			var item T
			payloadBytes, err := json.Marshal(entry.Data)
			if err != nil {
				log.Fatalf("Error marshaling to bytes: %v", err)
			}
			err = json.Unmarshal(payloadBytes, &item)
			if err != nil {
				log.Fatalf("Error unmarshaling to item: %v", err)
			}

			itemsAnkamaIdLookup[item.GetID()] = item
			continue
		}

		if entryType != "" {
			log.Warn("Unknown type", "type", entryType)
		}
	}

	out := JsonGameUnityRefLookup[T, A]{itemsRefIdLookup, itemsAnkamaIdLookup}
	return out
}

func ParseRawDataPartUnity[T HasId](fileSource string, result chan map[int]T, dir string) {
	file, err := os.ReadFile(filepath.Join(dir, fileSource))
	if err != nil {
		log.Fatal(err)
	}
	fileStr := CleanJSON(string(file))

	pattern := `"rid": (-?\d+)`
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal("Error compiling regex:", err)
	}
	replaceFunc := func(match string) string {
		return re.ReplaceAllString(match, `"rid": "$1"`)
	}
	fileStr = re.ReplaceAllStringFunc(fileStr, replaceFunc)

	var fileJson struct {
		References struct {
			Version int `json:"version"`
			RefIds  []struct {
				Rid  string `json:"rid"`
				Type struct {
					Class string `json:"class"`
				}
				Data interface{} // depends on class
			} `json:"RefIds"`
		} `json:"references"`
	}
	err = json.Unmarshal([]byte(fileStr), &fileJson)
	if err != nil {
		log.Fatal(err)
	}

	itemsAnkamaIdLookup := make(map[int]T)
	for _, entry := range fileJson.References.RefIds {
		var item T
		payloadBytes, err := json.Marshal(entry.Data)
		if err != nil {
			log.Fatalf("Error marshaling to bytes: %v", err)
		}
		err = json.Unmarshal(payloadBytes, &item)
		if err != nil {
			fmt.Println(err)
			continue
		}
		itemsAnkamaIdLookup[item.GetID()] = item
	}

	result <- itemsAnkamaIdLookup
}

type HasMerge[A any, B any] interface {
	Merge(other B) A
}

func ParseRawDataUnity(dir string) *JSONGameDataUnity {
	var data JSONGameDataUnity
	npcsChan := make(chan map[int]JSONGameNPCUnity)
	itemChan := make(chan map[int]JSONGameItemUnity)
	itemSetsChan := make(chan map[int]JSONGameSetUnity)
	itemTypeChan := make(chan map[int]JSONGameItemTypeUnity)
	itemEffectsChan := make(chan map[int]JSONGameEffectUnity)
	itemBonusesChan := make(chan map[int]JSONGameBonusUnity)
	itemRecipesChang := make(chan map[int]JSONGameRecipeUnity)
	spellsChan := make(chan map[int]JSONGameSpellUnity)
	//spellTypesChan := make(chan map[int]JSONGameSpellType)
	areasChan := make(chan map[int]JSONGameAreaUnity)
	mountsChan := make(chan map[int]JSONGameMountUnity)

	breedsChan := make(chan map[int]JSONGameBreedUnity)
	mountFamilyChan := make(chan map[int]JSONGameMountFamilyUnity)

	titlesChan := make(chan map[int]JSONGameTitleUnity)
	questsChan := make(chan map[int]JSONGameQuestUnity)
	questStepsChan := make(chan map[int]JSONGameQuestStepUnity)
	questObjectivesChan := make(chan map[int]JSONGameQuestObjectiveUnity)
	questStepRewardsChan := make(chan map[int]JSONGameQuestStepRewardsUnity)
	questCategoriesChan := make(chan map[int]JSONGameQuestCategoryUnity)
	almanaxCalendarsChan := make(chan map[int]JSONGameAlamanaxCalendarUnity)

	go func() {
		ParseRawDataPartUnity("npcs.json", npcsChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("mount_family.json", mountFamilyChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("breeds.json", breedsChan, dir)
	}()
	go func() {
		possibleEffectInstance := "EffectInstanceDice"
		mountLookup := ParseRawDataPartUnityMulti[JSONGameMountUnityRaw, JSONGameItemPossibleEffectUnity]("mounts.json", dir, "Mounts", &possibleEffectInstance)
		mounts := make(map[int]JSONGameMountUnity)
		for _, mount := range mountLookup.AnkamaId {
			mappedPossibleEffects := make([]*JSONGameItemPossibleEffectUnity, 0)
			for _, possibleEffectRef := range mount.Effects.Array {
				var possibleEffect *JSONGameItemPossibleEffectUnity = nil
				if possibleEffectRef.Ref != "-2" {
					res, err := strconv.ParseInt(possibleEffectRef.Ref, 10, 64)
					if err != nil {
						log.Fatal("Mount parsing", "err", err)
					}
					existingEffect := mountLookup.Ref[res]
					possibleEffect = &existingEffect
				}
				mappedPossibleEffects = append(mappedPossibleEffects, possibleEffect)
			}
			mergedMount := mount.Merge(mappedPossibleEffects)
			mounts[mount.Id] = mergedMount
		}
		mountsChan <- mounts
		return
	}()
	go func() {
		ParseRawDataPartUnity("areas.json", areasChan, dir)
	}()
	/*go func() {
		ParseRawDataPartUnity("spell_types.json", spellTypesChan, dir)
	}()*/
	go func() {
		ParseRawDataPartUnity("spells.json", spellsChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("recipes.json", itemRecipesChang, dir)
	}()
	go func() {
		possibleEffectInstance := "EffectInstanceDice"
		itemLookup := ParseRawDataPartUnityMulti[JSONGameItemUnityRaw, JSONGameItemPossibleEffectUnity]("items.json", dir, "*", &possibleEffectInstance)
		items := make(map[int]JSONGameItemUnity)
		for _, item := range itemLookup.AnkamaId {
			mappedPossibleEffects := make([]*JSONGameItemPossibleEffectUnity, 0)
			for _, possibleEffectRef := range item.PossibleEffects.Array {
				var possibleEffect *JSONGameItemPossibleEffectUnity = nil
				if possibleEffectRef.Ref != "-2" {
					res, err := strconv.ParseInt(possibleEffectRef.Ref, 10, 64)
					if err != nil {
						log.Fatal("Item parsing", "err", err)
					}
					existingEffect := itemLookup.Ref[res]

					possibleEffect = &existingEffect
				}
				mappedPossibleEffects = append(mappedPossibleEffects, possibleEffect)
			}
			mergedItem := item.Merge(mappedPossibleEffects)
			items[item.Id] = mergedItem
		}
		itemChan <- items
		return
	}()
	go func() {
		ParseRawDataPartUnity("item_types.json", itemTypeChan, dir)
	}()
	go func() {
		possibleEffectInstance := "EffectInstanceDice"
		setLookup := ParseRawDataPartUnityMulti[JSONGameSetUnityRaw, JSONGameItemPossibleEffectUnity]("item_sets.json", dir, "ItemSets", &possibleEffectInstance)
		sets := make(map[int]JSONGameSetUnity)
		for _, set := range setLookup.AnkamaId {
			mappedPossibleEffects := make([][]*JSONGameItemPossibleEffectUnity, 0)
			for _, possibleEffectsRef := range set.Effects.Array {
				mappedPossibleEffectsInner := make([]*JSONGameItemPossibleEffectUnity, 0)
				for _, possibleEffectRef := range possibleEffectsRef.Values.Array {
					var possibleEffect *JSONGameItemPossibleEffectUnity = nil
					if possibleEffectRef.Ref != "-2" {
						res, err := strconv.ParseInt(possibleEffectRef.Ref, 10, 64)
						if err != nil {
							log.Fatal("Set parsing", "err", err)
						}
						existingEffect := setLookup.Ref[res]
						possibleEffect = &existingEffect
					}
					mappedPossibleEffectsInner = append(mappedPossibleEffectsInner, possibleEffect)
				}
				mappedPossibleEffects = append(mappedPossibleEffects, mappedPossibleEffectsInner)
			}
			mergedSet := set.Merge(mappedPossibleEffects)
			sets[set.Id] = mergedSet
		}
		itemSetsChan <- sets
		return
	}()
	go func() {
		ParseRawDataPartUnity("bonuses.json", itemBonusesChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("effects.json", itemEffectsChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("titles.json", titlesChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("quests.json", questsChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("quest_objectives.json", questObjectivesChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("quest_step_rewards.json", questStepRewardsChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("quest_categories.json", questCategoriesChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("almanax.json", almanaxCalendarsChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("quest_steps.json", questStepsChan, dir)
	}()

	data.bonuses = <-itemBonusesChan
	close(itemBonusesChan)

	data.effects = <-itemEffectsChan
	close(itemEffectsChan)

	data.ItemTypes = <-itemTypeChan
	close(itemTypeChan)

	data.Sets = <-itemSetsChan
	close(itemSetsChan)

	data.Recipes = <-itemRecipesChang
	close(itemRecipesChang)

	data.spells = <-spellsChan
	close(spellsChan)

	//data.spellTypes = <-spellTypesChan
	//close(spellTypesChan)

	data.areas = <-areasChan
	close(areasChan)

	data.Mounts = <-mountsChan
	close(mountsChan)

	data.classes = <-breedsChan
	close(breedsChan)

	data.MountFamilys = <-mountFamilyChan
	close(mountFamilyChan)

	data.titles = <-titlesChan
	close(titlesChan)

	data.quests = <-questsChan
	close(questsChan)

	data.questObjectives = <-questObjectivesChan
	close(questObjectivesChan)

	data.questStepRewards = <-questStepRewardsChan
	close(questStepRewardsChan)

	data.questCategories = <-questCategoriesChan
	close(questCategoriesChan)

	data.almanaxCalendars = <-almanaxCalendarsChan
	close(almanaxCalendarsChan)

	data.questSteps = <-questStepsChan
	close(questStepsChan)

	data.npcs = <-npcsChan
	close(npcsChan)

	data.Items = <-itemChan
	close(itemChan)

	return &data
}

func ParseRawLanguagesUnity(dir string) map[string]LangDictUnity {
	data := make(map[string]LangDictUnity)
	for _, lang := range LanguagesUnity {
		data[lang] = ParseLangDictUnity(lang, dir)
	}
	return data
}

func ParseLangDictUnity(langCode string, dir string) LangDictUnity {
	var err error

	dataPath := filepath.Join(dir, "languages")
	var data LangDictUnity
	data.Texts = make(map[int]string)

	langFile, err := os.ReadFile(filepath.Join(dataPath, langCode+".json"))
	if err != nil {
		log.Fatal(err)
	}

	langFileStr := CleanJSON(string(langFile))
	var langJson JSONLangDictUnity
	err = json.Unmarshal([]byte(langFileStr), &langJson)
	if err != nil {
		log.Fatal(err)
	}

	for key, value := range langJson.Texts {
		keyParsed, err := strconv.Atoi(key)
		if err != nil {
			log.Fatal("langDict", err)
		}
		data.Texts[keyParsed] = value
	}

	return data
}

func ParseEffectsUnity(data *JSONGameDataUnity, allEffects [][]*JSONGameItemPossibleEffectUnity, langs *map[string]LangDictUnity) [][]*MappedMultilangEffect {
	var mappedAllEffects [][]*MappedMultilangEffect
	for _, effects := range allEffects {
		var mappedEffects []*MappedMultilangEffect
		for _, effect := range effects {

			if effect == nil {
				mappedEffects = append(mappedEffects, nil)
				continue
			}

			var mappedEffect MappedMultilangEffect
			currentEffect := data.effects[effect.EffectId]

			numIsSpell := false
			if strings.Contains((*langs)["de"].Texts[currentEffect.DescriptionId], "Zauberspruchs #1") || strings.Contains((*langs)["de"].Texts[currentEffect.DescriptionId], "Zaubers #1") {
				numIsSpell = true
			}

			isTitle := false
			if strings.Contains((*langs)["en"].Texts[currentEffect.DescriptionId], "Title:") {
				isTitle = true
			}

			mappedEffect.Type = make(map[string]string)
			mappedEffect.Templated = make(map[string]string)
			var minMaxRemove int
			var frNumSigned int = 2  // unset
			var frSideSigned int = 2 // unset
			for _, lang := range LanguagesUnity {
				var diceNum int
				var diceSide int
				var value int

				diceNum = effect.MinimumValue

				diceSide = effect.MaximumValue

				value = effect.Value

				effectName := (*langs)[lang].Texts[currentEffect.DescriptionId]
				if lang == "de" {
					effectName = strings.ReplaceAll(effectName, "{~ps}{~zs}", "") // german has error in template
				}

				if effectName == "#1" { // is spell description from dicenum 1
					effectName = "-special spell-"
					mappedEffect.Min = 0
					mappedEffect.Max = 0
					mappedEffect.Type[lang] = effectName
					mappedEffect.Templated[lang] = (*langs)[lang].Texts[data.spells[diceNum].DescriptionId]
					mappedEffect.IsMeta = true
				} else {
					templatedName := effectName
					var useDice bool
					if currentEffect.UseDice == 0 {
						useDice = false
					} else {
						useDice = true
					}
					templatedName, minMaxRemove = NumSpellFormatterUnity(templatedName, lang, data, langs, &diceNum, &diceSide, &value, currentEffect.DescriptionId, numIsSpell, useDice, &frNumSigned, &frSideSigned)
					if templatedName == "" { // found effect that should be discarded for now
						break
					}
					templatedName = SingularPluralFormatterUnity(templatedName, effect.MinimumValue, lang)

					if isTitle { // titles are Title: 0 after formatting; TODO move this into the NumSpellFormatter
						maleTitleNum, err := strconv.Atoi(data.titles[diceNum].NameMaleId) // TODO male default, idk how to make it neutral yet
						var replTitle string
						if err != nil {
							log.Warn("EffectParsing", "InvalidTitleId", data.titles[diceNum].Id, "PossibleId", diceNum)
							replTitle = "-invalid-"
						} else {
							replTitle = (*langs)[lang].Texts[maleTitleNum]
						}
						templatedName = strings.ReplaceAll(templatedName, "0", replTitle)
					}

					effectName = DeleteDamageFormatterUnity(effectName)
					effectName = SingularPluralFormatterUnity(effectName, 1, lang) // singularize the effect name for comparisons

					if isTitle {
						mappedEffect.Min = 0
						mappedEffect.Max = 0
						mappedEffect.IsMeta = true
					} else {
						mappedEffect.Min = diceNum
						mappedEffect.Max = diceSide
						mappedEffect.IsMeta = false
					}
					mappedEffect.Max = diceSide
					mappedEffect.Type[lang] = effectName
					mappedEffect.Templated[lang] = templatedName
				}

				if lang == "en" && mappedEffect.Type[lang] == "" {
					break
				}
			}

			if mappedEffect.Type["en"] == "()" || mappedEffect.Type["en"] == "" {
				// this happens way too often but we can't do anything about it
				//log.Warn("Effect", "emtpy_type", mappedEffect.Type["en"])
				mappedEffects = append(mappedEffects, nil)
				continue
			}

			if currentEffect.UseInFight == 0 {
				mappedEffect.Active = false
			} else {
				mappedEffect.Active = true
			}
			searchTypeEn := mappedEffect.Type["en"]
			if mappedEffect.Active {
				searchTypeEn += " (Active)"
			}
			key, foundKey := PersistedElements.Entries.GetKey(searchTypeEn)
			if foundKey {
				mappedEffect.ElementId = key.(int)
			} else {
				mappedEffect.ElementId = PersistedElements.NextId
				PersistedElements.Entries.Put(PersistedElements.NextId, searchTypeEn)
				PersistedElements.NextId++
			}

			mappedEffect.MinMaxIrrelevant = minMaxRemove

			mappedEffects = append(mappedEffects, &mappedEffect)
		}
		mappedAllEffects = append(mappedAllEffects, mappedEffects)
	}
	if len(mappedAllEffects) == 0 {
		return nil
	}
	return mappedAllEffects
}

func atomicConditionUnity(expression string, langs *map[string]LangDictUnity, data *JSONGameDataUnity) (bool, MappedMultilangCondition) {
	operators := []string{"<", ">", "=", "!"}

	var out MappedMultilangCondition
	out.Templated = make(map[string]string)

	foundCond := false
	for _, operator := range operators { // try every known operator against it
		if strings.Contains(expression, operator) {
			foundConditionElement := ConditionWithOperatorUnity(expression, operator, langs, &out, data)
			if foundConditionElement {
				foundCond = true
				break
			}
		}
	}

	return foundCond, out
}

func removeUnsupportedExpressionsUnity(node *ConditionTreeNode, langs *map[string]LangDictUnity, data *JSONGameDataUnity) *ConditionTreeNode {
	if node == nil {
		return nil
	}

	// If the node is an operand and not supported, return nil
	validCond, _ := atomicConditionUnity(node.Value, langs, data)
	if node.Type == Operand && !validCond {
		return nil
	}

	// Process children
	var validChildren []*ConditionTreeNode
	for _, child := range node.Children {
		processedChild := removeUnsupportedExpressionsUnity(child, langs, data)
		if processedChild != nil {
			validChildren = append(validChildren, processedChild)
		}
	}
	node.Children = validChildren

	return node
}

func ParseConditionUnity(condition string, langs *map[string]LangDictUnity, data *JSONGameDataUnity) *ConditionTreeNodeMapped {
	if condition == "" || (!strings.Contains(condition, "&") && !strings.Contains(condition, "|") && !strings.Contains(condition, "<") && !strings.Contains(condition, ">")) {
		return nil
	}

	// parse into ast
	tree := ParseExpression(condition)

	// strip tree to only known conditions
	tree = removeUnsupportedExpressionsUnity(tree, langs, data)
	tree = simplifyTree(tree)

	if tree == nil {
		return nil
	}

	// convert to mapped tree
	mappedTree := new(*ConditionTreeNodeMapped)
	buildMappedConditionTreeUnity(mappedTree, tree, langs, data)

	if *mappedTree == nil {
		log.Fatal("mapped tree is nil")
	}

	return *mappedTree
}

func buildMappedConditionTreeUnity(out **ConditionTreeNodeMapped, root *ConditionTreeNode, langs *map[string]LangDictUnity, data *JSONGameDataUnity) {
	if root == nil {
		return
	}

	if root.Type == Operand {
		foundCond, mappedCond := atomicConditionUnity(root.Value, langs, data)
		if !foundCond {
			log.Fatal("condition not found, should be handled before")
		}

		if *out == nil {
			*out = new(ConditionTreeNodeMapped)
		}

		(*out).Value = &mappedCond
		(*out).IsOperand = true
		return
	} else if root.Type == Operator {
		if *out == nil {
			*out = new(ConditionTreeNodeMapped)
		}
		(*out).IsOperand = false
		if root.Value == "&" {
			(*out).Relation = new(string)
			*(*out).Relation = "and"
		} else if root.Value == "|" {
			(*out).Relation = new(string)
			*(*out).Relation = "or"
		} else {
			log.Fatal("unknown operator")
		}
		(*out).Children = make([]*ConditionTreeNodeMapped, len(root.Children))
		for i, child := range root.Children {
			childOut := new(*ConditionTreeNodeMapped)
			buildMappedConditionTreeUnity(childOut, child, langs, data)
			(*out).Children[i] = *childOut
		}
		return
	} else {
		log.Fatal("unknown node type")
	}
}

func ParseItemComboUnity(effects [][]*MappedMultilangEffect) map[int][]MappedMultilangEffect {
	mappedEffects := make(map[int][]MappedMultilangEffect)
	for itemComboCounter, effectsPerCombo := range effects {
		humanComboCounter := itemComboCounter + 1

		mappedEffects[humanComboCounter] = make([]MappedMultilangEffect, 0)

		for _, effect := range effectsPerCombo {
			if effect == nil {
				continue
			}
			setEffect := MappedMultilangEffect{
				Min:              effect.Min,
				Max:              effect.Max,
				Type:             effect.Type,
				Templated:        effect.Templated,
				Active:           effect.Active,
				ElementId:        effect.ElementId,
				IsMeta:           effect.IsMeta,
				MinMaxIrrelevant: effect.MinMaxIrrelevant,
			}
			mappedEffects[humanComboCounter] = append(mappedEffects[humanComboCounter], setEffect)
		}
	}
	return mappedEffects
}
