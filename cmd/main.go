package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var builtins = map[string]func(args []string) {
	"cd": builtinCd,
	"help": builtinHelp,
	"exit":   builtinExit,
	"pwd": builtinPwd,
	"clear": builtinClear,
	"echo": builtinEcho,
	"export": builtinExport,
	"unset": builtinUnset,
}

func main() {
	
	reader := bufio.NewReader(os.Stdin) // access the input and create a buffered reade around it so i can read a chunk of it at a time
	fmt.Println("Enter your instruction")

	for {
		fmt.Print("The shell is ready")

		input, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		// trim empty space
		input = strings.TrimSpace(input)

		// if the input is empty, go back to prompt
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break // Break out of the infinite loop
		}

		parts := strings.Fields(input)
		command := parts[0]

		if builtinFunc, exists := builtins[command]; exists {
			builtinFunc(parts)
		} else {
			fmt.Printf("Command not found: %s\n", command)
		}
	}
	
}

func builtinExit(args []string) {
	fmt.Println("Goodbye!")
	os.Exit(0)
}

func builtinCd(args []string) {
	// this is the function to change the directory
	var path string

	// if no argument, go to home directory
	if len(args) < 2 {
		homeDir, err := os.UserHomeDir() // go to home directory

		if err != nil {
			fmt.Println(fmt.Errorf("cd: error getting home directory: %w", err))
		}

		path = homeDir
	} else {
		path = args[1]
	}

	// Change directory
	err := os.Chdir(path)

	if err != nil {
		fmt.Println(fmt.Errorf("cd: %w", err))
	}
}

func builtinPwd(args []string) {
	workingDir, err := os.Getwd()

	if err != nil {
		fmt.Println(fmt.Errorf("working directory: %w", err))
	} else {
		fmt.Printf("%s", workingDir)
	}
}

func builtinEcho(args []string) {
	if len(args) > 1 {
		output := strings.Join(args[1:], " ")
		fmt.Println(output)
	} else {
		fmt.Println()
	}
}

func builtinClear(args []string) {
	fmt.Println("\033[H\033[2J")
}

func builtinExport(args []string) {
	if len(args) < 2 {
		fmt.Println("export: usage: export VAR=value")
		return
	}

	assignment := strings.Join(args[1:], " ")
	parts := strings.SplitN(assignment, "=", 2)

	if len(parts) != 2 {
		fmt.Println("export: invalid format, use VAR=value")
		return
	}

	varName := strings.TrimSpace(parts[0])
	varValue := strings.TrimSpace(parts[1])
	
	err := os.Setenv(varName, varValue)
	if err != nil {
		fmt.Println("export:", err)
	}
}

func builtinUnset(args []string) {
	if len(args) < 2 {
		fmt.Println("unset: usage: unset VAR")
		return
	}
	
	varName := args[1]
	err := os.Unsetenv(varName)
	if err != nil {
		fmt.Println("unset:", err)
	}
}

func builtinHelp(args []string) {
	fmt.Println("Available builtin commands:")
	fmt.Println("  cd [directory]     - Change directory")
	fmt.Println("  pwd                - Print working directory")
	fmt.Println("  echo [args...]     - Print arguments")
	fmt.Println("  clear              - Clear the screen")
	fmt.Println("  export VAR=value   - Set environment variable")
	fmt.Println("  unset VAR          - Unset environment variable")
	fmt.Println("  help               - Show this help message")
	fmt.Println("  exit               - Exit the shell")
}