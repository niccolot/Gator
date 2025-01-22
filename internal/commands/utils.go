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

func parseAggregationInputs(
	s *state.State, 
	cmd *Command, 
	user *database.User) (timeBetweenReqs time.Duration, err error) {
	
	if len(cmd.Args) != 1 {
		return 0, fmt.Errorf("usage: aggregate <time between requests>")
	}

	following, errFollowing := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if errFollowing != nil {
		return 0, fmt.Errorf("error while retrieving followed feeds from database: %v", errFollowing)
	}

	if len(following) == 0 {
		return 0, fmt.Errorf("no feed is being currently followed")
	}

	timeBetweenReqs, errParse := time.ParseDuration(cmd.Args[0])
	if errParse != nil {
		return 0, fmt.Errorf("error while parsing fetching frequency: %v", errParse)
	}

	if timeBetweenReqs < time.Second {
		timeBetweenReqs = time.Second
		fmt.Println("Warning: time between request selected is too small, set to default 1s")
	}

	return timeBetweenReqs, nil
}