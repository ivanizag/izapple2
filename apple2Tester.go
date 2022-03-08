package izapple2

import (
	"github.com/ivanizag/izapple2/screen"
)

type apple2Tester struct {
	a                  *Apple2
	terminateCondition func(a *Apple2) bool
}

func makeApple2Tester(model string) *apple2Tester {
	a := newApple2()
	a.setup(0, true) // Full speed
	initModel(a, model, defaultInternal, defaultInternal)

	a.AddLanguageCard(0)

	var at apple2Tester
	at.a = a
	a.addTracer(&at)
	return &at
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
