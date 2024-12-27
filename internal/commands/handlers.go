package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/niccolot/BlogAggregator/internal/database"
	"github.com/niccolot/BlogAggregator/internal/rss"
	"github.com/niccolot/BlogAggregator/internal/state"
)

func middlewareLoggedIn(
	handler func(s *state.State, cmd Command, user *database.User) error) func(s *state.State, cmd Command) error {

	return func(s *state.State, cmd Command) error {
		user, err := s.Db.GetUser(context.Background(), s.Cfg.CurrentUserName)
		if err != nil {
			return err
		}

		return handler(s,cmd,&user)
	}
}

func handlerLogin(s *state.State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: gator login <username>")
	}

	name := cmd.Args[0]

	if name == s.Cfg.CurrentUserName {
		return fmt.Errorf("user '%s' already logged in", name)
	}

	user, _ := s.Db.GetUser(context.Background(), name)
	if user.Name != name {
		return fmt.Errorf("username '%s' not found", name)
	}

	errSet := s.Cfg.SetUser(user.Name, user.ID)
	if errSet != nil {
		return fmt.Errorf("error setting username: %v", errSet)
	}

	s.Cfg.CurrentUserID = user.ID
	s.Cfg.CurrentUserName = user.Name
	fmt.Printf("user '%s' succesfully logged in", name)

	return nil
}

func handlerRegister(s *state.State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: gator register <username>")
	}	

	name := cmd.Args[0]
	
	pars := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: name,
	}

	_, errTemp := s.Db.GetUser(context.Background(), name)
	if errTemp == nil {
		return fmt.Errorf("user '%s' already registered", name)
	}

	newUser, errRegister := s.Db.CreateUser(context.Background(), pars)
	if errRegister != nil {
		return fmt.Errorf("error registering user: %v", errRegister)
	}

	s.Cfg.SetUser(newUser.Name, newUser.ID)
	fmt.Printf("user %s succesfully registered", name)

	return nil
}

func handlerGetUsers(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: gator users")
	}

	users, errUsers := s.Db.GetUsers(context.Background())
	if errUsers != nil {
		return fmt.Errorf("error while quering database: %v", errUsers)
	}

	for _, user := range users {
		if user.Name == s.Cfg.CurrentUserName {
			fmt.Printf("* %s (current)", user.Name)
		} else {
			fmt.Printf("* %s", user.Name)
		}
	}

	return nil
}

func handlerAgg(s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		log.Printf("usage: gator agg <time between requests>")
		return nil
	}

	timeBetweenReqs, errParse := time.ParseDuration(cmd.Args[0])
	if errParse != nil {
		return fmt.Errorf("error while parsing fetching frequency: %v", errParse)
	}

	if timeBetweenReqs < time.Second {
		timeBetweenReqs = time.Second
		log.Println("Warning: time between request selected is too small, set to default 1s")
	}

	fmt.Printf("Collecting feeds every %s...\n", timeBetweenReqs)

	batchSize := int32(2)
	workers := 2
	wg := sync.WaitGroup{}	
	queueMux := sync.Mutex{}

	ticker := time.NewTicker(timeBetweenReqs)
	defer ticker.Stop()

	for range ticker.C {
		feedQueue, errScrape := rss.ScrapeFeeds(s, context.Background(), batchSize)
		fmt.Println(len(feedQueue))
		if errScrape != nil {
			log.Printf("Warning: error retrieving feeds: %v" ,errScrape)
			continue
		}

		for i := 0; i<workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
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
	
					ctxWithTimeout, cancel := context.WithTimeout(context.Background(), timeBetweenReqs)
					defer cancel()
	
					startTime := time.Now()
					if feed.Url == "" {
						return
					} else {
						fmt.Printf("fetching feed %s\n", feed.Url)
					}
					
					err := rss.FetchAndStoreFeed(s, &feed, ctxWithTimeout)
	
					if err != nil {
						log.Printf("[Worker %d] Timeout or failed to fetch feed '%s': %v", workerID, feed.Url, err)
					} else {
						log.Printf("[Worker %d] Succesfully fetched feed '%s' in %v", workerID, feed.Url, time.Since(startTime))
					}
				}
			}(i)
		}

		wg.Wait()
	}

	return nil
}

func handlerReset(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: gator reset")
	}

	errDelete := s.Db.Reset(context.Background())
	if errDelete != nil {
		return fmt.Errorf("error while resetting database: %v", errDelete)
	}

	fmt.Println("database succesfully reset")

	return nil
}

