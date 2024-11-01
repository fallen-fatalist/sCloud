package web

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Errors
var (
	ErrHelpCalled               = errors.New("help called in command line argument")
	ErrInvalidNumberOfArguments = errors.New("command line arguments number incorrect")
)

// Flags list
var (
	Port        = 4000
	storagePath = "data"
)

func Parse(args []string) (err error) {
	for _, arg := range args {
		if arg == "--help" {
			PrintHelp()
			return ErrHelpCalled
		}
	}

	if len(args)%2 != 0 {
		return ErrInvalidNumberOfArguments
	}

	for flagIdx := 0; flagIdx < len(args); flagIdx += 2 {
		flagName, flagValue := args[flagIdx], args[flagIdx+1]
		switch strings.TrimPrefix(flagName, "--") {
		case "port":
			Port, err = strconv.Atoi(flagValue)
			if err != nil {
				return fmt.Errorf("error while parsing the port: %w", err)
			}
		case "dir":
			storagePath = flagValue
		}
	}

	return nil
}

func PrintHelp() {
	fmt.Println("Simple Storage Service.")
	fmt.Println("")
	fmt.Println("**Usage:**")
	fmt.Println("\ttriple-s [--port <N>] [--dir <S>]")
	fmt.Println("\ttriple-s --help")
	fmt.Println("")
	fmt.Println("**Options:**")
	fmt.Println("- --help     Show this screen.")
	fmt.Println("- --port N   Port number")
	fmt.Println("- --dir S    Path to the directory")
}
