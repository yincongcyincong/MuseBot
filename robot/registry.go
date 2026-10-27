package robot

import (
	"context"

	"github.com/yincongcyincong/MuseBot/logger"
)

// Channel is one chat platform: how to tell it is configured and how to
// start it. Platforms self-register from their own file's init, so adding
// one means adding a file instead of touching StartRobot.
type Channel struct {
	Name string
	// MissingCfg returns the names of the config keys that are absent;
	// an empty result means the channel can start.
	MissingCfg func() []string
	Start      func(ctx context.Context)
}

var channels []Channel

// RegisterChannel appends a channel to the startup registry.
func RegisterChannel(c Channel) {
	if c.Name == "" {
		panic("robot: RegisterChannel called with empty name")
	}
	if c.MissingCfg == nil || c.Start == nil {
		panic("robot: RegisterChannel " + c.Name + " needs MissingCfg and Start")
	}
	channels = append(channels, c)
}

// required appends key to missing when value is absent.
func required(missing []string, key, value string) []string {
	if value == "" {
		return append(missing, key)
	}
	return missing
}

// startChannels starts every configured channel and logs the others with
// the keys they are missing (validate without executing platform code).
func startChannels(ctx context.Context) {
	for _, ch := range channels {
		missing := ch.MissingCfg()
		if len(missing) > 0 {
			logger.InfoCtx(ctx, "channel disabled, missing config", "channel", ch.Name, "missing", missing)
			continue
		}
		go ch.Start(ctx)
	}
}
