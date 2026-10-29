Stateful
========
<p align="center">
  <a href="https://github.com/winder/go-stateful/actions?query=workflow%3A%22tests+and+coverage%22+event%3Apush">
    <img title="GitHub Workflow Status (test @ main)" src="https://img.shields.io/github/actions/workflow/status/winder/go-stateful/test.yml?branch=main&label=test&style=flat-square">
  </a>
  <a href="https://codecov.io/gh/winder/go-stateful">
    <img title="Code Coverage" src="https://img.shields.io/codecov/c/github/winder/go-stateful/main?style=flat-square">
  </a>
  <a href="https://pkg.go.dev/github.com/winder/go-stateful">
    <img title="Go Documentation" src="https://pkg.go.dev/badge/github.com/winder/go-stateful?style=flat-square">
  </a>
</p>

Stateful is a simple and lightweight event driven state machine library. With
Stateful you can implement complex state dependent logic with a clean and simple
design.

This is a Go port of the Java [stateful](https://github.com/winder/stateful)
library.

## Installation

```
go get -u github.com/winder/go-stateful@latest
```

## Usage

The full API, with runnable examples, is documented on
[pkg.go.dev](https://pkg.go.dev/github.com/winder/go-stateful).

### Simple Transitions

```go
type State int

const (
	Init State = iota
	Running
	Completed
)

type Event int

const (
	Run Event = iota
	End
)

type Context struct {
	Message string
}

...

sm := stateful.NewBuilder[State, Event, *Context](Init).
	AddTransition(Init, Run, Running).
	AddTransition(Running, End, Completed).
	Build()

sm.State() // Init
sm.Apply(Run)
sm.State() // Running
```

### On State Enter/Exit Listeners

```go
sm := stateful.NewBuilder[State, Event, *Context](Init).
	AddTransition(Init, Run, Running).
	OnExit(Init, func(c *Context) { fmt.Println("Exiting Init!") }).
	OnEnter(Running, func(c *Context) { fmt.Println("Entering Running!") }).
	Build()

sm.Apply(Run)
```

In addition to applying simple events, a context may be passed in to allow
further decoupling of application concerns. Listeners receive the zero value of
the context type when `Apply` is used.

```go
sm := stateful.NewBuilder[State, Event, *Context](Init).
	AddTransition(Init, Run, Running).
	OnExit(Init, func(c *Context) { fmt.Println("Exiting Init:", c.Message) }).
	OnEnter(Running, func(c *Context) { fmt.Println("Entering Running:", c.Message) }).
	Build()

sm.ApplyWith(Run, &Context{Message: "Started at: " + time.Now().String()})
```

The context type can be anything, including a `context.Context`. Use `any` if
your listeners do not need one.

### Non-strict transition mode

By default, applying an event which does not cause a state transition returns
an `*UnexpectedEventError`. It matches `stateful.ErrUnexpectedEvent` with
`errors.Is`, and `errors.As` retrieves the state and event:

```go
err := sm.Apply(End)

errors.Is(err, stateful.ErrUnexpectedEvent) // true

var unexpected *stateful.UnexpectedEventError[State, Event]
if errors.As(err, &unexpected) {
	fmt.Println(unexpected.State, unexpected.Event)
}
```

This behavior can be disabled by setting `StrictTransitions` to false when
building the state machine.

```go
sm := stateful.NewBuilder[State, Event, *Context](Init).
	StrictTransitions(false).
	Build()

// No longer returns an error.
sm.Apply(Run)
```

### Concurrency

A `Machine` is not safe for concurrent use. Synchronize access externally
if events are applied from several goroutines.
