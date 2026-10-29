package stateful_test

import (
	"fmt"

	"github.com/winder/go-stateful"
)

// State is the state of a job.
type State int

const (
	Init State = iota
	Running
	Paused
	Completed
)

func (s State) String() string {
	return [...]string{"INIT", "RUNNING", "PAUSED", "COMPLETED"}[s]
}

// Event is something that happens to a job.
type Event int

const (
	Run Event = iota
	Pause
	End
)

func (e Event) String() string {
	return [...]string{"RUN", "PAUSE", "END"}[e]
}

// Context is passed to listeners when an event is applied.
type Context struct {
	Message string
}

func Example() {
	sm := stateful.NewBuilder[State, Event, *Context](Init).
		AddTransition(Init, Run, Running).
		AddTransition(Running, Pause, Paused).
		AddTransition(Paused, Run, Running).
		AddTransition(Running, End, Completed).
		OnExit(Init, func(c *Context) {
			fmt.Println("Exiting INIT:", c.Message)
		}).
		OnEnter(Running, func(c *Context) {
			fmt.Println("Entering RUNNING:", c.Message)
		}).
		OnEnter(Completed, func(c *Context) {
			fmt.Println("Entering COMPLETED:", c.Message)
		}).
		Build()

	fmt.Println(sm.State())

	events := []struct {
		event   Event
		message string
	}{
		{Run, "started"},
		{Pause, "taking a break"},
		{Run, "resumed"},
		{End, "finished"},
	}
	for _, e := range events {
		if err := sm.ApplyWithContext(e.event, &Context{Message: e.message}); err != nil {
			fmt.Println(err)
		}
	}

	// The job is complete, so it can no longer run.
	if err := sm.ApplyWithContext(Run, &Context{Message: "again"}); err != nil {
		fmt.Println(err)
	}

	fmt.Println(sm.State())

	// Output:
	// INIT
	// Exiting INIT: started
	// Entering RUNNING: started
	// Entering RUNNING: resumed
	// Entering COMPLETED: finished
	// an unexpected event of type RUN happened during the COMPLETED state
	// COMPLETED
}
