package dodumap

import (
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func ConditionWithOperatorUnity(input string, operator string, langs *map[string]LangDictUnity, out *MappedMultilangCondition, data *JSONGameDataUnity) bool {
	partSplit := strings.Split(input, operator)
	rawElement := ElementFromCode(partSplit[0])
	if rawElement == -1 {
		return false
	}
	out.Element = partSplit[0]
	out.Value, _ = strconv.Atoi(partSplit[1])
	for _, lang := range Languages {
		langStr := (*langs)[lang].Texts[rawElement]

		if lang == "en" {
			if langStr == "()" {
				return false
			}

			keySanitized := DeleteReplacer(langStr)

			if PersistedElements.Entries == nil {
				log.Fatal("Elements Entries is nil")
			}

			key, foundKey := PersistedElements.Entries.GetKey(keySanitized)
			if foundKey {
				out.ElementId = key.(int)
			} else {
				PersistedElements.Entries.Put(PersistedElements.NextId, keySanitized)
				PersistedElements.NextId++
			}
		}

		switch rawElement {
		case 837224: // %1 replace
			intVal, _ := strconv.Atoi(partSplit[1])
			langStr = strings.ReplaceAll(langStr, "%1", fmt.Sprint(intVal+1))
		case 335357: // anderes gebiet als %1
			langStr = strings.ReplaceAll(langStr, "%1", (*langs)[lang].Texts[data.areas[out.Value].NameId])
		case 637212: // reittier %1
		case 644231:
			langStr = strings.ReplaceAll(langStr, "%1", (*langs)[lang].Texts[data.Mounts[out.Value].NameId])
		}

		out.Templated[lang] = langStr
	}
	out.Operator = operator

	buggyConditions := []int{181}

	return !slices.Contains(buggyConditions, out.ElementId)
}

func NumSpellFormatterUnity(input string, lang string, gameData *JSONGameDataUnity, langs *map[string]LangDictUnity, diceNum *int, diceSide *int, value *int, effectNameId int, numIsSpell bool, useDice bool, frNumSigned *int, frSideSigned *int) (string, int) {
	diceNumIsSpellId := *diceNum > 12000 || numIsSpell
	diceSideIsSpellId := *diceSide > 12000
	valueIsSpellId := *value > 12000

	onlyNoMinMax := 0

	// when + xp
	if !useDice && *diceNum == 0 && *value == 0 && *diceSide != 0 {
		*value = *diceSide
		*diceSide = 0
	}

	delValue := false

	input, concatRegex := PrepareAndCreateRangeRegex(input, true)
	var numSigned bool
	var sideSigned bool
	var ptSideSigned bool
	_, ptSideSigned = ParseSigness(input)
	if *frNumSigned != 2 || *frSideSigned != 2 { // 2 is unset, 0 is false, 1 is true
		numSigned = *frNumSigned == 1
		sideSigned = *frSideSigned == 1
	} else {
		if lang == "fr" {
			numSigned, sideSigned = ParseSigness(input)
			if numSigned {
				*frNumSigned = 1
			} else {
				*frNumSigned = 0
			}
			if sideSigned {
				*frSideSigned = 1
			} else {
				*frSideSigned = 0
			}
		} else {
			log.Fatalf("frNumSigned and frSideSigned must be set for %s", lang)
		}
	}
	concatEntries := concatRegex.FindAllStringSubmatch(input, -1)

	if *diceSide == 0 { // only replace #1 with dice_num
		for _, extracted := range concatEntries {
			input = strings.ReplaceAll(input, extracted[0], "")
		}
	} else {
		for _, extracted := range concatEntries {
			input = strings.ReplaceAll(input, extracted[0], fmt.Sprintf(" %s", extracted[1]))
		}
	}

	num1Regex := regexp.MustCompile("([-,+]?)#1")
	num1Entries := num1Regex.FindAllStringSubmatch(input, -1)
	for _, extracted := range num1Entries {
		var diceNumStr string
		if diceNumIsSpellId {
			diceNumStr = (*langs)[lang].Texts[gameData.spells[*diceNum].NameId]
		} else {
			diceNumStr = fmt.Sprint(*diceNum)
		}
		input = strings.ReplaceAll(input, extracted[0], fmt.Sprintf("%s%s", extracted[1], diceNumStr))
	}

	if *diceSide == 0 {
		input = strings.ReplaceAll(input, "#2", "")
	} else {
		var diceSideStr string
		if diceSideIsSpellId {
			diceSideStr = (*langs)[lang].Texts[gameData.spells[*diceSide].NameId]
			//del_dice_side = true
		} else {
			if sideSigned && lang == "pt" && !ptSideSigned {
				diceSideStr = fmt.Sprintf("-%d", *diceSide)
			} else {
				diceSideStr = fmt.Sprint(*diceSide)
			}
		}
		input = strings.ReplaceAll(input, "#2", diceSideStr)
	}

	var valueStr string
	if valueIsSpellId {
		valueStr = (*langs)[lang].Texts[gameData.spells[*value].NameId]
		delValue = true
	} else {
		valueStr = fmt.Sprint(*value)
	}
	if effectNameId == 427090 { // go to <npc> for more info
		return "", -2
	}
	input = strings.ReplaceAll(input, "#3", valueStr)

	if delValue {
		*diceNum = Min(*diceNum, *diceSide)
	}

	if !useDice {
		// avoid min = 0, max > x
		if *diceNum == 0 && *diceSide != 0 {
			*diceNum = *diceSide
			*diceSide = 0
		}
	}

	if *diceNum == 0 && *diceSide == 0 {
		onlyNoMinMax = -2
	}

	if *diceNum != 0 && *diceSide == 0 {
		onlyNoMinMax = -1
	}

	input = strings.TrimSpace(input)

	if numSigned {
		*diceNum *= -1
	}

	if sideSigned {
		*diceSide *= -1
	}

	if *diceNum < 0 && *diceSide < 0 {
		*diceNum, *diceSide = *diceSide, *diceNum

		diceSideFmt := fmt.Sprint((*diceSide) * -1)
		diceNumFmt := fmt.Sprint((*diceNum) * -1)

		input = strings.ReplaceAll(input, diceSideFmt, "-diceSideFmt-")
		input = strings.ReplaceAll(input, diceNumFmt, diceSideFmt)
		input = strings.ReplaceAll(input, "-diceSideFmt-", diceNumFmt)

		if !strings.Contains(input, "-"+diceSideFmt) {
			input = strings.ReplaceAll(input, diceSideFmt, "-"+diceSideFmt)
		}

		if !strings.Contains(input, "-"+diceNumFmt) {
			input = strings.ReplaceAll(input, diceNumFmt, "-"+diceNumFmt)
		}
	}

	return input, onlyNoMinMax
}