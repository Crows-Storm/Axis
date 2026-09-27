package decorator

import "errors"

type CommandError error

var (
	InvalidCommandError CommandError = errors.New("invalid command")
)

// The InvalidCommand Returns a detailed error of type InvalidCommand
func InvalidCommand(err string) CommandError {
	return errors.New(err)
}
