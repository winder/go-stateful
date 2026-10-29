package stateful

// StatefulFunc is a listener called when a state machine enters or exits a
// state. It receives the context passed to [StateMachine.ApplyWithContext], or
// the zero value of C when the event was applied with [StateMachine.Apply].
type StatefulFunc[C any] func(c C)

// node is a single state in the transition graph along with its outgoing
// transitions and listeners.
type node[S, E comparable, C any] struct {
	state     S
	neighbors map[E]*node[S, E, C]
	onEnter   []StatefulFunc[C]
	onExit    []StatefulFunc[C]
}

func newNode[S, E comparable, C any](state S) *node[S, E, C] {
	return &node[S, E, C]{
		state:     state,
		neighbors: make(map[E]*node[S, E, C]),
	}
}

// neighbor returns the node reached from n when event is applied.
func (n *node[S, E, C]) neighbor(event E) (*node[S, E, C], bool) {
	next, ok := n.neighbors[event]
	return next, ok
}

// enter calls the enter listeners in registration order.
func (n *node[S, E, C]) enter(c C) {
	for _, f := range n.onEnter {
		f(c)
	}
}

// exit calls the exit listeners in registration order.
func (n *node[S, E, C]) exit(c C) {
	for _, f := range n.onExit {
		f(c)
	}
}
