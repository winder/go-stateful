// Package stateful is a simple and lightweight event driven state machine.
// With stateful you can implement complex state dependent logic with a clean
// and simple design.
//
// It is a Go port of the Java [stateful] library.
//
// # Overview
//
// A state machine is declared with a [Builder] and then created with
// [Builder.Build]. The builder records three things:
//
//   - transitions, which move the machine from a start state to an end state
//     when an event is applied ([Builder.AddTransition]),
//   - listeners, which are called when a state is entered or exited
//     ([Builder.OnEnter] and [Builder.OnExit]),
//   - whether an event with no matching transition is an error
//     ([Builder.StrictTransitions]).
//
// Events are applied to the resulting [StateMachine] with
// [StateMachine.Apply] or [StateMachine.ApplyWithContext], and the current
// state is read with [StateMachine.State].
//
// # Type parameters
//
// Every type in the package is parameterized by up to three types:
//
//   - S, the state type,
//   - E, the event type,
//   - C, the context type passed to listeners.
//
// S and E must be comparable. They are typically small enumerations declared
// with iota. If they implement [fmt.Stringer] their names appear in error
// messages. C may be any type. Use a pointer to your own struct, a
// [context.Context], or any if your listeners do not need a context.
//
//	type State int
//
//	const (
//		Init State = iota
//		Running
//		Completed
//	)
//
//	type Event int
//
//	const (
//		Run Event = iota
//		End
//	)
//
//	sm := stateful.NewBuilder[State, Event, any](Init).
//		AddTransition(Init, Run, Running).
//		AddTransition(Running, End, Completed).
//		Build()
//
//	sm.State()   // Init
//	sm.Apply(Run)
//	sm.State()   // Running
//
// # Listeners and contexts
//
// A [StatefulFunc] listener may be registered for any state. When an event
// causes a transition, the exit listeners of the current state are called,
// then the machine moves to the new state, then the enter listeners of the new
// state are called. Listeners for a state run in the order they were
// registered.
//
// [StateMachine.ApplyWithContext] passes a context value to every listener
// called by that transition. This keeps listeners decoupled from the code that
// applies events. [StateMachine.Apply] passes the zero value of C.
//
// # Strict transitions
//
// By default, applying an event that has no transition from the current state
// returns an [*UnexpectedEventError]. The error matches [ErrUnexpectedEvent]
// with [errors.Is], and [errors.As] retrieves the state and event. Call
// [Builder.StrictTransitions] with false to ignore such events instead. In
// either case the state does not change and no listeners are called.
//
// # Concurrency
//
// A [StateMachine] is not safe for concurrent use. If events are applied from
// several goroutines, or the state is read while events are applied,
// synchronize access externally. Listeners run synchronously on the goroutine
// that called Apply and should not apply events to the machine that called
// them.
//
// [stateful]: https://github.com/winder/stateful
package stateful
