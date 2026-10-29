package stateful

import (
	"errors"
	"fmt"
)

var ErrUnexpectedEvent = errors.New("unexpected event")

type UnexpectedEventError[S, E comparable] struct {
	State S
	Event E
}

func (e *UnexpectedEventError[S, E]) Error() string {
	return fmt.Sprintf("an unexpected event of type %v happened during the %v state", e.Event, e.State)
}

func (e *UnexpectedEventError[S, E]) Is(target error) bool {
	return target == ErrUnexpectedEvent
}
