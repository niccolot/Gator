package commands

import (
	"fmt"
	"log"

	"github.com/niccolot/BlogAggregator/internal/state"
)

type Command struct {
	CmdName string
	Args []string
	Description string
}

type Commands struct {
	Handlers map[string]func(*state.State, Command) error
}

func Run(cmd Command, cmds *Commands, s *state.State, logger *log.Logger) {
	if cmd.CmdName == "aggregate" {
		if s.Aggregating {
			fmt.Println("Aggregation already running in the background")
		} else {
			go func() {
				errCmd := cmds.Run(s, cmd)
				if errCmd != nil {
					logger.Println("Error in background task:", errCmd)
				}
			}()
			fmt.Println("Running 'aggregate' in the background...")
		}	
	} else {
		errCmd := cmds.Run(s, cmd)
		if errCmd != nil {
			logger.Println(errCmd)
		}
	}
}

func (c *Commands) RegisterCmd(name string, f func(*state.State, Command) error) {
	c.Handlers[name] = f
}

func (c *Commands) Run(s *state.State, cmd Command) error {
	handler, ok := c.Handlers[cmd.CmdName]
	if !ok {
		return fmt.Errorf("command %s not found", cmd.CmdName)
	}

	errCmd := handler(s,cmd)
	if errCmd != nil {
		return fmt.Errorf("error while executing '%s' Command: %v", cmd.CmdName, errCmd)
	}
	
	return nil
}

func (c *Commands) Init() {

	c.Handlers = make(map[string]func(*state.State, Command) error)

	c.RegisterCmd("login", handlerLogin)
	c.RegisterCmd("register", handlerRegister)
	c.RegisterCmd("resetusers", handlerResetUsers)
	c.RegisterCmd("resetfeeds", handlerResetFeeds)
	c.RegisterCmd("reset", handlerReset)
	c.RegisterCmd("users", handlerGetUsers)
	c.RegisterCmd("aggregate", middlewareLoggedIn(handlerAggregate))
	c.RegisterCmd("stopagg", handlerStopAgg)
	c.RegisterCmd("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.RegisterCmd("feeds", handlerFeeds)
	c.RegisterCmd("follow", middlewareLoggedIn(handlerFollow))
	c.RegisterCmd("following", middlewareLoggedIn(handlerFollowing))
	c.RegisterCmd("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.RegisterCmd("browse", middlewareLoggedIn(handlerBrowse))
	c.RegisterCmd("open", handlerOpen)
}