func handlerAddFeed(s *state.State, cmd Command, user *database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: gator <feed name> <feed url>")
	}

	currTime := time.Now()
	newFeedID := uuid.New()

	feedPars := &database.CreateFeedParams{
		ID: newFeedID,
		CreatedAt: currTime,
		UpdatedAt: currTime,
		Name: cmd.Args[0],
		Url: cmd.Args[1],
		UserID: user.ID,
	}

	feed, errFeed := s.Db.CreateFeed(context.Background(), *feedPars)
	if errFeed != nil {
		return fmt.Errorf("error while creating feed in database: %v", errFeed)
	}

	feedFollowsPars := &database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: currTime,
		UpdatedAt: currTime,
		UserID: user.ID,
		FeedID: newFeedID,
	}

	_, errFeedFollows := s.Db.CreateFeedFollow(context.Background(), *feedFollowsPars)
	if errFeedFollows != nil {
		return fmt.Errorf("error while updating following list: %v", errFeedFollows)
	}

	fmt.Printf("Feed ID: %s\n", feed.ID)
	fmt.Printf("Created at: %s\n", feed.CreatedAt)
	fmt.Printf("Updated at; %s\n", feed.UpdatedAt)
	fmt.Printf("Feed name: %s\n", feed.Name)
	fmt.Printf("URL: %s\n", feed.Url)
	fmt.Printf("UserID: %s\n", user.ID)

	return nil
}

func handlerFeeds(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: gator feeds")
	}

	feeds, errFeeds := s.Db.GetFeeds(context.Background())
	if errFeeds != nil {
		return fmt.Errorf("error while retrieving feeds from database: %v", errFeeds)
	}

	for _, feed := range(feeds) {
		user, errUser := s.Db.GetuserFromID(context.Background(), feed.UserID)
		if errUser != nil {
			return fmt.Errorf("error while retrieving username from database; %v", errUser)
		}
		
		fmt.Printf("Feed ID: %s\n", feed.ID)
		fmt.Printf("Created at: %s\n", feed.CreatedAt)
		fmt.Printf("Updated at; %s\n", feed.UpdatedAt)
		fmt.Printf("Feed name: %s\n", feed.Name)
		fmt.Printf("URL: %s\n", feed.Url)
		fmt.Printf("UserID: %s\n", feed.UserID)
		fmt.Printf("Author name: %s\n", user.Name)
	}

	return nil
}

func handlerFollow(s *state.State, cmd Command, user *database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: gator follow <feed url>")
	}

	currUserID := user.ID
	feed, errFeed := s.Db.GetFeedFromURL(context.Background(), cmd.Args[0])
	if errFeed != nil {
		return fmt.Errorf("error while retrieving feed from database; %v", errFeed)
	}

	feedFollowPars := &database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: currUserID,
		FeedID: feed.ID,
	}

	_, errFeedFollow := s.Db.CreateFeedFollow(context.Background(), *feedFollowPars)
	if errFeedFollow != nil {
		return fmt.Errorf("error while following feed: %v", errFeedFollow)
	}

	fmt.Printf("Followed feed name: %s\n", feed.Name)
	fmt.Printf("Current user: %s\n", user.Name)

	return nil
}

func handlerFollowing(s *state.State, cmd Command, user *database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: gator following")
	}
	
	following, errFollowing := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if errFollowing != nil {
		return fmt.Errorf("error while retrieving followed feeds from database: %v", errFollowing)
	}

	for _, feedFollow := range(following) {
		feed, errFeed := s.Db.GetFeedFromID(context.Background(), feedFollow.FeedID)
		if errFeed != nil {
			return fmt.Errorf("error while retrieving followed feeds details: %v", errFeed)
		}
		fmt.Println(feed.Name)
	}

	return nil
}

func handlerUnfollow(s *state.State, cmd Command, user *database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: gator unfollow <feed url>")
	}

	currUserID := user.ID
	unfollowPars := &database.UnfollowParams{
		UserID: currUserID,
		Url: cmd.Args[0],
	}

	errUnfollow := s.Db.Unfollow(context.Background(), *unfollowPars)
	if errUnfollow != nil {
		return fmt.Errorf("error while removing feed from following list: %v", errUnfollow)
	}

	return nil
}

func handlerBrowse(s *state.State, cmd Command, user *database.User) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("usage: gator browse [optional] <limit>")
	}

	var limitStr string;
	if len(cmd.Args) == 0 {
		limitStr = "2" // 2 posts as default limit
	} else {
		limitStr = cmd.Args[0]
	}

	limit, errConv := strconv.ParseInt(limitStr, 10, 32)
	if errConv != nil {
		return fmt.Errorf("failed to parse limit value: %v", errConv)
	}

	getPostsPars := &database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: int32(limit), // ParseInt is bugged and always returns int64 regardless of the choice
	}

	posts, errPosts := s.Db.GetPostsForUser(context.Background(), *getPostsPars)
	if errPosts != nil {
		return fmt.Errorf("failed to get posts from database: %v", errPosts)
	}

	for _, post := range(posts) {
		fmt.Println()
		fmt.Println("Feed: ", post.FeedName)
		fmt.Println(post.Title.String)
		if len(post.Description.String) != 0 {
			fmt.Println(post.Description.String)
		}
	}

	return nil
}