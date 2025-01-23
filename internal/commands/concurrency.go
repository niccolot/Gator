package commands

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/niccolot/BlogAggregator/internal/database"
	"github.com/niccolot/BlogAggregator/internal/rss"
)

func aggregate(pars *aggPars) error {
	batchSize := int32(pars.numFeeds)
	workers := pars.numFeeds
	wg := sync.WaitGroup{}
	
	feedQueue, errScrape := rss.ScrapeFeeds(pars.s, context.Background(), batchSize)
	if errScrape != nil {
		if pars.logging {
			log.Printf("Warning: error retrieving feeds: %v", errScrape)
		}
		return errScrape
	}

	workerPars := &workerPars{
		s: pars.s,
		wg: &wg,
		feedQueue: &feedQueue,
		timeBetweenReqs: pars.timeBetweenReqs,
		logging: pars.logging,
	}

	for i := 0; i<workers; i++ {
		wg.Add(1)
		go workerFunc(i, workerPars)
	}

	wg.Wait()

	return nil
}

func workerFunc(workerID int, pars *workerPars) {
	queueMux := sync.Mutex{}

	feedQueue := *pars.feedQueue

	defer pars.wg.Done()

	for {
		queueMux.Lock()
		var feed database.Feed
		if len(feedQueue) > 0 {
			feed = feedQueue[0]
			feedQueue = feedQueue[1:]
		}
		queueMux.Unlock()

		// no more feeds in the queue
		if feed.ID.String() == "" {
			return
		}

		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), pars.timeBetweenReqs)
		defer cancel()
		
		if feed.Url == "" {
			return
		}
		
		if pars.logging {
			startTime := time.Now()
			err := rss.FetchAndStoreFeed(pars.s, &feed, ctxWithTimeout)
			
			if err != nil {
				log.Printf("[Worker %d] Timeout or failed to fetch feed '%s': %v", workerID, feed.Url, err)
			} else {
				log.Printf("[Worker %d] Succesfully fetched feed '%s' in %v", workerID, feed.Url, time.Since(startTime))
			}
		} else {
			rss.FetchAndStoreFeed(pars.s, &feed, ctxWithTimeout)
		}
		
	}
}