package dodumap

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

var LanguagesUnity = []string{"fr", "en", "de", "es", "pt"}

func MapItemsUnity(data *JSONGameDataUnity, langs *map[string]LangDictUnity) []MappedMultilangItemUnity {
	var filteredItems []JSONGameItemUnity

	for key, value := range data.Items {
		if langs == nil || data == nil {
			log.Fatal("langs is nil")
		}
		frName := (*langs)["fr"].Texts[value.NameId]
		itemType := data.ItemTypes[value.TypeId]
		category := itemType.CategoryId
		deTypeName := (*langs)["de"].Texts[data.ItemTypes[value.TypeId].NameId]
		if frName == "" || category == 4 || deTypeName == "Hauptquesten" {
			continue // skip unnamed and hidden items
		}
		filteredItems = append(filteredItems, data.Items[key])
	}

	mappedItems := make([]MappedMultilangItemUnity, len(filteredItems))
	for idx, item := range filteredItems {
		mappedItems[idx].AnkamaId = item.Id
		mappedItems[idx].Level = item.Level
		mappedItems[idx].Pods = item.Pods
		mappedItems[idx].Image = fmt.Sprintf("https://static.ankama.com/dofus/www/game/items/200/%d.png", item.IconId)
		mappedItems[idx].Name = make(map[string]string, len(LanguagesUnity))
		mappedItems[idx].Description = make(map[string]string, len(LanguagesUnity))
		mappedItems[idx].Type.Name = make(map[string]string, len(LanguagesUnity))
		mappedItems[idx].IconId = item.IconId

		for _, lang := range LanguagesUnity {
			mappedItems[idx].Name[lang] = (*langs)[lang].Texts[item.NameId]
			mappedItems[idx].Description[lang] = (*langs)[lang].Texts[item.DescriptionId]
			mappedItems[idx].Type.Name[lang] = (*langs)[lang].Texts[data.ItemTypes[item.TypeId].NameId]
		}

		mappedItems[idx].Type.Id = item.TypeId
		mappedItems[idx].Type.SuperTypeId = data.ItemTypes[item.TypeId].SuperTypeId
		mappedItems[idx].Type.CategoryId = data.ItemTypes[item.TypeId].CategoryId

		searchTypeEn := mappedItems[idx].Type.Name["en"]
		key, foundKey := PersistedTypes.Entries.GetKey(searchTypeEn)
		if foundKey {
			mappedItems[idx].Type.ItemTypeId = key.(int)
		} else {
			mappedItems[idx].Type.ItemTypeId = PersistedTypes.NextId
			PersistedTypes.Entries.Put(PersistedTypes.NextId, searchTypeEn)
			PersistedTypes.NextId++
		}

		if len(item.RecipeIds.Array) > 0 {
			mappedItems[idx].UsedInRecipes = item.RecipeIds.Array
		}
		effectsArr := make([][]*JSONGameItemPossibleEffectUnity, 1)
		effectsArr[0] = item.PossibleEffects
		allEffectResult := ParseEffectsUnity(data, effectsArr, langs)
		if len(allEffectResult) > 0 {
			for _, effect := range allEffectResult[0] {
				if effect == nil {
					continue
				}
				mappedItems[idx].Effects = append(mappedItems[idx].Effects, *effect)
			}
		}
		mappedItems[idx].Range = item.Range
		mappedItems[idx].MinRange = item.MinRange
		mappedItems[idx].CriticalHitProbability = item.CriticalHitProbability
		mappedItems[idx].CriticalHitBonus = item.CriticalHitBonus
		mappedItems[idx].ApCost = item.ApCost
		mappedItems[idx].MaxCastPerTurn = item.MaxCastPerTurn
		if len(item.DropMonsterIds.Array) > 0 {
			mappedItems[idx].DropMonsterIds = item.DropMonsterIds.Array
		}
		mappedItems[idx].HasParentSet = item.ItemSetId != -1
		if mappedItems[idx].HasParentSet {
			mappedItems[idx].ParentSet.Id = item.ItemSetId
			mappedItems[idx].ParentSet.Name = make(map[string]string, len(LanguagesUnity))
			for _, lang := range LanguagesUnity {
				mappedItems[idx].ParentSet.Name[lang] = (*langs)[lang].Texts[data.Sets[item.ItemSetId].NameId]
			}
		}

		if len(item.Criteria) != 0 && mappedItems[idx].Type.Name["de"] != "Verwendbarer Temporis-Gegenstand" { // TODO Temporis got some weird conditions
			mappedItems[idx].Conditions = ParseConditionUnity(item.Criteria, langs, data)
		}
	}

	return mappedItems
}

