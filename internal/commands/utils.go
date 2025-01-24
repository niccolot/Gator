package commands

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

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

func ParseInput(text string) (command string, args []string) {
	/*
	* @brief parses the command name and the arguments. Preserves text 
	* surrounded by double quotes "" as a single argument
	*
	* @param text (string): the whole input from terminal
	*
	* @return command, args (string, []string): the command name and optional arguments slice 
	*/

	// removes trailing whitespaces and lowercases the command
	trimmedText := strings.TrimSpace(text)
	
	re := regexp.MustCompile(`"([^"]*)"|\S+`)
	matches := re.FindAllString(trimmedText, -1)

	var parts []string
	for _, match := range matches {
		if len(match) > 1 && match[0] == '"' && match[len(match)-1] == '"' {
			parts = append(parts, match[1:len(match)-1])
		} else {
			parts = append(parts, match)
		}
	}

	command = parts[0]
	args = parts[1:]

	return strings.ToLower(command), args
}

func setLogger(filename string) (*os.File, error) {
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	stats, err := logFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get logfile stats: %v", err)
	}

	maxFileSize := int64(2e+6) // 2MB max size
	if stats.Size() > maxFileSize {
		fmt.Println("Already existing logfile is too big, creating a new one...")
		newName := "new_" + filename
		file, err := os.OpenFile(newName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to create new file: %v", err)
		}
		return file, nil
	} else {
		return logFile, nil
	}
}

func parseAggregationInputs(s *state.State, cmd *Command, user *database.User) (pars aggInitPars, err error) {
	if len(cmd.Args) < 1 {
		return aggInitPars{}, fmt.Errorf("usage: aggregate <time between requests> [optional] -log")
	}

	var log bool
	if len(cmd.Args) == 2 {
		if cmd.Args[1] != "-log" {
			return aggInitPars{}, fmt.Errorf("usage: aggregate <time between requests> [optional] -log")
		} else {
			log = true
		}
	}

	following, errFollowing := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if errFollowing != nil {
		return aggInitPars{}, fmt.Errorf("error while retrieving followed feeds from database: %v", errFollowing)
	}

	numFollowing := len(following)
	if numFollowing == 0 {
		return aggInitPars{}, fmt.Errorf("no feed is being currently followed")
	}

	timeBetweenReqs, errParse := time.ParseDuration(cmd.Args[0])
	if errParse != nil {
		return aggInitPars{}, fmt.Errorf("error while parsing fetching frequency: %v", errParse)
	}

	if timeBetweenReqs < time.Second {
		timeBetweenReqs = time.Second
		fmt.Println("Warning: time between request selected is too small, set to default 1s")
	}

	pars = aggInitPars{
		numFollowing: numFollowing,
		timeBetweenReqs: timeBetweenReqs,
		logging: log,
	}


	return pars, nil
}