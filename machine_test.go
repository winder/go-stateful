package stateful_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/winder/go-stateful"
)

type state int

const (
	initState state = iota
	running
	paused
	completed
)

func (s state) String() string {
	switch s {
	case initState:
		return "INIT"
	case running:
		return "RUNNING"
	case paused:
		return "PAUSED"
	case completed:
		return "COMPLETED"
	}
	return "UNKNOWN"
}

type event int

const (
	run event = iota
	pause
	end
)

func (e event) String() string {
	switch e {
	case run:
		return "RUN"
	case pause:
		return "PAUSE"
	case end:
		return "END"
	}
	return "UNKNOWN"
}

type testContext struct {
	message string
}

type builder = stateful.Builder[state, event, *testContext]

func newBuilder() *builder {
	return stateful.NewBuilder[state, event, *testContext](initState)
}

func TestMachineTransitions(t *testing.T) {
	sm := newBuilder().
		AddTransition(initState, run, running).
		AddTransition(running, pause, paused).
		AddTransition(running, end, completed).
		AddTransition(paused, run, running).
		Build()

	assert.Equal(t, initState, sm.State())

	steps := []struct {
		event event
		want  state
	}{
		{run, running},
		{pause, paused},
		{run, running},
		{end, completed},
	}
	for _, step := range steps {
		require.NoError(t, sm.Apply(step.event))
		assert.Equal(t, step.want, sm.State(), "after %v", step.event)
	}
}

func TestOnStateEnterListener(t *testing.T) {
	entered := false
	sm := newBuilder().
		AddTransition(initState, run, running).
		OnEnter(running, func(*testContext) { entered = true }).
		Build()

	require.NoError(t, sm.Apply(run))
	assert.True(t, entered)
}

func TestOnStateExitListener(t *testing.T) {
	exited := false
	sm := newBuilder().
		AddTransition(initState, run, running).
		OnExit(initState, func(*testContext) { exited = true }).
		Build()

	require.NoError(t, sm.Apply(run))
	assert.True(t, exited)
}

func TestUnexpectedEventErrorIsReturned(t *testing.T) {
	sm := newBuilder().Build()

	err := sm.Apply(run)
	require.Error(t, err)
	assert.ErrorIs(t, err, stateful.ErrUnexpectedEvent)

	var unexpected *stateful.UnexpectedEventError[state, event]
	require.ErrorAs(t, err, &unexpected)
	assert.Equal(t, initState, unexpected.State)
	assert.Equal(t, run, unexpected.Event)
	assert.Equal(t, "an unexpected event of type RUN happened during the INIT state", err.Error())

	// The failed event does not change state.
	assert.Equal(t, initState, sm.State())
}

func TestUnexpectedEventErrorIsNotReturned(t *testing.T) {
	sm := newBuilder().
		StrictTransitions(false).
		Build()

	require.NoError(t, sm.Apply(run))
	assert.Equal(t, initState, sm.State())
}

func TestStrictTransitionsCanBeReenabled(t *testing.T) {
	sm := newBuilder().
		StrictTransitions(false).
		StrictTransitions(true).
		Build()

	assert.ErrorIs(t, sm.Apply(run), stateful.ErrUnexpectedEvent)
}

func TestUnexpectedEventDoesNotCallListeners(t *testing.T) {
	called := false
	listener := func(*testContext) { called = true }

	for _, strict := range []bool{true, false} {
		sm := newBuilder().
			OnExit(initState, listener).
			OnEnter(initState, listener).
			StrictTransitions(strict).
			Build()

		_ = sm.Apply(run)
		assert.False(t, called, "strict=%v", strict)
	}
}

func TestApplyWith(t *testing.T) {
	const expectedMessage = "message"
	var exitMessage, enterMessage string

	sm := newBuilder().
		AddTransition(initState, run, running).
		OnExit(initState, func(c *testContext) { exitMessage = c.message }).
		OnEnter(running, func(c *testContext) { enterMessage = c.message }).
		Build()

	require.NoError(t, sm.ApplyWith(run, &testContext{message: expectedMessage}))
	assert.Equal(t, expectedMessage, exitMessage)
	assert.Equal(t, expectedMessage, enterMessage)
}

func TestApplyPassesZeroContext(t *testing.T) {
	calls := 0
	sm := newBuilder().
		AddTransition(initState, run, running).
		OnExit(initState, func(c *testContext) {
			calls++
			assert.Nil(t, c)
		}).
		Build()

	require.NoError(t, sm.Apply(run))
	assert.Equal(t, 1, calls)
}

func TestValueContext(t *testing.T) {
	var got int
	sm := stateful.NewBuilder[state, event, int](initState).
		AddTransition(initState, run, running).
		OnEnter(running, func(c int) { got = c }).
		Build()

	require.NoError(t, sm.ApplyWith(run, 42))
	assert.Equal(t, 42, got)
}

func TestContextContext(t *testing.T) {
	type key struct{}
	var got any
	sm := stateful.NewBuilder[state, event, context.Context](initState).
		AddTransition(initState, run, running).
		OnEnter(running, func(ctx context.Context) { got = ctx.Value(key{}) }).
		Build()

	ctx := context.WithValue(context.Background(), key{}, "value")
	require.NoError(t, sm.ApplyWith(run, ctx))
	assert.Equal(t, "value", got)
}

