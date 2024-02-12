package izapple2

import (
	"strings"

	"github.com/ivanizag/izapple2/screen"
)

type terminateConditionFunc func(a *Apple2) bool

type apple2Tester struct {
	a                  *Apple2
	terminateCondition terminateConditionFunc
}

func makeApple2Tester(model string, overrides *configuration) (*apple2Tester, error) {
	models, _, err := loadConfigurationModelsAndDefault()
	if err != nil {
		return nil, err
	}

	config, err := models.getWithOverrides(model, overrides)
	if err != nil {
		return nil, err
	}
	config.set(confSpeed, "full")
	a, err := configure(config)
	if err != nil {
		return nil, err
	}

	var at apple2Tester
	a.addTracer(&at)
	return &at, nil
}

func (at *apple2Tester) connect(a *Apple2) {
	at.a = a
}

func (at *apple2Tester) inspect() {
	if at.terminateCondition(at.a) {
		at.a.SendCommand(CommandKill)
	}
}

func (at *apple2Tester) run() {
	at.a.Run()
}

func (at *apple2Tester) getText() string {
	return screen.RenderTextModeString(at.a, false, false, false, at.a.isApple2e)
}

func (at *apple2Tester) getText80() string {
	return screen.RenderTextModeString(at.a, true, false, false, at.a.isApple2e)
}

/*
	func buildTerminateConditionCycles(cycles uint64) terminateConditionFunc {
		return func(a *Apple2) bool {
			return a.cpu.GetCycles() > cycles
		}
	}
*/

const textCheckInterval = uint64(100_000)

func buildTerminateConditionText(at *apple2Tester, needle string, col80 bool, timeoutCycles uint64) terminateConditionFunc {
	needles := []string{needle}
	return buildTerminateConditionTexts(at, needles, col80, timeoutCycles)
}

func buildTerminateConditionTexts(at *apple2Tester, needles []string, col80 bool, timeoutCycles uint64) terminateConditionFunc {
	lastCheck := uint64(0)
	found := false
	return func(a *Apple2) bool {
		cycles := a.cpu.GetCycles()
		if cycles > timeoutCycles {
			return true
		}
		if cycles-lastCheck > textCheckInterval {
			lastCheck = cycles
			var text string
			if col80 {
				text = at.getText80()
			} else {
				text = at.getText()
			}
			for _, needle := range needles {
				if !strings.Contains(text, needle) {
					return false
				}
			}
			found = true
		}
		return found
	}
}
