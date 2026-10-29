package stateful

import (
	"errors"
	"fmt"
)

// ErrUnexpectedEvent matches every [*UnexpectedEventError] when used with
// [errors.Is]. Use it to check for unexpected events without naming the state
// and event types.
var ErrUnexpectedEvent = errors.New("unexpected event")

// UnexpectedEventError is returned when an event with no transition from the
// current state is applied to a state machine which has strict transitions
// enabled. See [Builder.StrictTransitions].
//
// Use [errors.As] to retrieve the state and event:
//
//	var unexpected *stateful.UnexpectedEventError[State, Event]
//	if errors.As(err, &unexpected) {
//		fmt.Println(unexpected.State, unexpected.Event)
//	}
type UnexpectedEventError[S, E comparable] struct {
	// State is the state the machine was in when the event was applied.
	State S
	// Event is the unexpected event.
	Event E
}

// Error describes the unexpected event and the state it was applied in.
func (e *UnexpectedEventError[S, E]) Error() string {
	return fmt.Sprintf("an unexpected event of type %v happened during the %v state", e.Event, e.State)
}

// Is reports whether target is [ErrUnexpectedEvent].
func (e *UnexpectedEventError[S, E]) Is(target error) bool {
	return target == ErrUnexpectedEvent
}
