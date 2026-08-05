package izapple2

import (
	"fmt"
	"strconv"
	"strings"

	"maps"
	"slices"
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
	hide          bool
	buildFunc     func(params map[string]string) (Card, error)
}

const noCardName = "empty"

const saveDirParamName = "savedir"

/*
saveDirParamSpec is declared by the cards that write to their disk images, so
that the changes can be kept apart and the images left alone.

Left empty, the card uses the save directory of the machine, the saveDir
command line option. Set to 'none' it writes back to the image even when the
machine has one.
*/
var saveDirParamSpec = paramSpec{
	saveDirParamName,
	"directory to keep what is written to the disks of this card, leaving the images unmodified. Defaults to the saveDir of the machine, 'none' to write back to the images",
	"",
}

// paramsGetSaveDir returns the directory a card keeps the changes to its disks
// in, empty when they go back to the images
func paramsGetSaveDir(params map[string]string) string {
	value := paramsGetPath(params, saveDirParamName)
	if value == "none" {
		return ""
	}
	return value
}

var commonParams = []paramSpec{
	{"trace", "Enable debug messages", "false"},
	{"tracess", "Trace softswitches", "false"},
	{"panicss", "Panic on unimplemented softswitches", "false"},
	{"tracemem", "Trace slot addressing accesses", "false"},
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
	cardFactory["mockingboard"] = newCardMockingboardBuilder()
	cardFactory["mouse"] = newCardMouseBuilder()
	cardFactory["multirom"] = newMultiRomCardBuilder()
	cardFactory["parallel"] = newCardParallelPrinterBuilder()
	cardFactory["prodosblock"] = newCardProDOSBlockStorageBuilder()
	cardFactory["prodosromdrive"] = newCardProDOSRomDriveBuilder()
	cardFactory["prodosromcard3"] = newCardProDOSRomCard3Builder()
	// cardFactory["prodosnvramdrive"] = newCardProDOSNVRAMDriveBuilder()
	cardFactory["profile"] = newCardProfileBuilder()
	cardFactory["saturn"] = newCardSaturnBuilder()
	cardFactory["smartport"] = newCardSmartPortStorageBuilder()
	cardFactory["swyftcard"] = newCardSwyftBuilder()
	cardFactory["thunderclock"] = newCardThunderClockPlusBuilder()
	cardFactory["videx"] = newCardVidexVideotermBuilder()
	cardFactory["videxultraterm"] = newCardVidexUltratermBuilder()
	cardFactory["vidhd"] = newCardVidHDBuilder()
	cardFactory["z80softcard"] = newCardZ80SoftCardBuilder()
	return cardFactory
}

func availableCards() []string {
	names := slices.Collect(maps.Keys(getCardFactory()))
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

// setupCard builds the card of a slot. saveDir is the one of the machine, used
// by the cards that write to their disk images and do not name one themselves.
func setupCard(a *Apple2, slot int, paramString string, saveDir string) (Card, error) {
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

	// A card that writes to its disks and does not name a save directory of its
	// own uses the one of the machine
	if value, ok := finalParams[saveDirParamName]; ok && value == "" {
		finalParams[saveDirParamName] = saveDir
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

	card.configure(
		builder.name,
		paramsGetBool(finalParams, "trace"),
		paramsGetBool(finalParams, "tracemem"),
	)

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
	for i := range 8 {
		result[i+1] = value[i] == '1'

	}
	return result, nil
}

func splitConfigurationString(s string, separator rune) []string {
	// Split by separator, but not inside quotes
	var result []string
	var current strings.Builder
	inQuote := false
	for _, c := range s {
		if c == '"' {
			inQuote = !inQuote
		}
		if c == separator && !inQuote {
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteRune(c)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}
