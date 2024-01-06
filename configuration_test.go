package izapple2

import (
	"testing"
)

func TestConfigurationModel(t *testing.T) {

	t.Run("test that the default model exists", func(t *testing.T) {
		models, err := initConfigurationModels()
		if err != nil {
			t.Fatal(err)
		}
		_, err = models.getFromModel(defaultConfiguration)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("test preconfigured models are complete", func(t *testing.T) {
		models, err := initConfigurationModels()
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
			model, err := models.getFromModel(modelName)
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
