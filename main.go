package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Krisp Interpreter")
		fmt.Println("Usage: krisp <path_to_file>")
		fmt.Println("Example: ./krisp test.ave")
		return
	}

	filename := os.Args[1]

	lowerName := strings.ToLower(filename)
	validExt := strings.HasSuffix(lowerName, ".ave") || strings.HasSuffix(lowerName, ".lyn")  || strings.HasSuffix(lowerName, ".Aveline") || strings.HasSuffix(lowerName, ".Avelyn") || strings.HasSuffix(lowerName, ".avn") || strings.HasSuffix(lowerName, ".avx") || strings.HasSuffix(lowerName, ".lne")
	if !validExt {
		fmt.Println("Error: Supported extensions are .ave, .Aveline, .Avelyn, .avn, .avx, .lne, .lyn")
		return
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error: Could not open file at '%s'\n", filename)
		return
	}

	// 4. Run the Interpreter Pipeline
	sourceCode := string(content)
	tokens := Tokenize(sourceCode)
	parser := NewParser(tokens)
	program := parser.ProduceAST()

	env := CreateGlobalEnv()
	_, evalErr := Evaluate(program, env)

	if evalErr != nil {
		// Skip printing return errors (they're normal control flow)
		if _, ok := evalErr.(*ReturnError); !ok {
			fmt.Printf("\n[Runtime Error]: %v\n", evalErr)
		}
	}
}
