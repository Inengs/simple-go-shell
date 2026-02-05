package executor

import (
	"fmt"     //for creating error messages
	"os"      // for accessing standard input/output/error streams
	"os/exec" // for running external commands
)

func ExecuteCommand(args []string) error{ // Function that takes a slice of strings (command + arguments)
	if len(args) == 0 { // if the slice is empty, return an error
        return fmt.Errorf("no command provided")
    }
	command := args[0] // Extract the command name from the first element
	path, err := exec.LookPath(command) // Searches your system's PATH for the executable

	if err != nil { // If the command wasn't found, return the error to the caller
		return err
	}

	cmd := exec.Command(path, args[1:]...) // Creates a new command object using the full path
	cmd.Stdout = os.Stdout // Connects the command's output to your shell's output
	cmd.Stderr = os.Stderr // Connects the command's error output to your shell's error output
	cmd.Stdin = os.Stdin  // Connects the command's input to your shell's input

	return cmd.Run() // Starts the command, waits for it to finish, and returns any error
}