package izapple2

import (
	"github.com/ivanizag/izapple2/screen"
)

type apple2Tester struct {
	a                  *Apple2
	terminateCondition func(a *Apple2) bool
}

func makeApple2Tester(model string, overrides *configuration) (*apple2Tester, error) {
	config, err := getConfigurationFromModel(model, overrides)
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
