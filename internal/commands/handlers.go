package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niccolot/BlogAggregator/internal/database"
	"github.com/niccolot/BlogAggregator/internal/state"
	"github.com/niccolot/BlogAggregator/internal/rss"
)

func handlerLogin(s *state.State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: gator login <username>")
	}

	name := cmd.Args[0]

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
	
	tempName, _ := s.Db.GetUser(context.Background(), name)
	if tempName.Name == name {
		return fmt.Errorf("user already registered")
	}
	
	pars := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: name,
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
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: gator agg")
	}

	feed, errFeed := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if errFeed != nil {
		return fmt.Errorf("error while getting feed: %v", errFeed)
	}

	fmt.Println(feed.Channel.Title)
	fmt.Println(feed.Channel.Link)
	fmt.Println(feed.Channel.Description)
	
	for _, item := range feed.Channel.Item {
		fmt.Println(item.Title)
		fmt.Println(item.Link)
		fmt.Println(item.Description)
		fmt.Println(item.PubDate)
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

func handlerAddFeed(s *state.State, cmd Command) error {
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
		UserID: s.Cfg.CurrentUserID,
	}

	feed, errFeed := s.Db.CreateFeed(context.Background(), *feedPars)
	if errFeed != nil {
		return fmt.Errorf("error while creating feed in database: %v", errFeed)
	}

	feedFollowsPars := &database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: currTime,
		UpdatedAt: currTime,
		UserID: s.Cfg.CurrentUserID,
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
	fmt.Printf("UserID: %s\n", s.Cfg.CurrentUserID)

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

func handlerFollow(s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: gator follow <feed url>")
	}

	currUserID := s.Cfg.CurrentUserID
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
	fmt.Printf("Current user: %s\n", s.Cfg.CurrentUserName)

	return nil
}

func handlerFollowing(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: gator following")
	}
	
	following, errFollowing := s.Db.GetFeedFollowsForUser(context.Background(), s.Cfg.CurrentUserID)
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