package engine

import (
	"fmt"
	"strings"
)

type Command struct {
	Name string
	Text string
}

func ParseCommand(input string) (Command, error) {
	switch input {
	case "h", "j", "k", "l", "w", "b", "0", "$", "gg", "G", "x", "dw", "dd", "yy", "p", "u", "n":
		return Command{Name: input}, nil
	}
	if strings.HasPrefix(input, "/") && len(input) > 1 {
		return Command{Name: "/", Text: input[1:]}, nil
	}
	if (strings.HasPrefix(input, "i") || strings.HasPrefix(input, "a")) && strings.HasSuffix(input, "<Esc>") {
		return Command{Name: input[:1], Text: strings.TrimSuffix(input[1:], "<Esc>")}, nil
	}
	return Command{}, fmt.Errorf("unknown command %q", input)
}

func IsAllowed(command Command, allowed []string) bool {
	for _, candidate := range allowed {
		if command.Name == candidate {
			return true
		}
	}
	return false
}
