package stateful_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/winder/go-stateful"
)

func ExampleNewBuilder() {
	// A turnstile is locked until a coin is inserted, and locks again after
	// someone pushes through. The listeners do not need a context, so any is
	// used for the context type.
	sm := stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked").
		AddTransition("unlocked", "push", "locked").
		Build()

	fmt.Println(sm.State())
	// Output: locked
}

func ExampleBuilder_AddTransition() {
	sm := stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked").
		AddTransition("unlocked", "push", "locked").
		// Inserting another coin while unlocked keeps the turnstile unlocked.
		AddTransition("unlocked", "coin", "unlocked").
		Build()

	for _, event := range []string{"coin", "coin", "push"} {
		_ = sm.Apply(event)
		fmt.Println(event, "->", sm.State())
	}
	// Output:
	// coin -> unlocked
	// coin -> unlocked
	// push -> locked
}

func ExampleBuilder_OnEnter() {
	sm := stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked").
		OnEnter("unlocked", func(any) { fmt.Println("Entering unlocked") }).
		OnEnter("unlocked", func(any) { fmt.Println("Please push through") }).
		Build()

	_ = sm.Apply("coin")
	// Output:
	// Entering unlocked
	// Please push through
}

func ExampleBuilder_OnExit() {
	sm := stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked").
		OnExit("locked", func(any) { fmt.Println("Exiting locked") }).
		OnEnter("unlocked", func(any) { fmt.Println("Entering unlocked") }).
		Build()

	_ = sm.Apply("coin")
	// Output:
	// Exiting locked
	// Entering unlocked
}

func ExampleBuilder_StrictTransitions() {
	strict := stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked").
		Build()

	lenient := stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked").
		StrictTransitions(false).
		Build()

	fmt.Println("strict:", strict.Apply("push"))
	fmt.Println("lenient:", lenient.Apply("push"))
	// Output:
	// strict: an unexpected event of type push happened during the locked state
	// lenient: <nil>
}

func ExampleBuilder_Build() {
	b := stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked")

	// Each machine built from the same builder has its own state.
	first := b.Build()
	second := b.Build()

	_ = first.Apply("coin")
	fmt.Println(first.State(), second.State())
	// Output: unlocked locked
}

func ExampleMachine_Apply() {
	sm := stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked").
		Build()

	if err := sm.Apply("coin"); err != nil {
		fmt.Println(err)
	}
	fmt.Println(sm.State())

	// There is no "coin" transition from unlocked.
	if err := sm.Apply("coin"); err != nil {
		fmt.Println(err)
	}
	fmt.Println(sm.State())
	// Output:
	// unlocked
	// an unexpected event of type coin happened during the unlocked state
	// unlocked
}

func ExampleMachine_ApplyWith() {
	type payment struct {
		cents int
	}

	sm := stateful.NewBuilder[string, string, *payment]("locked").
		AddTransition("locked", "coin", "unlocked").
		OnEnter("unlocked", func(p *payment) {
			fmt.Printf("Unlocked after a payment of %d cents\n", p.cents)
		}).
		Build()

	_ = sm.ApplyWith("coin", &payment{cents: 25})
	// Output: Unlocked after a payment of 25 cents
}

func ExampleMachine_ApplyWith_contextContext() {
	type requestIDKey struct{}

	// A context.Context works well as the context type.
	sm := stateful.NewBuilder[string, string, context.Context]("locked").
		AddTransition("locked", "coin", "unlocked").
		OnEnter("unlocked", func(ctx context.Context) {
			fmt.Println("Unlocked by request", ctx.Value(requestIDKey{}))
		}).
		Build()

	ctx := context.WithValue(context.Background(), requestIDKey{}, "abc123")
	_ = sm.ApplyWith("coin", ctx)
	// Output: Unlocked by request abc123
}

func ExampleMachine_State() {
	var sm *stateful.Machine[string, string, any]

	// Exit listeners see the old state, enter listeners see the new state.
	sm = stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked").
		OnExit("locked", func(any) { fmt.Println("exiting:", sm.State()) }).
		OnEnter("unlocked", func(any) { fmt.Println("entering:", sm.State()) }).
		Build()

	fmt.Println("before:", sm.State())
	_ = sm.Apply("coin")
	fmt.Println("after:", sm.State())
	// Output:
	// before: locked
	// exiting: locked
	// entering: unlocked
	// after: unlocked
}

func ExampleListener() {
	// A listener factory which logs state changes.
	logState := func(action, state string) stateful.Listener[any] {
		return func(any) { fmt.Println(action, state) }
	}

	sm := stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked").
		OnExit("locked", logState("exit", "locked")).
		OnEnter("unlocked", logState("enter", "unlocked")).
		Build()

	_ = sm.Apply("coin")
	// Output:
	// exit locked
	// enter unlocked
}

func ExampleUnexpectedEventError() {
	sm := stateful.NewBuilder[string, string, any]("locked").Build()

	err := sm.Apply("push")

	// Check for any unexpected event.
	fmt.Println(errors.Is(err, stateful.ErrUnexpectedEvent))

	// Retrieve the state and event.
	var unexpected *stateful.UnexpectedEventError[string, string]
	if errors.As(err, &unexpected) {
		fmt.Println(unexpected.State, unexpected.Event)
	}
	// Output:
	// true
	// locked push
}
