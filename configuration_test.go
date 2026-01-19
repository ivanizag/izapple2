package izapple2

import (
	"flag"
	"os"
	"strings"
	"testing"
)

func TestConfigurationModel(t *testing.T) {

	t.Run("test that the default model exists", func(t *testing.T) {
		_, _, err := loadConfigurationModelsAndDefault()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("test preconfigured models are complete", func(t *testing.T) {
		models, _, err := loadConfigurationModelsAndDefault()
		if err != nil {
			t.Fatal(err)
		}

		requiredFields := []string{
			confRom, confCharRom, confCpu, confSpeed, confRamworks, confNsc,
			confTrace, confProfile, confShowConfig, confForceCaps, confRgb, confRomx,
			confS0, confS1, confS2, confS3, confS4, confS5, confS6, confS7,
		}
		availabledModels := models.availableModels()
		for _, modelName := range availabledModels {
			model, err := models.get(modelName)
			if err != nil {
				t.Error(err)
			}

			for _, field := range requiredFields {
				if _, ok := model.getHas(field); !ok {
					t.Errorf("missing field '%s' in the preconfigured model '%s'", field, modelName)
				}
			}
		}
	})
}

func TestCommandLineHelp(t *testing.T) {
	t.Run("test command line help", func(t *testing.T) {
		models, configuration, err := loadConfigurationModelsAndDefault()
		if err != nil {
			t.Fatal(err)
		}

		prevFlags := flag.CommandLine
		flag.CommandLine = flag.NewFlagSet("izapple2", flag.ExitOnError)

		setupFlags(models, configuration)

		buffer := strings.Builder{}
		flag.CommandLine.SetOutput(&buffer)
		flag.Usage()
		usage := buffer.String()

		flag.CommandLine = prevFlags

		prevous, err := os.ReadFile("doc/usage.txt")
		if err != nil {
			t.Fatal(err)
		}

		if usage != string(prevous) {
			os.WriteFile("doc/usage_new.txt", []byte(usage), 0644)
			t.Errorf(`Usage has changed, check doc/usage_new.txt for the new version.
If it is correct, execute \"go run update_readme.go\" in the doc folder.`)
		}
	})
}

func TestExpandSlotConfiguration(t *testing.T) {
	t.Run("test empty configuration", func(t *testing.T) {
		result, err := expandSlotConfiguration("")
		if err != nil {
			t.Error(err)
		}
		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})

	t.Run("test special value 'empty'", func(t *testing.T) {
		result, err := expandSlotConfiguration("empty")
		if err != nil {
			t.Error(err)
		}
		if result != "empty" {
			t.Errorf("expected 'empty', got %s", result)
		}
	})

	t.Run("test full configuration with card name", func(t *testing.T) {
		input := "diskii,disk1=\"test.dsk\""
		result, err := expandSlotConfiguration(input)
		if err != nil {
			t.Error(err)
		}
		if result != input {
			t.Errorf("expected %s, got %s", input, result)
		}
	})

	t.Run("test configuration with parameters", func(t *testing.T) {
		input := "trace=true"
		result, err := expandSlotConfiguration(input)
		if err != nil {
			t.Error(err)
		}
		if result != input {
			t.Errorf("expected %s, got %s", input, result)
		}
	})

	t.Run("test single diskette file expansion", func(t *testing.T) {
		result, err := expandSlotConfiguration("resources/dos33.dsk")
		if err != nil {
			t.Error(err)
		}
		expected := "diskii,disk1=resources/dos33.dsk"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("test multiple diskette files expansion", func(t *testing.T) {
		result, err := expandSlotConfiguration("resources/dos33.dsk,resources/audit.dsk")
		if err != nil {
			t.Error(err)
		}
		expected := "diskii,disk1=resources/dos33.dsk,disk2=resources/audit.dsk"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("test too many diskettes error", func(t *testing.T) {
		_, err := expandSlotConfiguration("resources/dos33.dsk,resources/audit.dsk,resources/dos33.dsk")
		if err == nil {
			t.Error("expected error for more than 2 diskettes")
		}
	})

	t.Run("test disk alias expansion", func(t *testing.T) {
		result, err := expandSlotConfiguration("dos33")
		if err != nil {
			t.Error(err)
		}
		expected := "diskii,disk1=dos33"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})
}

func TestProcessPositionalFilenames(t *testing.T) {
	t.Run("test empty filenames", func(t *testing.T) {
		config := newConfiguration()
		err := processPositionalFilenames(config, []string{})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("test single diskette", func(t *testing.T) {
		config := newConfiguration()
		err := processPositionalFilenames(config, []string{"resources/dos33.dsk"})
		if err != nil {
			t.Error(err)
		}
		s6 := config.get(confS6)
		expected := "diskii,disk1=\"resources/dos33.dsk\""
		if s6 != expected {
			t.Errorf("expected s6=%s, got %s", expected, s6)
		}
	})

	t.Run("test two diskettes", func(t *testing.T) {
		config := newConfiguration()
		err := processPositionalFilenames(config, []string{"resources/dos33.dsk", "resources/audit.dsk"})
		if err != nil {
			t.Error(err)
		}
		s6 := config.get(confS6)
		expected := "diskii,disk1=\"resources/dos33.dsk\",disk2=\"resources/audit.dsk\""
		if s6 != expected {
			t.Errorf("expected s6=%s, got %s", expected, s6)
		}
	})

	t.Run("test three diskettes", func(t *testing.T) {
		config := newConfiguration()
		err := processPositionalFilenames(config, []string{
			"resources/dos33.dsk",
			"resources/audit.dsk",
			"resources/dos33.dsk",
		})
		if err != nil {
			t.Error(err)
		}
		s6 := config.get(confS6)
		s5 := config.get(confS5)
		expectedS6 := "diskii,disk1=\"resources/dos33.dsk\",disk2=\"resources/audit.dsk\""
		expectedS5 := "diskii,disk1=\"resources/dos33.dsk\""
		if s6 != expectedS6 {
			t.Errorf("expected s6=%s, got %s", expectedS6, s6)
		}
		if s5 != expectedS5 {
			t.Errorf("expected s5=%s, got %s", expectedS5, s5)
		}
	})

	t.Run("test four diskettes", func(t *testing.T) {
		config := newConfiguration()
		err := processPositionalFilenames(config, []string{
			"resources/dos33.dsk",
			"resources/audit.dsk",
			"resources/dos33.dsk",
			"resources/audit.dsk",
		})
		if err != nil {
			t.Error(err)
		}
		s6 := config.get(confS6)
		s5 := config.get(confS5)
		expectedS6 := "diskii,disk1=\"resources/dos33.dsk\",disk2=\"resources/audit.dsk\""
		expectedS5 := "diskii,disk1=\"resources/dos33.dsk\",disk2=\"resources/audit.dsk\""
		if s6 != expectedS6 {
			t.Errorf("expected s6=%s, got %s", expectedS6, s6)
		}
		if s5 != expectedS5 {
			t.Errorf("expected s5=%s, got %s", expectedS5, s5)
		}
	})

	t.Run("test too many diskettes error", func(t *testing.T) {
		config := newConfiguration()
		err := processPositionalFilenames(config, []string{
			"resources/dos33.dsk",
			"resources/dos33.dsk",
			"resources/dos33.dsk",
			"resources/dos33.dsk",
			"resources/dos33.dsk",
		})
		if err == nil {
			t.Error("expected error for more than 4 diskettes")
		}
	})

	t.Run("test block device", func(t *testing.T) {
		config := newConfiguration()
		// Use a file that won't be detected as a diskette
		err := processPositionalFilenames(config, []string{"resources/ProDOS_2_4_3.po"})
		if err != nil {
			t.Error(err)
		}
		// ProDOS .po files are detected as diskettes, so this will set s6
		s6 := config.get(confS6)
		if !strings.Contains(s6, "diskii") {
			t.Errorf("expected diskii configuration, got %s", s6)
		}
	})
}

func TestApplyDiskAliases(t *testing.T) {
	t.Run("test dos33 alias", func(t *testing.T) {
		result := applyDiskAliases("dos33")
		expected := "<internal>/dos33.dsk"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("test prodos alias", func(t *testing.T) {
		result := applyDiskAliases("prodos")
		expected := "<internal>/ProDOS_2_4_3.po"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("test non-alias filename", func(t *testing.T) {
		input := "test.dsk"
		result := applyDiskAliases(input)
		if result != input {
			t.Errorf("expected %s, got %s", input, result)
		}
	})
}
