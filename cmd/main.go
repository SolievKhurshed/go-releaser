package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const usageMessage = "" +
	`Usage of 'go-releaser': go-releaser [option] [flags]

Application used for get absolute from relative path.

Options (only one required)
    - info: info about application

Flags (all required):
	-f <path>: relative or absolute path

Examples:
	./go-releaser -f /var/lib/rpm/rpmdb.sqlite
	./go-releaser info`

const (
	SUCCESS      = 0  //OK
	EXIT_FAILURE = 1  //Failure.
	EINVAL       = 2  //Invalid argument.
	EIO          = 5  //Input/output error.
	EACCES       = 13 //Permission denied.
)

type StatusErr struct {
	Status  int
	Message string
}

func (se StatusErr) Error() string {
	return se.Message
}

// Application Info
var (
	version     = ""
	commitHash  = ""
	description = "Used for get absolute from relative path"
)

var (
	allowedFirstArgs = map[string]struct{}{
		"r":    {},
		"info": {},
	}
)

var (
	fileFlag = flag.String("f", "", "path to file")
)

func main() {
	os.Exit(mainImpl())
}

func mainImpl() int {
	if len(os.Args) > 1 {
		arg := os.Args[1]

		var genErr *StatusErr = ValidateFirstArg(arg)
		if genErr != nil {
			fmt.Fprintf(os.Stderr, "%s\n", genErr.Message)
			return genErr.Status
		}

		if arg == "info" {
			fmt.Fprintf(os.Stdout, "Info: %s\nVersion: %s\nCommit: %s\n", description, version, commitHash)
			return SUCCESS
		}

		os.Args = os.Args[1:]

		flag.Usage = func() {
			fmt.Fprintf(os.Stderr, "%s\n", usageMessage)
			flag.PrintDefaults()
			os.Exit(EINVAL)
		}
		flag.Parse()

		absPath, pathErr := resolvePath(*fileFlag)
		if pathErr != nil {
			var stErr StatusErr
			if errors.As(pathErr, &stErr) {
				fmt.Fprintf(os.Stderr, "%s\n", stErr.Message)
				return stErr.Status
			}
			return EXIT_FAILURE
		}

		fmt.Printf("Relative path: %s\n", *fileFlag)
		fmt.Printf("Absolute path: %s\n", absPath)

		return SUCCESS

	} else {
		fmt.Fprintf(os.Stderr, "%s\n", "no args are presented")
		fmt.Println(usageMessage)
		return SUCCESS
	}
}

func ValidateFirstArg(arg string) *StatusErr {
	if _, ok := allowedFirstArgs[arg]; !ok {
		s := ConcatStr(allowedFirstArgs)
		return &StatusErr{
			Status:  EINVAL,
			Message: fmt.Sprintf("invalid positional arg1. Allowed options are: %s", s),
		}
	}
	return nil
}

func ConcatStr(allowed map[string]struct{}) string {
	sb := strings.Builder{}
	for k := range allowed {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(k)
	}
	return sb.String()
}

func resolvePath(path string) (string, error) {
	if path == "" {
		return "", StatusErr{
			Status:  EINVAL,
			Message: "path cannot be empty",
		}
	}

	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", StatusErr{
				Status:  EIO,
				Message: fmt.Sprintf("failed to get home directory: %v", err),
			}
		}

		switch {
		case path == "~":
			path = homeDir
		case path == "~/":
			path = homeDir + string(filepath.Separator)
		case strings.HasPrefix(path, "~/"):
			path = filepath.Join(homeDir, path[2:])
		default:
			path = filepath.Join(homeDir, path[1:])
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", StatusErr{
			Status:  EINVAL,
			Message: fmt.Sprintf("failed to resolve absolute path: %v", err),
		}
	}

	return filepath.Clean(absPath), nil
}
