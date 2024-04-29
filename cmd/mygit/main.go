package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codecrafters-io/git-starter-go/internal/catfile"
	"github.com/codecrafters-io/git-starter-go/internal/hashobject"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage!
	catFileCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)
	catFilePrettyPrintBool := catFileCmd.Bool("p", false, "Pretty Print")

	hashObjectCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)
	hashObjectWriteMode := hashObjectCmd.Bool(
		"w",
		false,
		"Write the object into the object database",
	)

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		catFileCmd.Parse(os.Args[2:])
		args := catFileCmd.Args()
		if *catFilePrettyPrintBool {
			if len(args) != 1 {
				catFileCmd.PrintDefaults()
			}
			gitObjectName := args[0]
			catfile.CatFilePrettyPrint(gitObjectName)
		}
	case "hash-object":
		hashObjectCmd.Parse(os.Args[2:])
		args := hashObjectCmd.Args()
		if len(args) != 1 {
			hashObjectCmd.PrintDefaults()
		}
		hashobject.HashObject(args[0], *hashObjectWriteMode)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