func TestListenerOrder(t *testing.T) {
	var calls []string
	record := func(name string) stateful.Listener[*testContext] {
		return func(*testContext) { calls = append(calls, name) }
	}

	sm := newBuilder().
		OnEnter(running, record("enter running 1")).
		OnExit(initState, record("exit init 1")).
		AddTransition(initState, run, running).
		OnEnter(running, record("enter running 2")).
		OnExit(initState, record("exit init 2")).
		Build()

	require.NoError(t, sm.Apply(run))
	assert.Equal(t, []string{
		"exit init 1",
		"exit init 2",
		"enter running 1",
		"enter running 2",
	}, calls)
}

func TestStateDuringListeners(t *testing.T) {
	var sm *stateful.Machine[state, event, *testContext]
	var duringExit, duringEnter state

	sm = newBuilder().
		AddTransition(initState, run, running).
		OnExit(initState, func(*testContext) { duringExit = sm.State() }).
		OnEnter(running, func(*testContext) { duringEnter = sm.State() }).
		Build()

	require.NoError(t, sm.Apply(run))
	assert.Equal(t, initState, duringExit)
	assert.Equal(t, running, duringEnter)
}

func TestInitialStateEnterNotCalledByBuild(t *testing.T) {
	entered := false
	sm := newBuilder().
		OnEnter(initState, func(*testContext) { entered = true }).
		Build()

	assert.False(t, entered)
	assert.Equal(t, initState, sm.State())
}

func TestSelfTransition(t *testing.T) {
	var calls []string
	sm := newBuilder().
		AddTransition(initState, run, running).
		AddTransition(running, run, running).
		OnExit(running, func(*testContext) { calls = append(calls, "exit") }).
		OnEnter(running, func(*testContext) { calls = append(calls, "enter") }).
		Build()

	require.NoError(t, sm.Apply(run))
	require.NoError(t, sm.Apply(run))
	assert.Equal(t, running, sm.State())
	assert.Equal(t, []string{"enter", "exit", "enter"}, calls)
}

func TestAddTransitionReplacesExisting(t *testing.T) {
	sm := newBuilder().
		AddTransition(initState, run, running).
		AddTransition(initState, run, completed).
		Build()

	require.NoError(t, sm.Apply(run))
	assert.Equal(t, completed, sm.State())
}

func TestNilListenerIgnored(t *testing.T) {
	sm := newBuilder().
		AddTransition(initState, run, running).
		OnExit(initState, nil).
		OnEnter(running, nil).
		Build()

	require.NotPanics(t, func() { require.NoError(t, sm.Apply(run)) })
	assert.Equal(t, running, sm.State())
}

func TestListenerOnlyStatesAreNotTransitions(t *testing.T) {
	// Registering a listener for a state does not make it reachable.
	sm := newBuilder().
		OnEnter(running, func(*testContext) {}).
		Build()

	assert.ErrorIs(t, sm.Apply(run), stateful.ErrUnexpectedEvent)
}

func TestBuildIsIndependentOfLaterBuilderChanges(t *testing.T) {
	entered := false
	b := newBuilder().AddTransition(initState, run, running)
	sm := b.Build()

	b.AddTransition(running, end, completed).
		OnEnter(running, func(*testContext) { entered = true }).
		StrictTransitions(false)

	require.NoError(t, sm.Apply(run))
	assert.False(t, entered, "listener added after Build must not be called")
	assert.ErrorIs(t, sm.Apply(end), stateful.ErrUnexpectedEvent)

	sm2 := b.Build()
	require.NoError(t, sm2.Apply(run))
	assert.True(t, entered)
	require.NoError(t, sm2.Apply(end))
	assert.Equal(t, completed, sm2.State())
}

func TestMachinesFromSameBuilderAreIndependent(t *testing.T) {
	b := newBuilder().AddTransition(initState, run, running)
	sm1 := b.Build()
	sm2 := b.Build()

	require.NoError(t, sm1.Apply(run))
	assert.Equal(t, running, sm1.State())
	assert.Equal(t, initState, sm2.State())
}

func TestStringTypes(t *testing.T) {
	sm := stateful.NewBuilder[string, string, any]("locked").
		AddTransition("locked", "coin", "unlocked").
		AddTransition("unlocked", "push", "locked").
		Build()

	require.NoError(t, sm.Apply("coin"))
	assert.Equal(t, "unlocked", sm.State())
	require.NoError(t, sm.Apply("push"))
	assert.Equal(t, "locked", sm.State())

	err := sm.Apply("push")
	assert.EqualError(t, err, "an unexpected event of type push happened during the locked state")
}

func TestUnexpectedEventErrorIsOnlyMatchesSentinel(t *testing.T) {
	err := &stateful.UnexpectedEventError[state, event]{State: initState, Event: run}
	assert.ErrorIs(t, err, stateful.ErrUnexpectedEvent)
	assert.NotErrorIs(t, err, errors.New("unexpected event"))
}
