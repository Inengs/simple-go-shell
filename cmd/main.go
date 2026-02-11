package main

import (
	"bufio"
	"fmt"
	"os"
	"simple_sh/internal/jobs"
	"simple_sh/internal/parser"
	"simple_sh/internal/util"
	"strings"
)

var builtins = map[string]func(args []string) {
	"cd":     builtinCd,
	"help":   builtinHelp,
	"exit":   builtinExit,
	"pwd":    builtinPwd,
	"clear":  builtinClear,
	"echo":   builtinEcho,
	"export": builtinExport,
	"unset":  builtinUnset,
	"jobs":   builtinJobs,
}

func main() {
	// Setup signal handlers and load history
	util.SetupSignalHandlers()
	util.LoadHistory()
	
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintln(os.Stderr, "Welcome to Simple Shell!")
	fmt.Fprintln(os.Stderr, "Type 'help' for available commands")

	for {
		fmt.Fprint(os.Stderr, "shell> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		// Remove comments
		input = util.RemoveComments(input)
		if input == "" {
			continue
		}

		// Validate command
		if err := util.ValidateCommand(input); err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Expand variables
		input = util.ExpandVariables(input)

		// Save to history
		util.SaveToHistory(input)

		// Parse the command
		cmd, err := parser.Parse(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Parse error:", err)
			continue
		}

		// Handle redirection
		stdin, stdout, stderr, err := util.SetupRedirection(cmd)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Redirection error:", err)
			continue
		}

		// Save original streams
		oldStdin := os.Stdin
		oldStdout := os.Stdout
		oldStderr := os.Stderr

		// Apply redirections BEFORE executing
		if stdin != nil {
			os.Stdin = stdin
		}
		if stdout != nil {
			os.Stdout = stdout
		}
		if stderr != nil {
			os.Stderr = stderr
		}

		// Check if it's a builtin
		if builtinFunc, exists := builtins[cmd.Args[0]]; exists {
			builtinFunc(cmd.Args)
		} else {
			// Execute external command
			err = jobs.ExecuteCommandWithJobs(cmd.Args)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Execution error:", err)
			}
		}

		// Restore streams IMMEDIATELY
		util.RestoreStandardStreams(oldStdin, oldStdout, oldStderr)

		// Close files
		if stdin != nil {
			stdin.Close()
		}
		if stdout != nil {
			stdout.Close()
		}
		if stderr != nil {
			stderr.Close()
		}

		// Clean up finished jobs
		jobs.RemoveCompletedJobs()
	}
}

func builtinJobs(args []string) {
	jobs.ListJobs()
}

func builtinExit(args []string) {
	fmt.Println("Goodbye!")
	os.Exit(0)
}

func builtinCd(args []string) {
	var path string

	if len(args) < 2 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("cd: error getting home directory:", err)
			return
		}
		path = homeDir
	} else {
		// Expand tilde
		path = util.ExpandTilde(args[1])
	}

	err := os.Chdir(path)
	if err != nil {
		fmt.Println("cd:", err)
	}
}

func builtinPwd(args []string) {
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println("pwd:", err)
	} else {
		fmt.Println(workingDir)
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
	fmt.Print("\033[H\033[2J")
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
	fmt.Println("  jobs               - List background jobs")
	fmt.Println("  help               - Show this help message")
	fmt.Println("  exit               - Exit the shell")
}