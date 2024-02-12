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
			confTrace, confProfile, confForceCaps, confRgb, confRomx,
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
