package robot

import (
	"sync"
	"time"

	"github.com/yincongcyincong/MuseBot/conf"
)

// Reaction statuses for a turn, reported to Reactor implementations.
const (
	ReactionWorking = "working" // turn started
	ReactionOK      = "ok"      // turn finished successfully
	ReactionError   = "error"   // turn failed
)

// typingRefresh beats faster than the shortest platform indicator lifetime
// (Telegram's typing action fades after ~5s).
const typingRefresh = 4500 * time.Millisecond

// Reactor is implemented by platforms that can attach a reaction emoji to
// the user's message for the lifetime of a turn: working when it starts,
// ok or error when it ends. Status maps to a platform emoji that the
// platform is allowed to set (Telegram restricts bots to a whitelist).
type Reactor interface {
	react(status string)
}

// Typer is implemented by platforms that can show a typing indicator while
// the LLM is generating. Implementations must tolerate being called
// repeatedly until the turn ends.
type Typer interface {
	sendTyping()
}

// startTurnSignals reacts on the user's message and starts the typing loop.
// The returned stop function ends the loop and must be called exactly once.
func (r *RobotInfo) startTurnSignals() (stop func()) {
	stopCh := make(chan struct{})

	if re, ok := r.Robot.(Reactor); ok && conf.BaseConfInfo.EnableReaction {
		re.react(ReactionWorking)
	}

	if typer, ok := r.Robot.(Typer); ok && conf.BaseConfInfo.EnableTyping {
		go func() {
			ticker := time.NewTicker(typingRefresh)
			defer ticker.Stop()
			typer.sendTyping()
			for {
				select {
				case <-stopCh:
					return
				case <-r.Ctx.Done():
					return
				case <-ticker.C:
					typer.sendTyping()
				}
			}
		}()
	}

	var once sync.Once
	return func() { once.Do(func() { close(stopCh) }) }
}

// finishTurnSignals stops the typing loop and reports the final reaction.
func (r *RobotInfo) finishTurnSignals(stop func(), failed bool) {
	stop()

	if re, ok := r.Robot.(Reactor); ok && conf.BaseConfInfo.EnableReaction {
		if failed {
			re.react(ReactionError)
		} else {
			re.react(ReactionOK)
		}
	}
}
