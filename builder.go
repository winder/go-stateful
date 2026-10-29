package stateful

import "slices"

// Builder declares the transitions and listeners of a state machine. Create
// one with [NewBuilder], declare the machine with the chainable methods, then
// call [Builder.Build].
//
// S is the state type, E is the event type and C is the type of context passed
// to listeners. See the package documentation for details.
//
// A Builder is not safe for concurrent use.
type Builder[S, E comparable, C any] struct {
	nodes  map[S]*node[S, E, C]
	root   *node[S, E, C]
	strict bool
}

// NewBuilder returns a builder for a state machine which starts in
// initialState. Strict transitions are enabled by default.
//
// Go cannot infer the event and context types, so all three type arguments
// must be given:
//
//	b := stateful.NewBuilder[State, Event, *Context](Init)
func NewBuilder[S, E comparable, C any](initialState S) *Builder[S, E, C] {
	root := newNode[S, E, C](initialState)
	return &Builder[S, E, C]{
		nodes:  map[S]*node[S, E, C]{initialState: root},
		root:   root,
		strict: true,
	}
}

// node returns the node for state, creating it if necessary.
func (b *Builder[S, E, C]) node(state S) *node[S, E, C] {
	n, ok := b.nodes[state]
	if !ok {
		n = newNode[S, E, C](state)
		b.nodes[state] = n
	}
	return n
}

// AddTransition adds a transition from startState to endState which happens
// when event is applied while the machine is in startState. Adding a
// transition for a startState and event pair that already has one replaces it.
// startState and endState may be the same, in which case the state's exit and
// enter listeners are both called.
//
// It returns b so that calls can be chained.
func (b *Builder[S, E, C]) AddTransition(startState S, event E, endState S) *Builder[S, E, C] {
	start := b.node(startState)
	end := b.node(endState)
	start.neighbors[event] = end
	return b
}

// OnEnter adds a listener which is called each time the machine enters state.
// It is not called for the initial state when the machine is built. A state
// may have several enter listeners; they are called in the order they were
// added. A nil listener is ignored.
//
// It returns b so that calls can be chained.
func (b *Builder[S, E, C]) OnEnter(state S, onEnter Listener[C]) *Builder[S, E, C] {
	n := b.node(state)
	if onEnter != nil {
		n.onEnter = append(n.onEnter, onEnter)
	}
	return b
}

// OnExit adds a listener which is called each time the machine exits state.
// A state may have several exit listeners; they are called in the order they
// were added. A nil listener is ignored.
//
// It returns b so that calls can be chained.
func (b *Builder[S, E, C]) OnExit(state S, onExit Listener[C]) *Builder[S, E, C] {
	n := b.node(state)
	if onExit != nil {
		n.onExit = append(n.onExit, onExit)
	}
	return b
}

// StrictTransitions configures whether applying an event which has no
// transition from the current state is an error. When strict is true, the
// default, [Machine.Apply] and [Machine.ApplyWith] return an
// [*UnexpectedEventError]. When strict is false they return nil. In both cases
// the state does not change and no listeners are called.
//
// It returns b so that calls can be chained.
func (b *Builder[S, E, C]) StrictTransitions(strict bool) *Builder[S, E, C] {
	b.strict = strict
	return b
}

// Build returns a new state machine in the builder's initial state, using the
// transitions, listeners and settings declared so far.
//
// The machine gets its own copy of the declaration, so changes made to b after
// Build do not affect it. Build may be called several times to create
// independent machines.
func (b *Builder[S, E, C]) Build() *Machine[S, E, C] {
	// Copy the graph so that later changes to the builder do not affect
	// machines which have already been built.
	nodes := make(map[S]*node[S, E, C], len(b.nodes))
	for state, n := range b.nodes {
		nodes[state] = &node[S, E, C]{
			state:     state,
			neighbors: make(map[E]*node[S, E, C], len(n.neighbors)),
			onEnter:   slices.Clone(n.onEnter),
			onExit:    slices.Clone(n.onExit),
		}
	}
	for state, n := range b.nodes {
		for event, end := range n.neighbors {
			nodes[state].neighbors[event] = nodes[end.state]
		}
	}

	return &Machine[S, E, C]{
		current: nodes[b.root.state],
		strict:  b.strict,
	}
}
