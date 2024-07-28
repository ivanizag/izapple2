package izapple2

import (
	"strings"
	"testing"
)

func TestDan2Controller(t *testing.T) {
	overrides := newConfiguration()
	overrides.set(confS7, "dan2sd,slot1=resources/ProDOS_2_4_3.po")

	at, err := makeApple2Tester("2enh", overrides)
	if err != nil {
		t.Fatal(err)
	}

	at.terminateCondition = buildTerminateConditionText("NEW VOL", testTextMode40, 10_000_000)

	at.run()

	text := at.getText(testTextMode40)
	if !strings.Contains(text, "NEW VOL") {
		t.Errorf("Expected Bitsy Bye screen, got '%s'", text)
	}

}
