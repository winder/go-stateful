package stateful

// StateMachine is a simple event based state machine. Create one with
// [Builder.Build]; the zero value is not usable.
//
// A StateMachine is not safe for concurrent use.
type StateMachine[S, E comparable, C any] struct {
	current *node[S, E, C]
	strict  bool
}

// Apply applies event to the state machine with the zero value of C as the
// context. It is equivalent to calling [StateMachine.ApplyWithContext] with
// the zero value of C.
func (m *StateMachine[S, E, C]) Apply(event E) error {
	var zero C
	return m.ApplyWithContext(event, zero)
}

// ApplyWithContext applies event to the state machine and passes c to every
// listener called by the transition.
//
// If the current state has a transition for event, the exit listeners of the
// current state are called, the machine moves to the transition's end state,
// and the enter listeners of the end state are called. Listeners can observe
// the transition through [StateMachine.State]: exit listeners see the old
// state and enter listeners see the new state.
//
// If there is no transition for event, nothing happens. When strict
// transitions are enabled, see [Builder.StrictTransitions], ApplyWithContext
// returns an [*UnexpectedEventError]. Otherwise it returns nil.
//
// Despite the name, C does not need to be a [context.Context], although it
// may be.
func (m *StateMachine[S, E, C]) ApplyWithContext(event E, c C) error {
	next, ok := m.current.neighbor(event)
	if !ok {
		if m.strict {
			return &UnexpectedEventError[S, E]{State: m.current.state, Event: event}
		}
		return nil
	}

	m.current.exit(c)
	m.current = next
	m.current.enter(c)
	return nil
}

// State returns the current state of the state machine.
func (m *StateMachine[S, E, C]) State() S {
	return m.current.state
}
