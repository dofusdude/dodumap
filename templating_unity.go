package dodumap

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

func DeleteDamageFormatterUnity(input string) string {
	input, regex := PrepareAndCreateRangeRegexUnity(input, false)
	if strings.Contains(input, "+#1{{~1~2 to }} level #2") {
		return "level"
	}

	input = strings.ReplaceAll(input, "#1{{~1~2 -}}#2", "#1{{~1~2 - }}#2") // bug from ankama
	input = regex.ReplaceAllString(input, "")

	input = strings.ReplaceAll(input, "{{~1~2 to }}", "")
	input = DeleteReplacer(input)
	input = strings.ReplaceAll(input, "  ", " ")

	input = strings.TrimSpace(input)
	return input
}

func SingularPluralFormatterUnity(input string, amount int, lang string) string {
	str := input
	str = strings.ReplaceAll(str, "{{~s}}", "") // avoid only s without what to append
	str = strings.ReplaceAll(str, "{{~p}}", "") // same

	// delete unknown z
	unknownZRegex := regexp.MustCompile("{{~z[^}]*}}")
	str = unknownZRegex.ReplaceAllString(str, "")

	var indicator rune

	if amount > 1 {
		indicator = 'p'
	} else {
		indicator = 's'
	}

	indicators := []rune{'s', 'p'}
	var regexps []*regexp.Regexp
	for _, indicatorIt := range indicators {
		regex := fmt.Sprintf("{{~%c([^}]*)}}", indicatorIt) // capturing with everything inside ()
		regexExtract := regexp.MustCompile(regex)
		regexps = append(regexps, regexExtract)

		//	if lang == "es" || lang == "pt" {
		if indicatorIt != indicator {
			continue
		}

		if lang == "de" {
			regexExtract.ReplaceAllString(str, "") // german templating is just a mess so remove it
			continue
		}

		extractedEntries := regexExtract.FindAllStringSubmatch(str, -1)
		for _, extracted := range extractedEntries {
			str = strings.ReplaceAll(str, extracted[0], extracted[1])
		}
	}

	for _, regexIt := range regexps {
		str = regexIt.ReplaceAllString(str, "")
	}

	return str
}

func PrepareAndCreateRangeRegexUnity(input string, extract bool) (string, *regexp.Regexp) {
	var regexStr string
	combiningWords := "(und|et|and|bis|to|a|à|-|auf)"
	if extract {
		regexStr = fmt.Sprintf("{{~1~2 (%s [-,+]?)}}", combiningWords)
	} else {
		regexStr = fmt.Sprintf("[-,+]?#1{{~1~2 %s [-,+]?}}#2", combiningWords)
	}

	concatRegex := regexp.MustCompile(regexStr)

	return PrepareTextForRegex(input), concatRegex
}

// NOTE: When changing here, also needs changes in ConditionWithOperatorUnity lower function body because some special cases of replacments
func ElementFromCodeUnity(codeUndef string) []int {
	code := strings.ToLower(codeUndef)

	switch code {
	case "cs":
		return []int{501945} // "Strength"
	case "ci":
		return []int{501944} // "Intelligence"
	case "cv":
		return []int{501947} // "Vitality"
	case "ca":
		return []int{501941} // "Agility"
	case "cc":
		return []int{501942} // "Chance"
	case "cw":
		return []int{501946} // "Wisdom"
	case "pk":
		if codeUndef == "PK" {
			return []int{862811} //  "Kamas"
		} else {
			return []int{1092135} // "Set-Bonus"
		}
	case "pl":
		return []int{1096588} // "Be level {0} or higher"
	case "cm":
		return []int{501874, 1143881} // "Movement Points (MP)" TODO one of them vanishes
	case "cp":
		return []int{501948, 1143880} // "Action Points (AP)" TODO one of them vanished
	case "po":
		return []int{1092470} // "Different area to: {0}"
	case "pf":
		return []int{1095105} // "{0} not equipped"
	//case "": // Ps=1
	//	return 644230 // Ausgerüstetes %1-Reittier
	case "pa":
		return []int{1093891} // "Alignment level"
	//case "":
	//	return 637203 // Kein ausgerüstetes %1-Reittier haben
	case "of":
		return []int{1094822} // "Have a {0} mount equipped"
	case "pz":
		return []int{1093970} // "Be subscribed"
	}

	return nil
}

