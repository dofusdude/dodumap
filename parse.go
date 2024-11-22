package dodumap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/log"
	"github.com/emirpasic/gods/maps/treebidimap"
	gutils "github.com/emirpasic/gods/utils"
)

type PersistentStringKeysMap struct {
	Entries *treebidimap.Map `json:"entries"`
	NextId  int              `json:"next_id"`
}

var (
	PersistedElements PersistentStringKeysMap
	PersistedTypes    PersistentStringKeysMap
)

func LoadPersistedElements(persistenceDir string, release string) error {
	var elements []string
	var types []string
	if persistenceDir == "" {
		elementUrl := fmt.Sprintf("https://raw.githubusercontent.com/dofusdude/doduda/main/persistent/elements.%s.json", release)
		elementResponse, err := http.Get(elementUrl)
		if err != nil {
			log.Fatal(err)
		}

		elementBody, err := io.ReadAll(elementResponse.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(elementBody, &elements)
		if err != nil {
			log.Fatal(err)
		}

		itemTypeUrl := fmt.Sprintf("https://raw.githubusercontent.com/dofusdude/doduda/main/persistent/item_types.%s.json", release)
		itemTypeResponse, err := http.Get(itemTypeUrl)
		if err != nil {
			log.Fatal(err)
		}

		itemTypeBody, err := io.ReadAll(itemTypeResponse.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(itemTypeBody, &types)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		data, err := os.ReadFile(filepath.Join(persistenceDir, fmt.Sprintf("elements.%s.json", release)))
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &elements)
		if err != nil {
			fmt.Println(err)
		}

		data, err = os.ReadFile(filepath.Join(persistenceDir, fmt.Sprintf("item_types.%s.json", release)))
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &types)
		if err != nil {
			fmt.Println(err)
		}
	}

	PersistedElements = PersistentStringKeysMap{
		Entries: treebidimap.NewWith(gutils.IntComparator, gutils.StringComparator),
		NextId:  0,
	}

	for _, entry := range elements {
		PersistedElements.Entries.Put(PersistedElements.NextId, entry)
		PersistedElements.NextId++
	}

	PersistedTypes = PersistentStringKeysMap{
		Entries: treebidimap.NewWith(gutils.IntComparator, gutils.StringComparator),
		NextId:  0,
	}

	for _, entry := range types {
		PersistedTypes.Entries.Put(PersistedTypes.NextId, entry)
		PersistedTypes.NextId++
	}

	return nil
}

