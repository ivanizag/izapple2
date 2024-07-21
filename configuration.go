package izapple2

import (
	"embed"
	"flag"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

const configSuffix = ".cfg"
const defaultConfiguration = "2enh"

const (
	confParent = "parent"
	confModel  = "model"
	confName   = "name"
	confBoard  = "board"

	confRom       = "rom"
	confCharRom   = "charrom"
	confCpu       = "cpu"
	confSpeed     = "speed"
	confRamworks  = "ramworks"
	confNsc       = "nsc"
	confTrace     = "trace"
	confProfile   = "profile"
	confForceCaps = "forceCaps"
	confRgb       = "rgb"
	confRomx      = "romx"
	confMods      = "mods"
	confS0        = "s0"
	confS1        = "s1"
	confS2        = "s2"
	confS3        = "s3"
	confS4        = "s4"
	confS5        = "s5"
	confS6        = "s6"
	confS7        = "s7"
)

//go:embed configs/*.cfg
var configurationFiles embed.FS

type configurationModels struct {
	preconfiguredConfigs map[string]*configuration
}

type configuration struct {
	data map[string]string
}

func newConfiguration() *configuration {
	c := configuration{}
	c.data = make(map[string]string)
	return &c
}

func (c *configuration) getHas(key string) (string, bool) {
	key = strings.ToLower(key)
	value, ok := c.data[key]
	return value, ok
}

func (c *configuration) has(key string) bool {
	_, ok := c.getHas(key)
	return ok
}

func (c *configuration) get(key string) string {
	value, ok := c.getHas(key)
	if !ok {
		// Should not happen
		panic(fmt.Errorf("key %s not found", key))
	}
	return value
}

func (c *configuration) getFlag(key string) bool {
	return c.get(key) == "true"
}

func (c *configuration) set(key string, value string) {
	key = strings.ToLower(key)
	c.data[key] = value
}

func loadConfigurationModelsAndDefault() (*configurationModels, *configuration, error) {
	models := &configurationModels{}
	dir, err := configurationFiles.ReadDir("configs")
	if err != nil {
		return nil, nil, err
	}

	models.preconfiguredConfigs = make(map[string]*configuration)
	for _, file := range dir {
		if file.Type().IsRegular() && strings.HasSuffix(strings.ToLower(file.Name()), configSuffix) {
			content, err := configurationFiles.ReadFile("configs/" + file.Name())
			if err != nil {
				return nil, nil, err
			}
			lines := strings.Split(string(content), "\n")
			config := newConfiguration()
			for iLine, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				colonPos := strings.Index(line, ":")
				if colonPos < 0 {
					return nil, nil, fmt.Errorf("invalid configuration in %s:%d", file.Name(), iLine)
				}
				key := strings.TrimSpace(line[:colonPos])
				value := strings.TrimSpace(line[colonPos+1:])
				config.data[key] = value
			}
			name_no_ext := file.Name()[:len(file.Name())-len(configSuffix)]
			models.preconfiguredConfigs[name_no_ext] = config
		}
	}

	defaultConfig, err := models.get(defaultConfiguration)
	if err != nil {
		return nil, nil, err
	}
	defaultConfig.set(confModel, defaultConfiguration)

	return models, defaultConfig, nil
}

func mergeConfigs(base *configuration, addition *configuration) *configuration {
	result := newConfiguration()
	for k, v := range base.data {
		result.set(k, v)
	}
	for k, v := range addition.data {
		result.set(k, v)
	}
	return result
}

func (c *configurationModels) get(name string) (*configuration, error) {
	name = strings.TrimSpace(name)
	config, ok := c.preconfiguredConfigs[name]
	if !ok {
		return nil, fmt.Errorf("configuration %s.cfg not found", name)
	}

	parentName, hasParent := config.getHas(confParent)
	if !hasParent {
		return config, nil
	}

	parent, err := c.get(parentName)
	if err != nil {
		return nil, err
	}

	result := mergeConfigs(parent, config)
	return result, nil
}