func MapMountsUnity(data *JSONGameDataUnity, langs *map[string]LangDictUnity) []MappedMultilangMount {
	var mappedMounts []MappedMultilangMount
	for _, mount := range data.Mounts {
		var mappedMount MappedMultilangMount
		mappedMount.AnkamaId = mount.Id
		mappedMount.FamilyId = mount.FamilyId
		mappedMount.Name = make(map[string]string)
		mappedMount.FamilyName = make(map[string]string)

		for _, lang := range LanguagesUnity {
			mappedMount.Name[lang] = (*langs)[lang].Texts[mount.NameId]
			mappedMount.FamilyName[lang] = (*langs)[lang].Texts[data.MountFamilys[mount.FamilyId].NameId]
		}

		effectsArr := make([][]*JSONGameItemPossibleEffectUnity, 1)
		effectsArr[0] = mount.Effects
		allEffectResult := ParseEffectsUnity(data, effectsArr, langs)
		if len(allEffectResult) > 0 {
			for _, effect := range allEffectResult[0] {
				if effect == nil {
					continue
				}
				mappedMount.Effects = append(mappedMount.Effects, *effect)
			}
		}

		mappedMounts = append(mappedMounts, mappedMount)
	}

	if len(mappedMounts) == 0 {
		return nil
	}

	return mappedMounts
}

