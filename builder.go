package stateful

import "slices"

type Builder[S, E comparable, C any] struct {
	nodes  map[S]*node[S, E, C]
	root   *node[S, E, C]
	strict bool
}

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

func (b *Builder[S, E, C]) AddTransition(startState S, event E, endState S) *Builder[S, E, C] {
	start := b.node(startState)
	end := b.node(endState)
	start.neighbors[event] = end
	return b
}

func (b *Builder[S, E, C]) OnEnter(state S, onEnter StatefulFunc[C]) *Builder[S, E, C] {
	n := b.node(state)
	if onEnter != nil {
		n.onEnter = append(n.onEnter, onEnter)
	}
	return b
}

func (b *Builder[S, E, C]) OnExit(state S, onExit StatefulFunc[C]) *Builder[S, E, C] {
	n := b.node(state)
	if onExit != nil {
		n.onExit = append(n.onExit, onExit)
	}
	return b
}

func (b *Builder[S, E, C]) StrictTransitions(strict bool) *Builder[S, E, C] {
	b.strict = strict
	return b
}

func (b *Builder[S, E, C]) Build() *StateMachine[S, E, C] {
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

	return &StateMachine[S, E, C]{
		current: nodes[b.root.state],
		strict:  b.strict,
	}
}