func (c *configurationModels) availableModels() []string {
	models := make([]string, 0, len(c.preconfiguredConfigs)-1)
	for name := range c.preconfiguredConfigs {
		if !strings.HasPrefix(name, "_") {
			models = append(models, name)
		}
	}
	slices.Sort(models)
	return models
}

func (c *configurationModels) getWithOverrides(model string, overrides *configuration) (*configuration, error) {
	configValues, err := c.get(model)
	if err != nil {
		return nil, err
	}

	if overrides != nil {
		configValues = mergeConfigs(configValues, overrides)
	}
	return configValues, nil
}

func setupFlags(models *configurationModels, configuration *configuration) error {
	paramDescription := map[string]string{
		confModel:     "set base model",
		confRom:       "main rom file",
		confCharRom:   "rom file for the character generator",
		confCpu:       "cpu type, can be '6502' or '65c02'",
		confSpeed:     "cpu speed in Mhz, can be 'ntsc', 'pal', 'full' or a decimal nunmber",
		confMods:      "comma separated list of mods applied to the board, available mods are 'shift', 'four-colors",
		confRamworks:  "memory to use with RAMWorks card, max is 16384",
		confNsc:       "add a DS1216 No-Slot-Clock on the main ROM (use 'main') or a slot ROM",
		confTrace:     "trace CPU execution with one or more comma separated tracers",
		confProfile:   "generate profile trace to analyse with pprof",
		confForceCaps: "force all letters to be uppercased (no need for caps lock!)",
		confRgb:       "emulate the RGB modes of the 80col RGB card for DHGR",
		confRomx:      "emulate a RomX",
		confS0:        "slot 0 configuration.",
		confS1:        "slot 1 configuration.",
		confS2:        "slot 2 configuration.",
		confS3:        "slot 3 configuration.",
		confS4:        "slot 4 configuration.",
		confS5:        "slot 5 configuration.",
		confS6:        "slot 6 configuration.",
		confS7:        "slot 7 configuration.",
	}

	boolParams := []string{confProfile, confForceCaps, confRgb, confRomx}

	for name, description := range paramDescription {
		defaultValue, ok := configuration.getHas(name)
		if !ok {
			return fmt.Errorf("default value not found for %s", name)
		}
		if slices.Contains(boolParams, name) {
			flag.Bool(name, defaultValue == "true", description)
		} else {
			flag.String(name, defaultValue, description)
		}
	}

	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage:  %s [file]\n", flag.CommandLine.Name())
		fmt.Fprintf(out, "  file\n")
		fmt.Fprintf(out, "    	path to image to use on the boot device\n")
		flag.PrintDefaults()

		fmt.Fprintf(out, "\nThe available pre-configured models are:\n")
		for _, model := range models.availableModels() {
			config, _ := models.get(model)
			fmt.Fprintf(out, "  %s: %s\n", model, config.get(confName))
		}

		fmt.Fprintf(out, "\nThe available cards are:\n")
		for _, card := range availableCards() {
			builder := getCardFactory()[card]
			fmt.Fprintf(out, "  %s: %s\n", card, builder.description)
		}

		fmt.Fprintf(out, "\nThe available tracers are:\n")
		for _, tracer := range availableTracers() {
			builder := getTracerFactory()[tracer]
			fmt.Fprintf(out, "  %s: %s\n", tracer, builder.description)
		}
	}

	return nil
}

func getConfigurationFromCommandLine() (*configuration, string, error) {
	models, configuration, err := loadConfigurationModelsAndDefault()
	if err != nil {
		return nil, "", err
	}

	setupFlags(models, configuration)

	flag.Parse()

	modelFlag := flag.Lookup(confModel)
	if modelFlag != nil && strings.TrimSpace(modelFlag.Value.String()) != defaultConfiguration {
		// Replace the model
		configuration, err = models.get(modelFlag.Value.String())
		if err != nil {
			return nil, "", err
		}
	}

	flag.Visit(func(f *flag.Flag) {
		configuration.set(f.Name, f.Value.String())
	})

	filename := flag.Arg(0)

	return configuration, filename, nil
}
