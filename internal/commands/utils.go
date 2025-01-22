package commands

import (
	"regexp"
	"strings"
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