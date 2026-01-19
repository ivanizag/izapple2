package izapple2

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"slices"
)

const configSuffix = ".cfg"
const defaultConfiguration = "2enh"

const (
	confParent = "parent"
	confModel  = "model"
	confName   = "name"
	confBoard  = "board"

	confRom        = "rom"
	confCharRom    = "charrom"
	confCpu        = "cpu"
	confSpeed      = "speed"
	confRamworks   = "ramworks"
	confNsc        = "nsc"
	confTrace      = "trace"
	confProfile    = "profile"
	confShowConfig = "showConfig"
	confForceCaps  = "forceCaps"
	confRgb        = "rgb"
	confRomx       = "romx"
	confMods       = "mods"

	confS0 = "s0"
	confS1 = "s1"
	confS2 = "s2"
	confS3 = "s3"
	confS4 = "s4"
	confS5 = "s5"
	confS6 = "s6"
	confS7 = "s7"
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
			config, err := parseConfiguration(content, file.Name())
			if err != nil {
				return nil, nil, err
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

func parseConfiguration(content []byte, name string) (*configuration, error) {
	lines := strings.Split(string(content), "\n")
	config := newConfiguration()
	for iLine, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		colonPos := strings.Index(line, ":")
		if colonPos < 0 {
			return nil, fmt.Errorf("invalid configuration in %s:%d", name, iLine)
		}
		key := strings.TrimSpace(line[:colonPos])
		value := strings.TrimSpace(line[colonPos+1:])
		config.data[key] = value
	}
	return config, nil
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
		if filepath.Ext(strings.ToLower(name)) != ".cfg" {
			name = name + ".cfg"
		}
		if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
			path, exists := os.LookupEnv("IZAPPLE2_CUSTOM_MODEL")
			if exists {
				name = filepath.Join(path, name)
			}
		}
		content, err := os.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("configuration %s not found", name)
		}
		config, err = parseConfiguration(content, name)
		if err != nil {
			return nil, err
		}
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
		confModel:      "set base model",
		confRom:        "main rom file",
		confCharRom:    "rom file for the character generator",
		confCpu:        "cpu type, can be '6502' or '65c02'",
		confSpeed:      "cpu speed in Mhz, can be 'ntsc', 'pal', 'full' or a decimal nunmber",
		confMods:       "comma separated list of mods applied to the board, available mods are 'shift', 'four-colors",
		confRamworks:   "memory to use with RAMWorks card, max is 16384",
		confNsc:        "add a DS1216 No-Slot-Clock on the main ROM (use 'main') or a slot ROM",
		confTrace:      "trace CPU execution with one or more comma separated tracers",
		confProfile:    "generate profile trace to analyse with pprof",
		confShowConfig: "show the calculated configuration and exit",
		confForceCaps:  "force all letters to be uppercased (no need for caps lock!)",
		confRgb:        "emulate the RGB modes of the 80col RGB card for DHGR",
		confRomx:       "emulate a RomX",
		confS0:         "slot 0 configuration.",
		confS1:         "slot 1 configuration.",
		confS2:         "slot 2 configuration.",
		confS3:         "slot 3 configuration.",
		confS4:         "slot 4 configuration.",
		confS5:         "slot 5 configuration.",
		confS6:         "slot 6 configuration.",
		confS7:         "slot 7 configuration.",
	}

	boolParams := []string{confProfile, confShowConfig, confForceCaps, confRgb, confRomx}

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
		fmt.Fprintf(out, "Custom models may be specified by filename.  Use 'IZAPPLE2_CUSTOM_MODEL' to set default location.\n")

		fmt.Fprintf(out, "\nThe available cards are:\n")
		for _, card := range availableCards() {
			builder := getCardFactory()[card]
			if !builder.hide {
				fmt.Fprintf(out, "  %s: %s\n", card, builder.description)
			}
		}

		fmt.Fprintf(out, "\nThe available tracers are:\n")
		for _, tracer := range availableTracers() {
			builder := getTracerFactory()[tracer]
			fmt.Fprintf(out, "  %s: %s\n", tracer, builder.description)
		}
	}

	return nil
}

var diskAliases = map[string]string{
	"dos33":   "<internal>/dos33.dsk",
	"prodos":  "<internal>/ProDOS_2_4_3.po",
	"cpm":     "<internal>/cpm_2.20B_56K.po",
	"cardcat": "<internal>/Card Cat 1.7.dsk",
}

func applyDiskAliases(filename string) string {
	if alias, ok := diskAliases[filename]; ok {
		return alias
	}
	return filename
}

// classifyFile determines if a file is a diskette or block device
// Returns true if the file is a diskette, false if it's a block device
func classifyFile(filename string) bool {
	filename = applyDiskAliases(filename)
	_, err := LoadDiskette(filename)
	return err == nil
}

