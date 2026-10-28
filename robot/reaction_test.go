package robot

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
)

// reactionTestRobot embeds Web for the Robot interface and records every
// turn status it receives.
type reactionTestRobot struct {
	Web
	statuses chan string
}

func (r *reactionTestRobot) react(status string) { r.statuses <- status }

func (r *reactionTestRobot) sendTyping() {}

func newReactionTestRobot() (*reactionTestRobot, *RobotInfo) {
	test := &reactionTestRobot{statuses: make(chan string, 4)}
	info := NewRobot(WithRobot(test))
	return test, info
}

func setTurnSignals(reaction, typing bool) (restore func()) {
	previousR, previousT := conf.BaseConfInfo.EnableReaction, conf.BaseConfInfo.EnableTyping
	conf.BaseConfInfo.EnableReaction, conf.BaseConfInfo.EnableTyping = reaction, typing
	return func() {
		conf.BaseConfInfo.EnableReaction, conf.BaseConfInfo.EnableTyping = previousR, previousT
	}
}

func TestTurnSignalsReportWorkingThenFinalStatus(t *testing.T) {
	defer setTurnSignals(true, false)()

	test, info := newReactionTestRobot()

	stop := info.startTurnSignals()
	info.finishTurnSignals(stop, false)

	assert.Equal(t, ReactionWorking, <-test.statuses)
	assert.Equal(t, ReactionOK, <-test.statuses)
}

func TestTurnSignalsReportErrorStatus(t *testing.T) {
	defer setTurnSignals(true, false)()

	test, info := newReactionTestRobot()

	stop := info.startTurnSignals()
	info.finishTurnSignals(stop, true)

	assert.Equal(t, ReactionWorking, <-test.statuses)
	assert.Equal(t, ReactionError, <-test.statuses)
}

// platforms without the capabilities must not break the turn flow
func TestTurnSignalsWithoutCapabilities(t *testing.T) {
	defer setTurnSignals(true, false)()

	info := NewRobot(WithRobot(&Web{}))

	stop := info.startTurnSignals()
	finishDone := make(chan struct{})
	go func() {
		info.finishTurnSignals(stop, false)
		close(finishDone)
	}()

	select {
	case <-finishDone:
	case <-time.After(time.Second):
		t.Fatal("finishTurnSignals blocked on a robot without capabilities")
	}
}

func TestTurnSignalsDisabled(t *testing.T) {
	defer setTurnSignals(false, false)()

	test, info := newReactionTestRobot()

	stop := info.startTurnSignals()
	info.finishTurnSignals(stop, false)

	select {
	case status := <-test.statuses:
		t.Fatalf("expected no reaction, got %q", status)
	case <-time.After(100 * time.Millisecond):
	}
}

// the typing loop must exit when the stop function fires
func TestTurnSignalsTypingStops(t *testing.T) {
	defer setTurnSignals(false, true)()

	typingDone := make(chan struct{})
	test := &reactionTestRobot{statuses: make(chan string, 4)}
	test.Web.Robot = nil // unused; only the Typer path runs
	info := NewRobot(WithRobot(&typingStub{reactionTestRobot: test, done: typingDone}))

	stop := info.startTurnSignals()
	stop()

	select {
	case <-typingDone:
	case <-time.After(time.Second):
		t.Fatal("typing loop did not stop")
	}
}

// typingStub reports the first typing call so the test can observe the loop.
type typingStub struct {
	*reactionTestRobot
	done chan struct{}
}

func (t *typingStub) sendTyping() {
	select {
	case <-t.done:
	default:
		close(t.done)
	}
}
