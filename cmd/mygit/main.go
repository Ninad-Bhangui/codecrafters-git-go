package main

import (
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage!
	catFileCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)
	catFilePrettyPrintBool := catFileCmd.Bool("p", false, "Pretty Print")

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
			catFilePrettyPrint(gitObjectName)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func catFilePrettyPrint(gitObjectName string) {
	gitObject, err := os.Open(
		fmt.Sprintf(".git/objects/%s/%s", gitObjectName[:2], gitObjectName[2:]),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: Not a valid object name: %s\n", err)
		return
	}

	r, err := zlib.NewReader(gitObject) // TODO: Handle error
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: Unable to decompress the object file: %s\n", err)
	}
	defer r.Close()
	decompressedGitObject, _ := io.ReadAll(r)
	data := string(decompressedGitObject[:])
	parts := strings.Split(data, "\x00")
	content := parts[1]
	fmt.Fprintf(os.Stdout, content)
}
