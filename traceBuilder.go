package izapple2

import (
	"fmt"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type executionTracer interface {
	connect(a *Apple2)
	inspect()
}

type traceBuilder struct {
	name            string
	description     string
	executionTracer executionTracer
	connectFunc     func(a *Apple2)
}

var traceFactory map[string]*traceBuilder

func getTracerFactory() map[string]*traceBuilder {
	if traceFactory != nil {
		return traceFactory
	}

	tracerFactory := make(map[string]*traceBuilder)

	tracerFactory["mos"] = &traceBuilder{
		name:            "mos",
		description:     "Trace MOS calls with Applecorn skipping terminal IO",
		executionTracer: newTraceApplecorn(true),
	}
	tracerFactory["mosfull"] = &traceBuilder{
		name:            "mosfull",
		description:     "Trace MOS calls with Applecorn",
		executionTracer: newTraceApplecorn(false),
	}
	tracerFactory["mli"] = &traceBuilder{
		name:            "mli",
		description:     "Trace ProDOS MLI calls",
		executionTracer: newTraceProDOS(),
	}
	tracerFactory["ucsd"] = &traceBuilder{
		name:            "ucsd",
		description:     "Trace UCSD system calls",
		executionTracer: newTracePascal(),
	}
	tracerFactory["cpu"] = &traceBuilder{
		name:        "cpu",
		description: "Trace CPU execution",
		connectFunc: func(a *Apple2) {
			a.cpuTrace = true
			a.cpu.SetTrace(true)
		},
	}
	tracerFactory["ss"] = &traceBuilder{
		name:        "ss",
		description: "Trace sotfswiches calls",
		connectFunc: func(a *Apple2) { a.io.setTrace(true) },
	}
	tracerFactory["ssreg"] = &traceBuilder{
		name:        "ssreg",
		description: "Trace sotfswiches registrations",
		connectFunc: func(a *Apple2) { a.io.setTraceRegistrations(true) },
	}
	tracerFactory["panicss"] = &traceBuilder{
		name:        "panicss",
		description: "Panic on unimplemented softswitches",
		connectFunc: func(a *Apple2) { a.io.setPanicNotImplemented(true) },
	}
	tracerFactory["cpm65"] = &traceBuilder{
		name:            "cpm65",
		description:     "Trace CPM65 BDOS calls",
		executionTracer: newTraceCpm65(false),
	}
	return tracerFactory
}

func availableTracers() []string {
	names := maps.Keys(getTracerFactory())
	slices.Sort(names)
	return names
}

func setupTracers(a *Apple2, paramString string) error {
	tracerFactory := getTracerFactory()
	tracerNames := splitConfigurationString(paramString, ',')
	for _, tracer := range tracerNames {
		tracer = strings.ToLower(strings.TrimSpace(tracer))
		if tracer == "none" {
			continue
		}
		builder, ok := tracerFactory[tracer]
		if !ok {
			return fmt.Errorf("unknown tracer %s", tracer)
		}
		if builder.connectFunc != nil {
			builder.connectFunc(a)
		}
		if builder.executionTracer != nil {
			a.addTracer(builder.executionTracer)
		}
	}
	return nil
}

func (a *Apple2) addTracer(tracer executionTracer) {
	tracer.connect(a)
	a.tracers = append(a.tracers, tracer)
}
