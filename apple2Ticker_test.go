package izapple2

import (
	"testing"
)

// testTicker counts its calls and asks for the cycle in next
type testTicker struct {
	calls  int
	called []uint64
	next   uint64
	a      *Apple2
}

func (tt *testTicker) tick() uint64 {
	tt.calls++
	tt.called = append(tt.called, tt.a.GetCycles())
	return tt.next
}

func TestTickersRunWhenDue(t *testing.T) {
	a, err := CreateAppleFromModel("2plus", map[string]string{"s6": "empty"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	a.Init()
	tt := &testTicker{a: a, next: tickNever}
	a.registerTicker(tt)

	// The first call comes on the next instruction
	a.RunCycles(100)
	if tt.calls != 1 {
		t.Fatalf("expected one call after registering, got %v", tt.calls)
	}

	// Not before the cycle it asks for
	tt.next = a.GetCycles() + 10_000
	a.requestTickAt(a.GetCycles())
	a.RunCycles(100)
	calls := tt.calls
	a.RunCycles(5_000)
	if tt.calls != calls {
		t.Errorf("expected no calls before the cycle asked for, got %v", tt.calls-calls)
	}
	due := tt.next
	tt.next = tickNever
	a.RunCycles(10_000)
	if tt.calls != calls+1 || tt.called[len(tt.called)-1] < due {
		t.Errorf("expected one call at cycle %v, got calls at %v", due, tt.called[calls:])
	}

	// A request brings it forward
	a.requestTickAt(a.GetCycles() + 50)
	a.RunCycles(100)
	if tt.calls != calls+2 {
		t.Errorf("expected the call requested, got %v calls", tt.calls-calls-1)
	}
}
