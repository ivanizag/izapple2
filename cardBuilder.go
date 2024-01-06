package izapple2

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
)

type paramSpec struct {
	name         string
	description  string
	defaultValue string
}

type cardBuilder struct {
	name          string
	description   string
	defaultParams *[]paramSpec
	requiresIIe   bool
	buildFunc     func(params map[string]string) (Card, error)
}

const noCardName = "empty"

var cardFactory map[string]*cardBuilder

func getCardFactory() map[string]*cardBuilder {
	if cardFactory != nil {
		return cardFactory
	}
	cardFactory = make(map[string]*cardBuilder)
	cardFactory["brainboard"] = newCardBrainBoardIIBuilder()
	cardFactory["diskii"] = newCardDisk2Builder()
	cardFactory["diskiiseq"] = newCardDisk2SequencerBuilder()
	cardFactory["fastchip"] = newCardFastChipBuilder()
	cardFactory["fujinet"] = newCardSmartPortFujinetBuilder()
	cardFactory["inout"] = newCardInOutBuilder()
	cardFactory["language"] = newCardLanguageBuilder()
	cardFactory["softswitchlogger"] = newCardLoggerBuilder()
	cardFactory["memexp"] = newCardMemoryExpansionBuilder()
	cardFactory["mouse"] = newCardMouseBuilder()
	cardFactory["parallel"] = newCardParallelPrinterBuilder()
	cardFactory["saturn"] = newCardSaturnBuilder()
	cardFactory["smartport"] = newCardSmartPortStorageBuilder()
	cardFactory["swyftcard"] = newCardSwyftBuilder()
	cardFactory["thunderclock"] = newCardThunderClockPlusBuilder()
	cardFactory["videx"] = newCardVidexBuilder()
	cardFactory["vidhd"] = newCardVidHDBuilder()
	return cardFactory
}

func availableCards() []string {
	return maps.Keys(getCardFactory())
}

func setupCard(a *Apple2, slot int, paramString string) (Card, error) {
	paramsArgs := splitConfigurationString(paramString, ',')

	cardName := paramsArgs[0]
	if cardName == "" || cardName == noCardName {
		return nil, nil
	}

	builder, ok := getCardFactory()[cardName]
	if !ok {
		return nil, fmt.Errorf("unknown card %s", cardName)
	}

	if builder.requiresIIe && !a.isApple2e {
		return nil, fmt.Errorf("card %s requires an Apple IIe", builder.name)
	}

	finalParams := make(map[string]string)
	if builder.defaultParams != nil {
		for _, defaultParam := range *builder.defaultParams {
			finalParams[defaultParam.name] = defaultParam.defaultValue
		}
	}

	for i := 1; i < len(paramsArgs); i++ {
		paramArgSides := splitConfigurationString(paramsArgs[i], '=')

		if _, ok := finalParams[paramArgSides[0]]; !ok {
			return nil, fmt.Errorf("unknown parameter %s", paramArgSides[0])
		}
		if len(paramArgSides) > 2 {
			return nil, fmt.Errorf("invalid parameter value for %s", paramArgSides[0])
		}
		if len(paramArgSides) == 1 {
			finalParams[paramArgSides[0]] = "true"
		} else {
			finalParams[paramArgSides[0]] = paramArgSides[1]
		}
	}

	card, err := builder.buildFunc(finalParams)
	if err != nil {
		return nil, err
	}

	cardBase, ok := card.(*cardBase)
	if err == nil && ok {
		cardBase.name = builder.name
	}

	card.assign(a, slot)
	a.cards[slot] = card

	return card, err
}

func paramsGetBool(params map[string]string, name string) bool {
	value, ok := params[name]
	if !ok {
		value = "false"
	}
	return value == "true"
}

func paramsGetString(params map[string]string, name string) string {
	value, ok := params[name]
	if !ok {
		value = ""
	}
	return value
}

func paramsGetPath(params map[string]string, name string) string {
	value := paramsGetString(params, name)
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = value[1 : len(value)-1]
	}
	return value
}

func paramsGetInt(params map[string]string, name string) (int, error) {
	value, ok := params[name]
	if !ok {
		value = "0"
	}
	return strconv.Atoi(value)
}

func splitConfigurationString(s string, separator rune) []string {
	// Split by comma, but not inside quotes
	var result []string
	var current string
	inQuote := false
	for _, c := range s {
		if c == '"' {
			inQuote = !inQuote
		}
		if c == separator && !inQuote {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
