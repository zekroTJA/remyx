package scheduler

import (
	"github.com/robfig/cron"
	"github.com/zekrotja/remyx/internal/myxer"
	"github.com/zekrotja/rogu/log"
)

func Run(mxr *myxer.Myxer) {
	log := log.Tagged("Scheduler")

	s := cron.New()

	// TODO: make the schedule configurable
	s.AddFunc("0 0 * * *", func() {
		err := mxr.ScheduleSyncs()
		if err != nil {
			log.Error().Err(err).Msg("auto-sync failed")
		}
	})

	s.Start()
}
