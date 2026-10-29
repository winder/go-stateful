package stateful

type StatefulFunc[C any] func(c C)

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

func (n *node[S, E, C]) neighbor(event E) (*node[S, E, C], bool) {
	next, ok := n.neighbors[event]
	return next, ok
}

func (n *node[S, E, C]) enter(c C) {
	for _, f := range n.onEnter {
		f(c)
	}
}

func (n *node[S, E, C]) exit(c C) {
	for _, f := range n.onExit {
		f(c)
	}
}