func MapAlmanaxUnity(data *JSONGameDataUnity, langs *map[string]LangDictUnity) []MappedMultilangNPCAlmanax {
	var mappedAlmanax []MappedMultilangNPCAlmanax

	for _, almCat := range data.questCategories[31].QuestIds.Array {
		quest := data.quests[almCat]
		if (*langs)["en"].Texts[quest.NameId][:8] != "Offering" {
			continue
		}

		step := data.questSteps[quest.StepIds.Array[0]]
		objective := data.questObjectives[step.ObjectiveIds.Array[0]].Parameters
		item := data.Items[objective.Parameter1]
		itemQuantity := objective.Parameter2
		stepRewards := data.questStepRewards[step.RewardsIds.Array[0]]

		kamasRatio := stepRewards.KamasRatio
		maxLevel := stepRewards.LevelMax

		var kamasScaleWithPlayerLevel bool
		if stepRewards.KamasScaleWithPlayerLevel == 0 {
			kamasScaleWithPlayerLevel = false
		} else {
			kamasScaleWithPlayerLevel = true
		}

		duration := step.Duration
		optimalLevel := step.OptimalLevel

		rewardKamas := questKamasReward(maxLevel, optimalLevel, kamasRatio, duration, kamasScaleWithPlayerLevel)
		questObjectiveNpc := data.questObjectives[step.ObjectiveIds.Array[2]].Parameters.Parameter0

		var currAlm JSONGameAlamanaxCalendarUnity
		found := false
		for _, almCal := range data.almanaxCalendars {
			if almCal.NpcId == questObjectiveNpc {
				currAlm = almCal
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("Could not find almanax calendar for NPC %d", questObjectiveNpc)
		}

		var mappedNPCAlmanax MappedMultilangNPCAlmanax
		mappedNPCAlmanax.OfferingReceiver = (*langs)["en"].Texts[quest.NameId][13:] // remove "Offering to ". The name is the same in all languages.
		itemNames := make(map[string]string)
		mappedNPCAlmanax.Bonus = make(map[string]string)
		mappedNPCAlmanax.BonusType = make(map[string]string)
		for _, lang := range LanguagesUnity {
			itemNames[lang] = (*langs)[lang].Texts[item.NameId]
			numId, err := strconv.Atoi(currAlm.DescId)
			if err != nil {
				log.Fatal("Could not convert", "err", err, "val", currAlm.DescId)
			}
			mappedNPCAlmanax.Bonus[lang] = (*langs)[lang].Texts[numId]
			mappedNPCAlmanax.BonusType[lang] = (*langs)[lang].Texts[currAlm.NameId]

			mappedNPCAlmanax.Bonus[lang] = strings.ReplaceAll(mappedNPCAlmanax.Bonus[lang], "<b>", "")
			mappedNPCAlmanax.Bonus[lang] = strings.ReplaceAll(mappedNPCAlmanax.Bonus[lang], "</b>", "")
		}
		mappedNPCAlmanax.Offering.ItemId = item.Id
		mappedNPCAlmanax.Offering.ItemName = itemNames
		mappedNPCAlmanax.Offering.Quantity = itemQuantity
		mappedNPCAlmanax.RewardKamas = rewardKamas

		ImgBaseUrl := "https://api.dofusdu.de/dofus2/img/item/" + strconv.Itoa(item.IconId)
		mappedNPCAlmanax.Offering.ImageUrls.HD = ImgBaseUrl + "-800.png"
		mappedNPCAlmanax.Offering.ImageUrls.HQ = ImgBaseUrl + "-400.png"
		mappedNPCAlmanax.Offering.ImageUrls.SD = ImgBaseUrl + "-200.png"
		mappedNPCAlmanax.Offering.ImageUrls.Icon = ImgBaseUrl + ".png"

		mappedAlmanax = append(mappedAlmanax, mappedNPCAlmanax)
	}

	if len(mappedAlmanax) == 0 {
		return nil
	}

	return mappedAlmanax
}

func MapRecipesUnity(data *JSONGameDataUnity) []MappedMultilangRecipe {
	var mappedRecipes []MappedMultilangRecipe

	for _, recipe := range data.Recipes {
		ingredientCount := len(recipe.IngredientIds.Array)
		var mappedRecipe MappedMultilangRecipe
		mappedRecipe.ResultId = recipe.Id
		mappedRecipe.Entries = make([]MappedMultilangRecipeEntry, ingredientCount)
		for i := 0; i < ingredientCount; i++ {
			var recipeEntry MappedMultilangRecipeEntry
			recipeEntry.ItemId = recipe.IngredientIds.Array[i]
			recipeEntry.Quantity = recipe.Quantities.Array[i]
			mappedRecipe.Entries[i] = recipeEntry
		}
		mappedRecipes = append(mappedRecipes, mappedRecipe)
	}

	if len(mappedRecipes) == 0 {
		return nil
	}

	return mappedRecipes
}

func MapSetsUnity(data *JSONGameDataUnity, langs *map[string]LangDictUnity) []MappedMultilangSetUnity {
	var mappedSets []MappedMultilangSetUnity
	for _, set := range data.Sets {
		var mappedSet MappedMultilangSetUnity
		mappedSet.AnkamaId = set.Id
		mappedSet.ItemIds = set.ItemIds
		parseEffects := ParseEffectsUnity(data, set.Effects, langs)

		parseCombi := ParseItemComboUnity(parseEffects)
		if len(parseCombi) > 0 {
			mappedSet.Effects = parseCombi
		}

		allItemsCosmetic := len(set.ItemIds) > 0

		highestLevel := 0
		for _, item := range set.ItemIds {
			if data.Items[item].Level > highestLevel {
				highestLevel = data.Items[item].Level
			}

			// 5 == "cosmetic"
			if data.ItemTypes[data.Items[item].TypeId].CategoryId != 5 {
				allItemsCosmetic = false
			}
		}

		mappedSet.IsCosmetic = allItemsCosmetic
		mappedSet.Level = highestLevel

		mappedSet.Name = make(map[string]string)
		for _, lang := range LanguagesUnity {
			mappedSet.Name[lang] = (*langs)[lang].Texts[set.NameId]
		}

		mappedSets = append(mappedSets, mappedSet)
	}

	if len(mappedSets) == 0 {
		return nil
	}

	return mappedSets
}
