package main

/*
#include <stdlib.h>
#include "libretro.h"
#include "shim.h"
*/
import "C"

import (
	"slices"
	"strings"
	"unsafe"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"
)

const (
	// defaultModel is the machine set up for playing games, see
	// configs/gaming.cfg
	defaultModel = "gaming"

	// unsupportedModel is left out of the core options, its reason to exist is
	// the Videx Ultraterm video that this core does not render
	unsupportedModel = "ultraterm"
)

// The strings handed to the frontend have to outlive the call, so they are
// allocated once and never freed. What the core is, its name and the extensions
// it takes, is answered in C by shim.c.
var (
	cKeyModel  = C.CString("izapple2_model")
	cKeyScreen = C.CString("izapple2_screen")
)

type options struct {
	model      string
	screenMode int
}

// readOptions returns the settings currently chosen in the frontend
func readOptions() options {
	return options{
		model:      readVariable(cKeyModel, defaultModel),
		screenMode: screenModeByName(readVariable(cKeyScreen, "color")),
	}
}

// screenModeByName is the monitor to emulate. Both are the modes without scan
// lines, the frontend draws those with a shader if it wants them.
func screenModeByName(name string) int {
	if name == "green" {
		return screen.ScreenModeGreen
	}
	return screen.ScreenModeColor
}

/*
videoOverrides takes out the cards that generate the video this core does not
render. The VidHD card, that adds super hi-res, is always in slot 2. The Videx
80 column card is in slot 3 of the two models that carry it.
*/
func videoOverrides(model string) map[string]string {
	overrides := map[string]string{"s2": "empty"}
	switch model {
	case "2plus", "basis108":
		overrides["s3"] = "empty"
	}
	return overrides
}

// buildOverrides is everything the core imposes on the model chosen
func buildOverrides(model string) map[string]string {
	overrides := videoOverrides(model)
	for key, value := range saveOverrides() {
		overrides[key] = value
	}
	return overrides
}

func readVariable(key *C.char, fallback string) string {
	variable := C.struct_retro_variable{key: key}
	ok := C.shim_environment(C.RETRO_ENVIRONMENT_GET_VARIABLE, unsafe.Pointer(&variable))
	if !ok || variable.value == nil {
		return fallback
	}
	value := C.GoString(variable.value)
	if value == "" {
		return fallback
	}
	return value
}

// setVariables declares the core options. The frontend keeps the array, so it
// is built in C memory.
func setVariables() {
	variables := []struct{ key, value string }{
		{"izapple2_model", "Machine model; " + modelChoices()},
		{"izapple2_screen", "Monitor; color|green"},
	}

	// One extra entry, the array is terminated by a pair of null pointers
	count := len(variables) + 1
	size := C.size_t(count) * C.size_t(unsafe.Sizeof(C.struct_retro_variable{}))
	array := (*C.struct_retro_variable)(C.malloc(size))
	entries := unsafe.Slice(array, count)

	for i, variable := range variables {
		entries[i].key = C.CString(variable.key)
		entries[i].value = C.CString(variable.value)
	}
	entries[count-1].key = nil
	entries[count-1].value = nil

	C.shim_environment(C.RETRO_ENVIRONMENT_SET_VARIABLES, unsafe.Pointer(array))
}

// modelChoices lists the models for the option, the default one first
func modelChoices() string {
	models, err := izapple2.AvailableModels()
	if err != nil {
		logf("could not list the models: %v", err)
		return defaultModel
	}

	choices := []string{defaultModel}
	for _, model := range models {
		if model != defaultModel && model != unsupportedModel {
			choices = append(choices, model)
		}
	}
	slices.Sort(choices[1:])

	return strings.Join(choices, "|")
}
