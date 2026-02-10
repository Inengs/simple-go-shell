package util

import (
	"os"
	"strings"
)

func ExpandVariables(input string) string {
	var result strings.Builder

	for i := 0; i < len(input); i++ {
		// Handle escaped dollar sign \$
		if i < len(input)-1 && input[i] == '\\' && input[i+1] == '$' {
			result.WriteByte('$')
			i+=2
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
					i += len(varName) + 1 // skip varName and }
				} else {
					result.WriteString("${")
				}
				continue
			}

			 // Handle $VAR syntax (alphanumeric and underscore only)
            varName := extractVarName(input, i)
            if varName != "" {
                result.WriteString(os.Getenv(varName))
                i += len(varName)
            } else {
                result.WriteByte('$') // just a lone $
            }
            continue
		}
		        // Regular character
        result.WriteByte(input[i])
        i++
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


func ExpandTilde() {
	if 
}