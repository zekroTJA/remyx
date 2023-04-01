package scheduler

import (
	"github.com/robfig/cron"
	"github.com/zekrotja/remyx/internal/config"
	"github.com/zekrotja/remyx/internal/myxer"
	"github.com/zekrotja/rogu/log"
)

func Run(mxr *myxer.Myxer, cfg config.Config) error {
	log := log.Tagged("Scheduler")

	s := cron.New()

	// TODO: make the schedule configurable
	err := s.AddFunc(cfg.SyncSchedule, func() {
		err := mxr.ScheduleSyncs()
		if err != nil {
			log.Error().Err(err).Msg("auto-sync failed")
		}
	})
	if err != nil {
		return err
	}

	s.Start()

	return nil
}
