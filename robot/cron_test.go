package robot

import (
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/db"
)

// InitCron builds the scheduler at the very end of the startup sequence, so a
// cron edit coming from a chat can arrive while Cron is still nil
func TestAddCronWithoutScheduler(t *testing.T) {
	previous := Cron
	Cron = nil
	t.Cleanup(func() { Cron = previous })

	err := AddCron(&db.Cron{CronSpec: "@every 5s"})
	assert.Error(t, err)
}

func TestAddCronRejectsBadSpec(t *testing.T) {
	previous := Cron
	Cron = cron.New(cron.WithSeconds())
	t.Cleanup(func() { Cron = previous })

	err := AddCron(&db.Cron{CronSpec: "not a cron spec"})
	assert.Error(t, err)
}
