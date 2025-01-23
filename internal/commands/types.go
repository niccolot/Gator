package commands

import (
	"sync"
	"time"

	"github.com/niccolot/BlogAggregator/internal/database"
	"github.com/niccolot/BlogAggregator/internal/state"
)

type aggInitPars struct {
	numFollowing int
	timeBetweenReqs time.Duration
	logging bool
}

type aggPars struct {
	s *state.State
	timeBetweenReqs time.Duration
	numFeeds int
	logging bool
}

type workerPars struct {
	s *state.State
	wg *sync.WaitGroup
	feedQueue *[]database.Feed
	timeBetweenReqs time.Duration
	logging bool
}