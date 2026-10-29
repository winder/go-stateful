package stateful

type StateMachine[S, E comparable, C any] struct {
	current *node[S, E, C]
	strict  bool
}

func (m *StateMachine[S, E, C]) Apply(event E) error {
	var zero C
	return m.ApplyWithContext(event, zero)
}

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

func (m *StateMachine[S, E, C]) State() S {
	return m.current.state
}