// processPositionalFilenames handles filenames passed as positional arguments
// and configures slots s6, s5, s7 based on the file types
func processPositionalFilenames(config *configuration, filenames []string) error {
	if len(filenames) == 0 {
		return nil
	}

	diskettes := []string{}
	blockDevices := []string{}

	for _, filename := range filenames {
		filename = applyDiskAliases(filename)
		if classifyFile(filename) {
			diskettes = append(diskettes, filename)
		} else {
			blockDevices = append(blockDevices, filename)
		}
	}

	// Configure diskette slots (s6 and s5)
	if len(diskettes) == 1 {
		config.set(confS6, fmt.Sprintf("diskii,disk1=\"%s\"", diskettes[0]))
	} else if len(diskettes) >= 2 {
		config.set(confS6, fmt.Sprintf("diskii,disk1=\"%s\",disk2=\"%s\"", diskettes[0], diskettes[1]))
	}
	if len(diskettes) == 3 {
		config.set(confS5, fmt.Sprintf("diskii,disk1=\"%s\"", diskettes[2]))
	} else if len(diskettes) >= 4 {
		config.set(confS5, fmt.Sprintf("diskii,disk1=\"%s\",disk2=\"%s\"", diskettes[2], diskettes[3]))
	}
	if len(diskettes) > 4 {
		return fmt.Errorf("up to 4 diskettes can be loaded, %v found", len(diskettes))
	}

	// Configure block device slots (s7 and s5)
	if len(blockDevices) > 8 {
		return fmt.Errorf("up to 8 block devices can be loaded, %v found", len(blockDevices))
	}
	if len(blockDevices) > 0 {
		config.set(confS7, fmt.Sprintf("smartport,image1=\"%s\"", blockDevices[0]))
		if len(blockDevices) > 1 {
			smartportConfig := "smartport"
			for i, filename := range blockDevices {
				if i == 0 {
					continue
				}
				smartportConfig += fmt.Sprintf(",image%v=\"%s\"", i+1, filename)
			}
			config.set(confS5, smartportConfig)
		}
	}

	return nil
}

// expandSlotConfiguration expands a partial slot configuration into a full one
// If the configuration is just filenames, it detects the file type and creates
// the appropriate card configuration (diskii for diskettes, smartport for block devices)
func expandSlotConfiguration(configString string) (string, error) {
	if configString == "" {
		return "", nil
	}

	// Split by comma to get parts (but respect quotes)
	parts := splitConfigurationString(configString, ',')
	if len(parts) == 0 {
		return configString, nil
	}

	// Check if first part is a card name or a filename
	firstPart := strings.TrimSpace(parts[0])

	// Skip expansion for special values
	if firstPart == noCardName {
		return configString, nil
	}

	// If it contains '=' it's already a parameter, so it's a full config
	if strings.Contains(firstPart, "=") {
		return configString, nil
	}

	// Check if first part is a known card name
	_, isCard := getCardFactory()[strings.ToLower(firstPart)]
	if isCard {
		// Already a full configuration
		return configString, nil
	}

	// It's a partial configuration - just filenames
	// Detect the file types and build the appropriate configuration
	diskettes := []string{}
	blockDevices := []string{}

	for _, part := range parts {
		filename := strings.TrimSpace(part)
		if filename == "" {
			continue
		}

		// Apply disk aliases
		filename = applyDiskAliases(filename)

		// Try to load as diskette
		_, err := LoadDiskette(filename)
		if err == nil {
			diskettes = append(diskettes, part) // Keep original part (may have quotes)
		} else {
			blockDevices = append(blockDevices, part)
		}
	}

	// Build the configuration based on what we found
	if len(diskettes) > 0 && len(blockDevices) == 0 {
		// All diskettes - create diskii configuration
		config := "diskii"
		for i, disk := range diskettes {
			diskNum := i + 1
			if diskNum > 2 {
				return "", fmt.Errorf("diskii card supports maximum 2 disks, got %d", len(diskettes))
			}
			config += fmt.Sprintf(",disk%d=%s", diskNum, disk)
		}
		return config, nil
	} else if len(blockDevices) > 0 && len(diskettes) == 0 {
		// All block devices - create smartport configuration
		config := "smartport"
		for i, device := range blockDevices {
			imageNum := i + 1
			config += fmt.Sprintf(",image%d=%s", imageNum, device)
		}
		return config, nil
	} else if len(diskettes) > 0 && len(blockDevices) > 0 {
		return "", fmt.Errorf("cannot mix diskettes and block devices in the same slot configuration")
	}

	// No valid files found
	return configString, nil
}

func getConfigurationFromCommandLine() (*configuration, error) {
	models, configuration, err := loadConfigurationModelsAndDefault()
	if err != nil {
		return nil, err
	}

	setupFlags(models, configuration)

	flag.Parse()

	modelFlag := flag.Lookup(confModel)
	if modelFlag != nil && strings.TrimSpace(modelFlag.Value.String()) != defaultConfiguration {
		// Replace the model
		configuration, err = models.get(modelFlag.Value.String())
		if err != nil {
			return nil, err
		}
	}

	flag.Visit(func(f *flag.Flag) {
		configuration.set(f.Name, f.Value.String())
	})

	// Expand partial slot configurations (e.g., "-s4 disk.dsk" -> "-s4 diskii,disk1=disk.dsk")
	slotParams := []string{confS0, confS1, confS2, confS3, confS4, confS5, confS6, confS7}
	for _, slotParam := range slotParams {
		if configuration.has(slotParam) {
			slotConfig := configuration.get(slotParam)
			expandedConfig, err := expandSlotConfiguration(slotConfig)
			if err != nil {
				return nil, fmt.Errorf("error expanding slot configuration for %s: %w", slotParam, err)
			}
			if expandedConfig != slotConfig {
				configuration.set(slotParam, expandedConfig)
			}
		}
	}

	// Process positional filenames (e.g., "program disk.dsk")
	filenames := flag.Args()
	err = processPositionalFilenames(configuration, filenames)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}

func (c *configuration) dump() {
	fmt.Println("Configuration:")

	keys := slices.Sorted(maps.Keys(c.data))

	for _, key := range keys {
		fmt.Printf("  %s: %s\n", key, c.data[key])
	}
}
