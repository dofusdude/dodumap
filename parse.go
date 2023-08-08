package dodumap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

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

func isActiveEffect(name map[string]string) bool {
	regex := regexp.MustCompile(`^\(.*\)$`)
	if regex.Match([]byte(name["en"])) {
		return true
	}
	if strings.Contains(name["de"], "(Ziel)") {
		return true
	}
	return false
}

func ParseEffects(data *JSONGameData, allEffects [][]JSONGameItemPossibleEffect, langs *map[string]LangDict) [][]MappedMultilangEffect {
	var mappedAllEffects [][]MappedMultilangEffect
	for _, effects := range allEffects {
		var mappedEffects []MappedMultilangEffect
		for _, effect := range effects {

			var mappedEffect MappedMultilangEffect
			currentEffect := data.effects[effect.EffectId]

			numIsSpell := false
			if strings.Contains((*langs)["de"].Texts[currentEffect.DescriptionId], "Zauberspruchs #1") || strings.Contains((*langs)["de"].Texts[currentEffect.DescriptionId], "Zaubers #1") {
				numIsSpell = true
			}

			mappedEffect.Type = make(map[string]string)
			mappedEffect.Templated = make(map[string]string)
			var minMaxRemove int
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
					templatedName, minMaxRemove = NumSpellFormatter(templatedName, lang, data, langs, &diceNum, &diceSide, &value, currentEffect.DescriptionId, numIsSpell, currentEffect.UseDice)
					if templatedName == "" { // found effect that should be discarded for now
						break
					}
					templatedName = SingularPluralFormatter(templatedName, effect.MinimumValue, lang)

					effectName = DeleteDamageFormatter(effectName)
					effectName = SingularPluralFormatter(effectName, effect.MinimumValue, lang)

					mappedEffect.Min = diceNum
					mappedEffect.Max = diceSide
					mappedEffect.Type[lang] = effectName
					mappedEffect.Templated[lang] = templatedName
					mappedEffect.IsMeta = false
				}

				if lang == "en" && mappedEffect.Type[lang] == "" {
					break
				}
			}

			if mappedEffect.Type["en"] == "()" || mappedEffect.Type["en"] == "" {
				continue
			}

			mappedEffect.Active = isActiveEffect(mappedEffect.Type)
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

func ParseCondition(condition string, langs *map[string]LangDict, data *JSONGameData) []MappedMultilangCondition {
	if condition == "" || (!strings.Contains(condition, "&") && !strings.Contains(condition, "<") && !strings.Contains(condition, ">")) {
		return nil
	}

	condition = strings.ReplaceAll(condition, "\n", "")

	lower := strings.ToLower(condition)

	var outs []MappedMultilangCondition

	var parts []string
	if strings.Contains(lower, "&") {
		parts = strings.Split(lower, "&")
	} else {
		parts = []string{lower}
	}

	operators := []string{"<", ">", "=", "!"}

	for _, part := range parts {
		var out MappedMultilangCondition
		out.Templated = make(map[string]string)

		foundCond := false
		for _, operator := range operators { // try every known operator against it
			if strings.Contains(part, operator) {
				var outTmp MappedMultilangCondition
				outTmp.Templated = make(map[string]string)
				foundConditionElement := ConditionWithOperator(part, operator, langs, &out, data)
				if foundConditionElement {
					foundCond = true
				}
			}
		}

		if foundCond {
			outs = append(outs, out)
		}
	}

	if len(outs) == 0 {
		return nil
	}

	return outs
}

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
	file, err := os.ReadFile(filepath.Join(dir, "data", fileSource))
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

	return &data
}

func ParseLangDict(langCode string, dir string) LangDict {
	var err error

	dataPath := filepath.Join(dir, "data", "languages")
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
