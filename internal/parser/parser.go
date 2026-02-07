package parser

import (
	"fmt"
	"strings"
)

type CommandDetails struct{
	CommandString string
	Args []string
	InputFile string
	OutputFile string 
	Append bool
	ErrorFile string
	Background bool
}

// eg CommandString: ls -l >> out.txt &
// Expected result: 
// Args: ["ls", "-l"]
// OutputFile: out.txt
// Append: true
// Background: true

// eg 2 CommandString: cat < in.txt 2> err.txt
// Args: ["cat"]
// InputFile: in.txt
// ErrorFile: err.txt

func Parse(line string) (*Result, error){
	return
}

func tokenize(input string) ([]string, error) { // this is to produce []string where each element is one argument/operator/filename
	var tokens []string
	var current strings.Builder // used to build strings by appending data without creating many temporary string objects
	var inSingleQuotes, inDoubleQuotes bool
	var backlash bool

	inSingleQuotes = false
	inDoubleQuotes = false 
	backlash = false

	var ch1 rune = '"'
	var ch2 rune = '\''

	for _, character := range input{
		// if the previous character was a backlash, treat this character as normal data
		if backlash {
			current.WriteRune(character)
			backlash = false
			continue
		} 
		
		// if the current character is a backlash, escape the next character
		if character == '\\' {
			backlash = true
			continue
		}
		
		// if we are in single quotes
		if inSingleQuotes {
			if character == ch2 { // closing single quote
				inSingleQuotes = false 
			} else {
				current.WriteRune(character) // everything else is literal
			}
			continue
		}

		if inDoubleQuotes {
			if character == ch1 { // closing double quote
				inDoubleQuotes = false  
			} else {
				current.WriteRune(character) // spaces included
			}

			continue
		}

		
	// Outside quotes: opening quotes
    if character == '\'' {
        inSingleQuotes = true
        continue
    }

	    if character == '"' {
        inDoubleQuotes = true
        continue
    }

    // 6) Outside quotes: space ends a token
    if character == ' ' {
        if current.Len() > 0 {
            tokens = append(tokens, current.String())
            current.Reset()
        }
        continue
    }

    // 7) Normal character outside quotes
    current.WriteRune(character)
	}

	if current.Len() > 0 {
    tokens = append(tokens, current.String())
}

if inSingleQuotes || inDoubleQuotes {
    return nil, fmt.Errorf("unclosed quote in input")
}

if backlash {
        return nil, fmt.Errorf("trailing backslash at end of input")
    }

return tokens, nil
}	