func PersistElements(elementPath string, itemTypePath string) error {
	elements := make([]string, PersistedElements.NextId)
	it := PersistedElements.Entries.Iterator()
	for it.Next() {
		elements[it.Key().(int)] = it.Value().(string)
	}

	elementsJson, err := json.MarshalIndent(elements, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(elementPath, elementsJson, 0644)
	if err != nil {
		return err
	}

	types := make([]string, PersistedTypes.NextId)
	it = PersistedTypes.Entries.Iterator()
	for it.Next() {
		types[it.Key().(int)] = it.Value().(string)
	}

	typesJson, err := json.MarshalIndent(types, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(itemTypePath, typesJson, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ParseItemCombo(rawEffects [][]*JSONGameItemPossibleEffect, effects [][]MappedMultilangEffect) [][]MappedMultilangSetEffect {
	var mappedEffects [][]MappedMultilangSetEffect
	i := 0
	for itemComboCounter, effectsPerCombo := range rawEffects {
		var mappedEffectsPerCombo []MappedMultilangSetEffect

		j := 0
		for _, effect := range effectsPerCombo {
			if effect == nil {
				continue
			}

			parsedEffects := effects[i]
			parsedEffect := parsedEffects[j]
			setEffect := MappedMultilangSetEffect{
				Min:              parsedEffect.Min,
				Max:              parsedEffect.Max,
				Type:             parsedEffect.Type,
				Templated:        parsedEffect.Templated,
				Active:           parsedEffect.Active,
				ElementId:        parsedEffect.ElementId,
				IsMeta:           parsedEffect.IsMeta,
				MinMaxIrrelevant: parsedEffect.MinMaxIrrelevant,
				ItemCombination:  uint(itemComboCounter + 1),
			}
			mappedEffectsPerCombo = append(mappedEffectsPerCombo, setEffect)
			j += 1
		}
		if len(mappedEffectsPerCombo) > 0 {
			mappedEffects = append(mappedEffects, mappedEffectsPerCombo)
			i += 1
		}
	}
	return mappedEffects
}

func ParseEffects(data *JSONGameData, allEffects [][]*JSONGameItemPossibleEffect, langs *map[string]LangDict) [][]MappedMultilangEffect {
	var mappedAllEffects [][]MappedMultilangEffect
	for _, effects := range allEffects {
		var mappedEffects []MappedMultilangEffect
		for _, effect := range effects {

			if effect == nil {
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
			for _, lang := range Languages {
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
					templatedName, minMaxRemove = NumSpellFormatter(templatedName, lang, data, langs, &diceNum, &diceSide, &value, currentEffect.DescriptionId, numIsSpell, currentEffect.UseDice, &frNumSigned, &frSideSigned)
					if templatedName == "" { // found effect that should be discarded for now
						break
					}
					templatedName = SingularPluralFormatter(templatedName, effect.MinimumValue, lang)

					if isTitle { // titles are Title: 0 after formatting; TODO move this into the NumSpellFormatter
						templatedName = strings.ReplaceAll(templatedName, "0", (*langs)[lang].Texts[data.titles[diceNum].NameMaleId]) // TODO male default, idk how to make it neutral yet
					}

					effectName = DeleteDamageFormatter(effectName)
					effectName = SingularPluralFormatter(effectName, effect.MinimumValue, lang)

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
				continue
			}

			mappedEffect.Active = currentEffect.UseInFight
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

			mappedEffects = append(mappedEffects, mappedEffect)
		}
		if len(mappedEffects) > 0 {
			mappedAllEffects = append(mappedAllEffects, mappedEffects)
		}
	}
	if len(mappedAllEffects) == 0 {
		return nil
	}
	return mappedAllEffects
}

// NewNode creates a new Node
func newNode(value string, nodeType NodeType) *ConditionTreeNode {
	return &ConditionTreeNode{
		Value:    value,
		Type:     nodeType,
		Children: []*ConditionTreeNode{},
	}
}

// AddChild adds a child node
func (n *ConditionTreeNode) AddChild(child *ConditionTreeNode) {
	n.Children = append(n.Children, child)
}

func ParseExpression(exp string) *ConditionTreeNode {
	var stack []*ConditionTreeNode
	var current *ConditionTreeNode
	var operandBuilder strings.Builder

	conditionOperators := []rune{'<', '>', '=', '!'}

	for _, char := range exp {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || slices.Contains(conditionOperators, char) {
			operandBuilder.WriteRune(char) // continue building operand
		} else {
			if operandBuilder.Len() > 0 {
				// finalize and add the operand
				if current != nil && current.Type == Operator {
					current.AddChild(newNode(operandBuilder.String(), Operand))
				} else {
					current = newNode(operandBuilder.String(), Operand)
				}
				operandBuilder.Reset()
			}

			switch char {
			case '(':
				if current != nil {
					stack = append(stack, current)
				}
				current = nil
			case ')':
				if len(stack) > 0 {
					parent := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					parent.AddChild(current)
					current = parent
				}
			case '&', '|': // expression operators
				operator := newNode(string(char), Operator)
				if current != nil {
					operator.AddChild(current)
				}
				current = operator
			}
		}
	}

	if operandBuilder.Len() > 0 {
		if current != nil && current.Type == Operator {
			current.AddChild(newNode(operandBuilder.String(), Operand))
		} else {
			current = newNode(operandBuilder.String(), Operand)
		}
	}

	return current
}

func atomicCondition(expression string, langs *map[string]LangDict, data *JSONGameData) (bool, MappedMultilangCondition) {
	operators := []string{"<", ">", "=", "!"}

	var out MappedMultilangCondition
	out.Templated = make(map[string]string)

	foundCond := false
	for _, operator := range operators { // try every known operator against it
		if strings.Contains(expression, operator) {
			foundConditionElement := ConditionWithOperator(expression, operator, langs, &out, data)
			if foundConditionElement {
				foundCond = true
				break
			}
		}
	}

	return foundCond, out
}

func removeUnsupportedExpressions(node *ConditionTreeNode, langs *map[string]LangDict, data *JSONGameData) *ConditionTreeNode {
	if node == nil {
		return nil
	}

	// If the node is an operand and not supported, return nil
	validCond, _ := atomicCondition(node.Value, langs, data)
	if node.Type == Operand && !validCond {
		return nil
	}

	// Process children
	var validChildren []*ConditionTreeNode
	for _, child := range node.Children {
		processedChild := removeUnsupportedExpressions(child, langs, data)
		if processedChild != nil {
			validChildren = append(validChildren, processedChild)
		}
	}
	node.Children = validChildren

	return node
}

func simplifyTree(node *ConditionTreeNode) *ConditionTreeNode {
	if node == nil {
		return nil
	}

	// Process children
	var validChildren []*ConditionTreeNode
	for _, child := range node.Children {
		processedChild := simplifyTree(child)
		if processedChild != nil {
			validChildren = append(validChildren, processedChild)
		}
	}
	node.Children = validChildren

	// If an operator node has only one child, replace it with its child
	if node.Type == Operator && len(node.Children) == 1 {
		return node.Children[0]
	}

	// when no children in operator, return nil
	if node.Type == Operator && len(node.Children) == 0 {
		return nil
	}

	return node
}

func buildMappedConditionTree(out **ConditionTreeNodeMapped, root *ConditionTreeNode, langs *map[string]LangDict, data *JSONGameData) {
	if root == nil {
		return
	}

	if root.Type == Operand {
		foundCond, mappedCond := atomicCondition(root.Value, langs, data)
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
			buildMappedConditionTree(childOut, child, langs, data)
			(*out).Children[i] = *childOut
		}
		return
	} else {
		log.Fatal("unknown node type")
	}
}

func PrintTree(node *ConditionTreeNode, level int) {
	if node == nil {
		return
	}

	fmt.Printf("%s%s\n", strings.Repeat(" ", level*2), node.Value)
	for _, child := range node.Children {
		PrintTree(child, level+1)
	}
}

func buildHistoricAndConnectionArray(mappedTree *ConditionTreeNodeMapped, out *[]MappedMultilangCondition) {
	if mappedTree == nil {
		return
	}

	if mappedTree.IsOperand {
		*out = append(*out, *mappedTree.Value)
		return
	}

	if *mappedTree.Relation == "and" {
		for _, child := range mappedTree.Children {
			buildHistoricAndConnectionArray(child, out)
		}
	}
}

func ParseCondition(condition string, langs *map[string]LangDict, data *JSONGameData) ([]MappedMultilangCondition, *ConditionTreeNodeMapped) {
	if condition == "" || (!strings.Contains(condition, "&") && !strings.Contains(condition, "|") && !strings.Contains(condition, "<") && !strings.Contains(condition, ">")) {
		return nil, nil
	}

	// parse into ast
	tree := ParseExpression(condition)

	// strip tree to only known conditions
	tree = removeUnsupportedExpressions(tree, langs, data)
	tree = simplifyTree(tree)

	if tree == nil {
		return nil, nil
	}

	// convert to mapped tree
	mappedTree := new(*ConditionTreeNodeMapped)
	buildMappedConditionTree(mappedTree, tree, langs, data)

	if *mappedTree == nil {
		log.Fatal("mapped tree is nil")
	}

	// for historical reasons, still return the old format but only for &-connected conditions
	// check the tree and combine all children that are connected with & to a single array
	var mappedConditions []MappedMultilangCondition
	buildHistoricAndConnectionArray(*mappedTree, &mappedConditions)
	if len(mappedConditions) == 0 {
		mappedConditions = nil
	}

	return mappedConditions, *mappedTree
}

// TODO: previous "conditions" convert to only be with simple & operator

type HasId interface {
	GetID() int
}

func CleanJSON(jsonStr string) string {
	jsonStr = strings.ReplaceAll(jsonStr, "NaN", "null")
	jsonStr = strings.ReplaceAll(jsonStr, "\"null\"", "null")
	jsonStr = strings.ReplaceAll(jsonStr, " ", " ")
	return jsonStr
}

func ParseRawDataPart[T HasId](fileSource string, result chan map[int]T, dir string) {
	file, err := os.ReadFile(filepath.Join(dir, fileSource))
	if err != nil {
		fmt.Print(err)
	}
	fileStr := CleanJSON(string(file))
	var fileJson []T
	err = json.Unmarshal([]byte(fileStr), &fileJson)
	if err != nil {
		fmt.Println(err)
	}
	items := make(map[int]T)
	for _, item := range fileJson {
		items[item.GetID()] = item
	}
	result <- items
}

func ParseRawData(dir string) *JSONGameData {
	var data JSONGameData
	itemChan := make(chan map[int]JSONGameItem)
	itemTypeChan := make(chan map[int]JSONGameItemType)
	itemSetsChan := make(chan map[int]JSONGameSet)
	itemEffectsChan := make(chan map[int]JSONGameEffect)
	itemBonusesChan := make(chan map[int]JSONGameBonus)
	itemRecipesChang := make(chan map[int]JSONGameRecipe)
	spellsChan := make(chan map[int]JSONGameSpell)
	spellTypesChan := make(chan map[int]JSONGameSpellType)
	areasChan := make(chan map[int]JSONGameArea)
	mountsChan := make(chan map[int]JSONGameMount)
	breedsChan := make(chan map[int]JSONGameBreed)
	mountFamilyChan := make(chan map[int]JSONGameMountFamily)
	npcsChan := make(chan map[int]JSONGameNPC)
	titlesChan := make(chan map[int]JSONGameTitle)
	questsChan := make(chan map[int]JSONGameQuest)
	questObjectivesChan := make(chan map[int]JSONGameQuestObjective)
	questStepRewardsChan := make(chan map[int]JSONGameQuestStepRewards)
	questCategoriesChan := make(chan map[int]JSONGameQuestCategory)
	questStepsChan := make(chan map[int]JSONGameQuestStep)
	almanaxCalendarsChan := make(chan map[int]JSONGameAlamanaxCalendar)

	go func() {
		ParseRawDataPart("npcs.json", npcsChan, dir)
	}()
	go func() {
		ParseRawDataPart("mount_family.json", mountFamilyChan, dir)
	}()
	go func() {
		ParseRawDataPart("breeds.json", breedsChan, dir)
	}()
	go func() {
		ParseRawDataPart("mounts.json", mountsChan, dir)
	}()
	go func() {
		ParseRawDataPart("areas.json", areasChan, dir)
	}()
	go func() {
		ParseRawDataPart("spell_types.json", spellTypesChan, dir)
	}()
	go func() {
		ParseRawDataPart("spells.json", spellsChan, dir)
	}()
	go func() {
		ParseRawDataPart("recipes.json", itemRecipesChang, dir)
	}()
	go func() {
		ParseRawDataPart("items.json", itemChan, dir)
	}()
	go func() {
		ParseRawDataPart("item_types.json", itemTypeChan, dir)
	}()
	go func() {
		ParseRawDataPart("item_sets.json", itemSetsChan, dir)
	}()
	go func() {
		ParseRawDataPart("bonuses.json", itemBonusesChan, dir)
	}()
	go func() {
		ParseRawDataPart("effects.json", itemEffectsChan, dir)
	}()
	go func() {
		ParseRawDataPart("titles.json", titlesChan, dir)
	}()
	go func() {
		ParseRawDataPart("quests.json", questsChan, dir)
	}()
	go func() {
		ParseRawDataPart("quest_objectives.json", questObjectivesChan, dir)
	}()
	go func() {
		ParseRawDataPart("quest_step_rewards.json", questStepRewardsChan, dir)
	}()
	go func() {
		ParseRawDataPart("quest_categories.json", questCategoriesChan, dir)
	}()
	go func() {
		ParseRawDataPart("almanax.json", almanaxCalendarsChan, dir)
	}()
	go func() {
		ParseRawDataPart("quest_steps.json", questStepsChan, dir)
	}()

	data.Items = <-itemChan
	close(itemChan)

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

	data.spellTypes = <-spellTypesChan
	close(spellTypesChan)

	data.areas = <-areasChan
	close(areasChan)

	data.Mounts = <-mountsChan
	close(mountsChan)

	data.classes = <-breedsChan
	close(breedsChan)

	data.MountFamilys = <-mountFamilyChan
	close(mountFamilyChan)

	data.npcs = <-npcsChan
	close(npcsChan)

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

	return &data
}

func ParseLangDict(langCode string, dir string) LangDict {
	var err error

	dataPath := filepath.Join(dir, "languages")
	var data LangDict
	data.IdText = make(map[int]int)
	data.Texts = make(map[int]string)
	data.NameText = make(map[string]int)

	langFile, err := os.ReadFile(filepath.Join(dataPath, langCode+".json"))
	if err != nil {
		fmt.Print(err)
	}

	langFileStr := CleanJSON(string(langFile))
	var langJson JSONLangDict
	err = json.Unmarshal([]byte(langFileStr), &langJson)
	if err != nil {
		fmt.Println(err)
	}

	for key, value := range langJson.IdText {
		keyParsed, err := strconv.Atoi(key)
		if err != nil {
			fmt.Println(err)
		}
		data.IdText[keyParsed] = value
	}

	for key, value := range langJson.Texts {
		keyParsed, err := strconv.Atoi(key)
		if err != nil {
			fmt.Println(err)
		}
		data.Texts[keyParsed] = value
	}
	data.NameText = langJson.NameText
	return data
}

func ParseRawLanguages(dir string) map[string]LangDict {
	data := make(map[string]LangDict)
	for _, lang := range Languages {
		data[lang] = ParseLangDict(lang, dir)
	}
	return data
}
