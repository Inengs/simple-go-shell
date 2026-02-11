package util

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"simple_sh/internal/parser"
	"strings"
	"syscall"
)

func ExpandVariables(input string) string {
	var result strings.Builder

	for i := 0; i < len(input); i++ {
		// Handle escaped dollar sign \$
		if i < len(input)-1 && input[i] == '\\' && input[i+1] == '$' {
			result.WriteByte('$')
			i+=1
			continue
		}

		// Check for variable expansion $VAR or ${VAR}
		if input[i] == '$' {
			i++ // move past $

			if i < len(input) && input[i] == '{' {
				i++ // move past {
				varName := extractUntil(input, i, '}')
				
				if varName != "" {
					result.WriteString(os.Getenv(varName))
					i += len(varName) // skip varName and }
				} else {
					result.WriteString("${")
				}
				continue
			}

			 // Handle $VAR syntax (alphanumeric and underscore only)
            varName := extractVarName(input, i)
            if varName != "" {
                result.WriteString(os.Getenv(varName))
                i += len(varName) -1
            } else {
                result.WriteByte('$') // just a lone $
                i--  // ✅ Add this (we already incremented, need to go back 1)
            }
            continue
		}
		        // Regular character
        result.WriteByte(input[i])
        // i++
    }
    
    return result.String()
}

// Extract variable name (letters, digits, underscore)
func extractVarName(input string, start int) string {
    var varName strings.Builder
    
    for i := start; i < len(input); i++ {
        ch := input[i]
        if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || 
           (ch >= '0' && ch <= '9') || ch == '_' {
            varName.WriteByte(ch)
        } else {
            break
        }
    }
    
    return varName.String()
}

// Extract everything until delimiter
func extractUntil(input string, start int, delimiter byte) string {
    for i := start; i < len(input); i++ {
        if input[i] == delimiter {
            return input[start:i]
        }
    }
    return "" // delimiter not found
}


func ExpandTilde(path string) string {
    // Check if path is empty or doesn't start with ~
    if len(path) == 0 || path[0] != '~' {
        return path
    }

    // Just ~ or ~/something
	if len(path) == 1 || path[1] == '/' {
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return path // return original on error
        }

        if len(path) == 1 {
            return homeDir
        }

        return homeDir + path[1:]
    }

    return path
}

func ExpandGlob(pattern string) []string{
    matches, err := filepath.Glob(pattern)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }

    if len(matches) == 0 {
        return []string{pattern}
    }

    return matches
}

func ValidateCommand(input string) error {
    var inSingleQuote, inDoubleQuote bool
    var escaped bool

    for i := 0; i < len(input); i++ {
        ch := input[i] // get the ith element

        if escaped {
            escaped = false
            continue
        }

        if ch == '\\' {
            escaped = true
            continue
        }

        if ch == '\'' && !inDoubleQuote {
            inSingleQuote = !inSingleQuote
        }

        if ch == '"' && !inSingleQuote {
            inDoubleQuote = !inDoubleQuote
        }
    }

    if inSingleQuote || inDoubleQuote {
        return fmt.Errorf("unclosed quote in command")
    }

    // Check for invalid redirections
    if strings.Contains(input, "> ") && len(strings.TrimSpace(strings.Split(input, ">")[1])) == 0 {
        return fmt.Errorf("invalid redirection: missing filename after >")
    }

        // Check for pipe errors (|| at start/end, or ||)
    trimmed := strings.TrimSpace(input)
    if strings.HasPrefix(trimmed, "|") || strings.HasSuffix(trimmed, "|") {
        return fmt.Errorf("invalid pipe syntax")
    }

    return nil // valid
}


func ResolvePath(path string) (string, error) {
    absolutePath, err := filepath.Abs(path) // converts to absolute path
    if err != nil {
        return "", err
    }

    realPath, err := filepath.EvalSymlinks(absolutePath)
    if err != nil {
        return "", err
    }

    return realPath, nil
}

