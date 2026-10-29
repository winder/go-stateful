package stateful

// Machine is a simple event based state machine. Create one with
// [Builder.Build]; the zero value is not usable.
//
// A Machine is not safe for concurrent use.
type Machine[S, E comparable, C any] struct {
	current *node[S, E, C]
	strict  bool
}

// Apply applies event to the state machine. It is equivalent to calling
// [Machine.ApplyWith] with the zero value of C.
func (m *Machine[S, E, C]) Apply(event E) error {
	var zero C
	return m.ApplyWith(event, zero)
}

// ApplyWith applies event to the state machine and passes the context value c
// to every listener called by the transition. C can be any type, including a
// [context.Context].
//
// If the current state has a transition for event, the exit listeners of the
// current state are called, the machine moves to the transition's end state,
// and the enter listeners of the end state are called. Listeners can observe
// the transition through [Machine.State]: exit listeners see the old state and
// enter listeners see the new state.
//
// If there is no transition for event, nothing happens. When strict
// transitions are enabled, see [Builder.StrictTransitions], ApplyWith returns
// an [*UnexpectedEventError]. Otherwise it returns nil.
func (m *Machine[S, E, C]) ApplyWith(event E, c C) error {
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
func (m *Machine[S, E, C]) State() S {
	return m.current.state
}
