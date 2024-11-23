package dodumap

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

func ParseRawDataPartUnityMulti[T HasId, A any](fileSource string, result chan JsonGameUnityRefLookup[T, A], dir string, selfType string, otherType *string) {
	file, err := os.ReadFile(filepath.Join(dir, fileSource))
	if err != nil {
		log.Fatal(err)
	}
	fileStr := CleanJSON(string(file))
	var fileJson map[string]interface{}
	err = json.Unmarshal([]byte(fileStr), &fileJson)
	if err != nil {
		log.Fatal(err)
	}

	refs := fileJson["references"].(map[string]interface{})
	refIds := refs["RefIds"].([]map[string]interface{})

	itemsAnkamaIdLookup := make(map[int]T)
	itemsRefIdLookup := make(map[int64]A)
	for _, entry := range refIds {
		rid := entry["rid"].(int64)
		entryType := entry["type"].(struct {
			Class string `json:"class"`
		}).Class

		if entryType == selfType {
			item := entry["data"].(T)
			itemsAnkamaIdLookup[item.GetID()] = item
			continue
		}

		if otherType != nil && entryType == *otherType {
			itemsRefIdLookup[rid] = entry["data"].(A)
			continue
		}

		log.Warn("Unknown type: %s", entryType)
	}

	out := JsonGameUnityRefLookup[T, A]{itemsRefIdLookup, itemsAnkamaIdLookup}
	result <- out
}

func ParseRawDataPartUnity[T HasId](fileSource string, result chan map[int]T, dir string) {
	file, err := os.ReadFile(filepath.Join(dir, fileSource))
	if err != nil {
		log.Fatal(err)
	}
	fileStr := CleanJSON(string(file))
	var fileJson map[string]interface{}
	err = json.Unmarshal([]byte(fileStr), &fileJson)
	if err != nil {
		log.Fatal(err)
	}

	refs := fileJson["references"].(map[string]interface{})
	refIds := refs["RefIds"].([]map[string]interface{})

	itemsAnkamaIdLookup := make(map[int]T)
	for _, entry := range refIds {
		if item, ok := entry["data"].(T); ok {
			itemsAnkamaIdLookup[item.GetID()] = item
		}
	}

	result <- itemsAnkamaIdLookup
}

type HasMerge[A any, B any] interface {
	Merge(other B) A
}

func ParseRawDataUnity(dir string) *JSONGameDataUnity {
	var data JSONGameDataUnity
	itemRawChan := make(chan JsonGameUnityRefLookup[JSONGameItemUnityRaw, JSONGameItemPossibleEffectUnity])
	itemChan := make(chan map[int]JSONGameItemUnity)

	itemSetsRawChan := make(chan JsonGameUnityRefLookup[JSONGameSetUnityRaw, JSONGameItemPossibleEffectUnity])
	itemSetsChan := make(chan map[int]JSONGameSetUnity)

	itemTypeChan := make(chan map[int]JSONGameItemTypeUnity)
	itemEffectsChan := make(chan map[int]JSONGameEffectUnity)
	itemBonusesChan := make(chan map[int]JSONGameBonusUnity)
	itemRecipesChang := make(chan map[int]JSONGameRecipeUnity)
	/*spellsChan := make(chan map[int]JSONGameSpell)
	spellTypesChan := make(chan map[int]JSONGameSpellType)
	areasChan := make(chan map[int]JSONGameArea)*/

	mountsRawChan := make(chan JsonGameUnityRefLookup[JSONGameMountUnityRaw, JSONGameItemPossibleEffectUnity])
	mountsChan := make(chan map[int]JSONGameMountUnity)

	breedsChan := make(chan map[int]JSONGameBreedUnity)
	mountFamilyChan := make(chan map[int]JSONGameMountFamilyUnity)
	npcsChan := make(chan map[int]JSONGameNPCUnity)
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
		ParseRawDataPartUnityMulti("mounts.json", mountsRawChan, dir, "Mounts", &possibleEffectInstance)
		for {
			select {
			case mountLookup := <-mountsRawChan:
				mounts := make(map[int]JSONGameMountUnity)
				for _, mount := range mountLookup.AnkamaId {
					mappedPossibleEffects := make([]JSONGameItemPossibleEffectUnity, 0)
					for _, possibleEffectRef := range mount.Effects.Array {
						if possibleEffectRef.Ref == -2 {
							continue
						}
						possibleEffect := mountLookup.Ref[int64(possibleEffectRef.Ref)]
						mappedPossibleEffects = append(mappedPossibleEffects, possibleEffect)
					}
					mergedMount := mount.Merge(mappedPossibleEffects)
					mounts[mount.Id] = mergedMount
				}
				mountsChan <- mounts
				return
			}
		}
	}()
	/*go func() {
		ParseRawDataPartUnity("areas.json", areasChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("spell_types.json", spellTypesChan, dir)
	}()
	go func() {
		ParseRawDataPartUnity("spells.json", spellsChan, dir)
		}()*/
	go func() {
		ParseRawDataPartUnity("recipes.json", itemRecipesChang, dir)
	}()
	go func() {
		possibleEffectInstance := "EffectInstanceDice"
		ParseRawDataPartUnityMulti("items.json", itemRawChan, dir, "Items", &possibleEffectInstance)
		for {
			select {
			case itemLookup := <-itemRawChan:
				items := make(map[int]JSONGameItemUnity)
				for _, item := range itemLookup.AnkamaId {
					mappedPossibleEffects := make([]JSONGameItemPossibleEffectUnity, 0)
					for _, possibleEffectRef := range item.PossibleEffects.Array {
						if possibleEffectRef.Ref == -2 {
							continue
						}
						possibleEffect := itemLookup.Ref[int64(possibleEffectRef.Ref)]
						mappedPossibleEffects = append(mappedPossibleEffects, possibleEffect)
					}
					mergedItem := item.Merge(mappedPossibleEffects)
					items[item.Id] = mergedItem
				}
				itemChan <- items
				return
			}
		}
	}()
	go func() {
		ParseRawDataPartUnity("item_types.json", itemTypeChan, dir)
	}()
	go func() {
		possibleEffectInstance := "EffectInstanceDice"
		ParseRawDataPartUnityMulti("item_sets.json", itemSetsRawChan, dir, "ItemSets", &possibleEffectInstance)
		for {
			select {
			case setLookup := <-itemSetsRawChan:
				sets := make(map[int]JSONGameSetUnity)
				for _, set := range setLookup.AnkamaId {
					mappedPossibleEffects := make([][]JSONGameItemPossibleEffectUnity, 0)
					for _, possibleEffectsRef := range set.Effects.Array {
						mappedPossibleEffectsInner := make([]JSONGameItemPossibleEffectUnity, 0)
						for _, possibleEffectRef := range possibleEffectsRef.Values.Array {
							if possibleEffectRef.Ref == -2 {
								continue
							}
							possibleEffect := setLookup.Ref[int64(possibleEffectRef.Ref)]
							mappedPossibleEffectsInner = append(mappedPossibleEffectsInner, possibleEffect)
						}
						mappedPossibleEffects = append(mappedPossibleEffects, mappedPossibleEffectsInner)
					}
					mergedSet := set.Merge(mappedPossibleEffects)
					sets[set.Id] = mergedSet
				}
				itemSetsChan <- sets
				return
			}
		}
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

	/*data.spells = <-spellsChan
	close(spellsChan)

	data.spellTypes = <-spellTypesChan
	close(spellTypesChan)

	data.areas = <-areasChan
	close(areasChan)*/

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
