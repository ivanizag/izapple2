package izapple2

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
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

var commonParams = []paramSpec{
	{"trace", "Enable debug messages", "false"},
	{"tracess", "Trace softswitches", "false"},
	{"panicss", "Panic on unimplemented softswitches", "false"},
}

var cardFactory map[string]*cardBuilder

func getCardFactory() map[string]*cardBuilder {
	if cardFactory != nil {
		return cardFactory
	}
	cardFactory = make(map[string]*cardBuilder)
	cardFactory["brainboard"] = newCardBrainBoardBuilder()
	cardFactory["brainboard2"] = newCardBrainBoardIIBuilder()
	cardFactory["dan2sd"] = newCardDan2ControllerBuilder()
	cardFactory["diskii"] = newCardDisk2Builder()
	cardFactory["diskiiseq"] = newCardDisk2SequencerBuilder()
	cardFactory["fastchip"] = newCardFastChipBuilder()
	cardFactory["fujinet"] = newCardSmartPortFujinetBuilder()
	cardFactory["inout"] = newCardInOutBuilder()
	cardFactory["language"] = newCardLanguageBuilder()
	cardFactory["softswitchlogger"] = newCardLoggerBuilder()
	cardFactory["memexp"] = newCardMemoryExpansionBuilder()
	cardFactory["mouse"] = newCardMouseBuilder()
	cardFactory["multirom"] = newMultiRomCardBuilder()
	cardFactory["parallel"] = newCardParallelPrinterBuilder()
	cardFactory["prodosromdrive"] = newCardProDOSRomDriveBuilder()
	cardFactory["prodosromcard3"] = newCardProDOSRomCard3Builder()
	// cardFactory["prodosnvramdrive"] = newCardProDOSNVRAMDriveBuilder()
	cardFactory["saturn"] = newCardSaturnBuilder()
	cardFactory["smartport"] = newCardSmartPortStorageBuilder()
	cardFactory["swyftcard"] = newCardSwyftBuilder()
	cardFactory["thunderclock"] = newCardThunderClockPlusBuilder()
	cardFactory["videx"] = newCardVidexBuilder()
	cardFactory["vidhd"] = newCardVidHDBuilder()
	return cardFactory
}

func availableCards() []string {
	names := maps.Keys(getCardFactory())
	slices.Sort(names)
	return names
}

func (cb *cardBuilder) fullDefaultParams() map[string]string {
	finalParams := make(map[string]string)
	for _, commonParam := range commonParams {
		finalParams[commonParam.name] = commonParam.defaultValue
	}
	if cb.defaultParams != nil {
		for _, defaultParam := range *cb.defaultParams {
			finalParams[defaultParam.name] = defaultParam.defaultValue
		}
	}

	return finalParams
}

func setupCard(a *Apple2, slot int, paramString string) (Card, error) {
	actualArgs := splitConfigurationString(paramString, ',')

	cardName := actualArgs[0]
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

	finalParams := builder.fullDefaultParams()
	for i := 1; i < len(actualArgs); i++ {
		actualArgSides := splitConfigurationString(actualArgs[i], '=')
		actualArgName := strings.ToLower(actualArgSides[0])

		if _, ok := finalParams[actualArgName]; !ok {
			return nil, fmt.Errorf("unknown parameter %s", actualArgSides[0])
		}
		if len(actualArgSides) > 2 {
			return nil, fmt.Errorf("invalid parameter value for %s", actualArgSides[0])
		}
		if len(actualArgSides) == 1 {
			finalParams[actualArgName] = "true"
		} else {
			finalParams[actualArgName] = actualArgSides[1]
		}
	}

	card, err := builder.buildFunc(finalParams)
	if err != nil {
		return nil, err
	}

	// Common parameters
	if paramsGetBool(finalParams, "tracess") {
		a.io.traceSlot(slot)
	}

	if paramsGetBool(finalParams, "panicss") {
		a.io.panicNotImplementedSlot(slot)
	}

	debug := paramsGetBool(finalParams, "trace")

	card.setName(builder.name)
	card.setDebug(debug)
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
		return 0, fmt.Errorf("missing parameter %s", name)
	}
	return strconv.Atoi(value)
}

func paramsGetUInt8(params map[string]string, name string) (uint8, error) {
	value, ok := params[name]
	if !ok {
		return 0, fmt.Errorf("missing parameter %s", name)
	}
	result, err := strconv.ParseUint(value, 10, 8)
	return uint8(result), err
}

// Returns a 1 based array of bools
func paramsGetDIPs(params map[string]string, name string, size int) ([]bool, error) {
	value, ok := params[name]
	if !ok {
		return nil, fmt.Errorf("missing parameter %s", name)
	}
	if len(value) != 8 {
		return nil, fmt.Errorf("DIP switches must be 8 characters long")
	}
	result := make([]bool, size+1)
	for i := 0; i < 8; i++ {
		result[i+1] = value[i] == '1'

	}
	return result, nil
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
