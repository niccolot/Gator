package commands

import "strings"

func ParseInput(text string) (command string, args []string) {
	/*
	* @param text (string): the whole input from terminal
	*
	* @return command, args (string, []string): the command name and optional arguments slice 
	*/

	// removes trailing whitespaces and lowercases the command
	trimmedText := strings.TrimSpace(text)
	lowercasedText := strings.ToLower(trimmedText)
	
	parts := strings.Split(lowercasedText, " ")

	command = parts[0]
	args = parts[1:]

	return command, args
}