// O_APPEND int = syscall.O_APPEND // append data to the file when writing.
// O_CREATE int = syscall.O_CREAT  // create a new file if none exists.
// O_WRONLY int = syscall.O_WRONLY // open the file write-only.
var history []string // memory list
func SaveToHistory(command string) {
    // Add to memory
    history = append(history, command)

    // Append to file
    f, err := os.OpenFile(".shell_history", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return
    }

    defer f.Close()

    f.WriteString(command + "\n")
}

func LoadHistory() {
    data, err := os.ReadFile(".shell_history")
    if err != nil {
        return
    }
    history = strings.Split(strings.TrimSpace(string(data)), "\n")
}


func SetupSignalHandlers() {
    sigChan := make(chan os.Signal, 1) // Create a channel that can receive OS signals, with a buffer size of 1.
    signal.Notify(sigChan, syscall.SIGINT) // “Tell Go to send the SIGINT operating system signal into sigChan whenever it happens.” SIGINT is the signal sent when you press: Ctrl + C

    go func() {
        for range sigChan {
            fmt.Println("\n^C")  // print newline, continue shell
            // Don't exit - just ignore the signal
        }
    }()
}


func SetupRedirection(cmd *parser.CommandDetails) (*os.File, *os.File, *os.File, error) {
    var stdin, stdout, stderr *os.File

    // Handle input redirection
    if cmd.InputFile != "" {
        f, err := os.Open(cmd.InputFile)
        if err != nil {
            return nil, nil, nil, fmt.Errorf("cannot open input file: %w", err)
        }
        stdin = f
    }

    // Handle output redirection
    if cmd.OutputFile != "" {
        flags := os.O_CREATE | os.O_WRONLY
        if cmd.Append {
            flags |= os.O_APPEND // File operations use bit flags to combine multiple options in one integer.
        } else {
            flags |= os.O_TRUNC
        }

        f, err := os.OpenFile(cmd.OutputFile, flags, 0644)
        if err != nil {
            return nil, nil, nil, fmt.Errorf("cannot open output file: %w", err)
        }
        stdout = f
    }

    // Handle error redirection
    if cmd.ErrorFile != "" {
        f, err := os.OpenFile(cmd.ErrorFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
        if err != nil {
            return nil, nil, nil, fmt.Errorf("cannot open error file: %w", err)
        }
        stderr = f
    }

    return stdin, stdout, stderr, nil
}

func RestoreStandardStreams(originalStdin, originalStdout, originalStderr *os.File) {
    if originalStdin != nil {
        os.Stdin = originalStdin
    }

    if originalStdout != nil {
        os.Stdout = originalStdout
    }

    if originalStderr != nil {
        os.Stderr = originalStderr
    }
}

func SplitCommandLine(input string) []string {
    var commands []string
    var current strings.Builder
    var inQuotes bool

    for _, ch := range input {
        if ch == '"' || ch == '\'' {
            inQuotes = !inQuotes
            current.WriteRune(ch)
            continue
        }

        if ch == '|' && !inQuotes {
            commands = append(commands, strings.TrimSpace(current.String()))
            current.Reset()
            continue
        }

        current.WriteRune(ch)
    }

    if current.Len() > 0 {
        commands = append(commands, strings.TrimSpace(current.String()))
    }

    return commands
}

func RemoveComments(input string) string {
    var inSingleQuote, inDoubleQuote bool //my toggle bool

    for i := 0; i < len(input); i++ {
        ch := input[i] // capture the ith character

        // Track quotes
        if ch == '\'' && !inDoubleQuote { // if the character is ' and not in double quote
            inSingleQuote = !inSingleQuote // toggle the single quote state
        }
        if ch == '"' && !inSingleQuote { // if the character is " and not in single quote
            inDoubleQuote = !inDoubleQuote // toggle the double quote state
        }

        // Remove comment if # is outside quotes
        if ch == '#' && !inSingleQuote && !inDoubleQuote { // if character is # and not in single quote and not in double quote
            return strings.TrimSpace(input[:i]) // capture the string before #
        }
    }

    return input
}

