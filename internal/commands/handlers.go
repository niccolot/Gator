package commands

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"

	"golang.org/x/term"

	"github.com/google/uuid"
	"github.com/niccolot/BlogAggregator/internal/auth"
	"github.com/niccolot/BlogAggregator/internal/database"
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
		return fmt.Errorf("usage: login <username>")
	}

	name := cmd.Args[0]

	if name == s.Cfg.CurrentUserName {
		return fmt.Errorf("user '%s' already logged in", name)
	}

	user, _ := s.Db.GetUser(context.Background(), name)
	if user.Name != name {
		return fmt.Errorf("username '%s' not found", name)
	}

	fmt.Println("Insert password: ")
	pass, err := term.ReadPassword(0)
	if err != nil {
		return fmt.Errorf("failed to read password: %v", err)
	}

	errCheck := auth.CheckPasswordHash(string(pass), user.HashedPassword)
	if errCheck != nil {
		return fmt.Errorf("invalid password")
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
		return fmt.Errorf("usage: register <username>")
	}	

	name := cmd.Args[0]

	fmt.Println("Choose a password: ")
	pass, err := term.ReadPassword(0)
	if err != nil {
		return fmt.Errorf("failed to read password: %v", err)
	}

	fmt.Println("Repeat password: ")
	pass2, err := term.ReadPassword(0)
	if err != nil {
		return fmt.Errorf("failed to read password: %v", err)
	}

	if string(pass) != string(pass2) {
		return fmt.Errorf("enter the same password")
	}

	hashed_password, err := auth.HashPassword(string(pass))
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	pars := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: name,
		HashedPassword: hashed_password,
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
		return fmt.Errorf("usage: users")
	}

	users, errUsers := s.Db.GetUsers(context.Background())
	if errUsers != nil {
		return fmt.Errorf("error while quering database: %v", errUsers)
	}

	for _, user := range users {
		if user.Name == s.Cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

func handlerAggregate(s *state.State, cmd Command, user *database.User) error {
	pars, err := parseAggregationInputs(s, &cmd, user)
	if err != nil {
		return err
	}

	numFeeds := pars.numFollowing
	timeBetweenReqs := pars.timeBetweenReqs
	logging := pars.logging
	
	if logging {
		logFile, err := setLogger("aggregation.log")
		if err != nil {
			return fmt.Errorf("failed to set logger: %v", err)
		}

		defer logFile.Close()

		log.SetOutput(logFile)

		log.Printf("Collecting feeds every %s...\n", timeBetweenReqs)
	}
	
	// aggregation
	ticker := time.NewTicker(timeBetweenReqs)
	defer ticker.Stop()

	s.Aggregating = true

	aggPars := &aggPars{
		s: s,
		timeBetweenReqs: timeBetweenReqs,
		numFeeds: numFeeds,
		logging: logging,
	}

	/*
	every timeBetweenReqs a total of 'workers' goroutines
	are spawned and a total of 'batchSize' feeds are fetched 
	*/
	for range ticker.C {
		select {
		case <-s.StopAggregation:
			fmt.Println("Aggregation stopped")
			return nil
		default:
			aggregate(aggPars)
		}
	} 

	return nil
}

func handlerStopAgg(s *state.State, cmd Command) error {
	if !s.Aggregating {
		fmt.Println("Aggregation already not running")
	} else {
		s.StopAggregation <- true
	}

	return nil
}

func handlerResetUsers(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: resetusers")
	}

	errDelete := s.Db.ResetUsers(context.Background())
	if errDelete != nil {
		return fmt.Errorf("error while resetting database: %v", errDelete)
	}

	fmt.Println("users succesfully reset")

	return nil
}

func handlerResetFeeds(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: resetfeeds")
	}

	errDelete := s.Db.ResetFeeds(context.Background())
	if errDelete != nil {
		return fmt.Errorf("error while resetting database: %v", errDelete)
	}

	fmt.Println("feeds succesfully reset")

	return nil
}

func handlerReset(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: reset")
	}

	errDeleteUsers := s.Db.ResetUsers(context.Background())
	if errDeleteUsers != nil {
		return fmt.Errorf("error while resetting users: %v", errDeleteUsers)
	}

	errDeleteFeeds := s.Db.ResetUsers(context.Background())
	if errDeleteFeeds != nil {
		return fmt.Errorf("error while resetting feeds: %v", errDeleteFeeds)
	}

	fmt.Println("database succesfully reset")

	return nil
}

func handlerAddFeed(s *state.State, cmd Command, user *database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: addfeed \"<feed name>\" \"<feed url>\"")
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

	fmt.Printf("Feed name: %s\n", feed.Name)
	fmt.Println(feed.ID)
	fmt.Printf("Created at: %s\n", feed.CreatedAt)
	fmt.Printf("Updated at; %s\n", feed.UpdatedAt)
	fmt.Printf("URL: %s\n", feed.Url)

	return nil
}

func handlerFeeds(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: feeds")
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
		
		fmt.Println()
		fmt.Printf("Feed name: %s\n", feed.Name)
		fmt.Printf("Feed ID: %s\n", feed.ID)
		fmt.Printf("Created at: %s\n", feed.CreatedAt)
		fmt.Printf("Updated at; %s\n", feed.UpdatedAt)
		fmt.Printf("URL: %s\n", feed.Url)
		fmt.Printf("UserID: %s\n", feed.UserID)
		fmt.Printf("Author name: %s\n", user.Name)
	}

	return nil
}

func handlerFollow(s *state.State, cmd Command, user *database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: follow <feed url>")
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
		return fmt.Errorf("usage: following")
	}
	
	following, errFollowing := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if errFollowing != nil {
		return fmt.Errorf("error while retrieving followed feeds from database: %v", errFollowing)
	}

	for _, feedFollow := range(following) {
		feed, errFeed := s.Db.GetFeedFromID(context.Background(), feedFollow.FeedID)
		if errFeed != nil {
			fmt.Printf("error while retrieving followed feed details: %v", errFeed)
			continue
		}
		fmt.Println()
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(feed.ID)
	}

	return nil
}

func handlerUnfollow(s *state.State, cmd Command, user *database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: unfollow <feed url> [or] unfollow \"<feed name>\"")
	}

	currUserID := user.ID
	unfollowPars := &database.UnfollowParams{
		UserID: currUserID,
		Url: cmd.Args[0], // url or name
	}

	errUnfollow := s.Db.Unfollow(context.Background(), *unfollowPars)
	if errUnfollow != nil {
		return fmt.Errorf("error while removing feed from following list: %v", errUnfollow)
	}

	fmt.Println("feed succesfully unfollowed")

	return nil
}

func handlerBrowse(s *state.State, cmd Command, user *database.User) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("usage: browse [optional] <limit>")
	}

	following, errFollowing := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if errFollowing != nil {
		return fmt.Errorf("error while retrieving followed feeds from database: %v", errFollowing)
	}

	if len(following) == 0 {
		return fmt.Errorf("no feed is being currently followed")
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
		fmt.Println("Published at: ", post.PublishedAt.Time)
		fmt.Println("Link: ", post.Url)
	}

	return nil
}

func handlerOpen(s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: open <post url>")
	}

	_, err := s.Db.GetPostFromUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to find post: %v", err)	
	}

	errOpen := exec.Command("open", cmd.Args[0]).Run()
	if errOpen != nil {
		return fmt.Errorf("error opening url: %v", errOpen)
	}

	fmt.Println("opening post in default browser...")

	return nil
}

