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

	feedPars := &database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.Args[0],
		Url: cmd.Args[1],
		UserID: s.Cfg.CurrentUserID,
	}

	feed, errFeed := s.Db.CreateFeed(context.Background(), *feedPars)
	if errFeed != nil {
		return fmt.Errorf("error while creating feed in database: %v", errFeed)
	}

	fmt.Printf("Feed ID: %s\n", feed.ID)
	fmt.Printf("Created at: %s", feed.CreatedAt)
	fmt.Printf("Updated at; %s", feed.UpdatedAt)
	fmt.Printf("Feed name: %s", feed.Name)
	fmt.Printf("URL: %s", feed.Url)
	fmt.Printf("UserID: %s", s.Cfg.CurrentUserID)

	return nil
}

/*
type Feed struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Url       string
	UserID    uuid.UUID
}
*/