func ConditionWithOperatorUnity(input string, operator string, langs *map[string]LangDictUnity, out *MappedMultilangCondition, data *JSONGameDataUnity) bool {
	partSplit := strings.Split(input, operator)
	rawElement := ElementFromCodeUnity(partSplit[0])
	if rawElement == nil {
		return false
	}
	actualElement := -1
	out.Element = partSplit[0]
	out.Value, _ = strconv.Atoi(partSplit[1])
	for _, lang := range LanguagesUnity {
		langStr, ok := (*langs)[lang].Texts[rawElement[0]]
		// TODO remove this mess when 3.2 is out for some time, else ambiguity. And we have hardcoded size 2 here, which is also not really stable.
		if !ok {
			langStr, ok = (*langs)[lang].Texts[rawElement[1]]
			if !ok {
				log.Fatalf("Could not find condition translation for %s", partSplit[1])
			} else {
				actualElement = rawElement[1]
			}
		} else {
			actualElement = rawElement[0]
		}

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

		switch actualElement {
		case 1096588: // %1 replace
			intVal, _ := strconv.Atoi(partSplit[1])
			langStr = strings.ReplaceAll(langStr, "%1", fmt.Sprint(intVal+1))
		case 1092470: // anderes gebiet als %1
			langStr = strings.ReplaceAll(langStr, "%1", (*langs)[lang].Texts[data.areas[out.Value].NameId])
		case 1094822: // reittier %1
		case 1095105:
			langStr = strings.ReplaceAll(langStr, "%1", (*langs)[lang].Texts[data.Mounts[out.Value].NameId])
		}

		out.Templated[lang] = langStr
	}
	out.Operator = operator

	buggyConditions := []int{181}

	return !slices.Contains(buggyConditions, out.ElementId)
}

func NumSpellFormatterUnity(input string, lang string, gameData *JSONGameDataUnity, langs *map[string]LangDictUnity, diceNum *int, diceSide *int, value *int, effectNameId int, numIsSpell bool, useDice bool, frNumSigned *int, frSideSigned *int) (string, int) {
	diceNumIsSpellId := *diceNum > 8000 || numIsSpell
	diceSideIsSpellId := *diceSide > 8000
	valueIsSpellId := *value > 8000

	onlyNoMinMax := 0

	// when + xp
	if !useDice && *diceNum == 0 && *value == 0 && *diceSide != 0 {
		*value = *diceSide
		*diceSide = 0
	}

	delValue := false

	input, concatRegex := PrepareAndCreateRangeRegexUnity(input, true)
	var numSigned bool
	var sideSigned bool
	var ptSideSigned bool
	_, ptSideSigned = ParseSignessUnity(input)
	if *frNumSigned != 2 || *frSideSigned != 2 { // 2 is unset, 0 is false, 1 is true
		numSigned = *frNumSigned == 1
		sideSigned = *frSideSigned == 1
	} else {
		if lang == "fr" {
			numSigned, sideSigned = ParseSignessUnity(input)
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

func ParseSignessUnity(input string) (bool, bool) {
	numSigness := false
	sideSigness := false

	regexNum := regexp.MustCompile("(([+,-])?#1)")
	entriesNum := regexNum.FindAllStringSubmatch(input, -1)
	for _, extracted := range entriesNum {
		for _, entry := range extracted {
			if entry == "-" {
				numSigness = true
			}
		}
	}

	regexSide := regexp.MustCompile("([+,-])?(}})?#2")
	entriesSide := regexSide.FindAllStringSubmatch(input, -1)
	for _, extracted := range entriesSide {
		for _, entry := range extracted {
			if entry == "-" {
				sideSigness = true
			}
		}
	}

	return numSigness, sideSigness
}
