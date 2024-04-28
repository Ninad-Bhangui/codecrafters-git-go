package main

import (
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
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

	r, err := zlib.NewReader(gitObject)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: Unable to decompress the object file: %s\n", err)
	}
	defer r.Close()
	buf := make([]byte, 1)
	var header []byte
	// Read until null byte encountered
	for {
		n, err := r.Read(buf)
		if n > 0 && buf[0] == 0 {
			break
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "fatal: Could not read object file: %s\n", err)
		}

	  header = append(header, buf[0])
	}
  headerParts := strings.Split(string(header), " ")
  // kind := headerParts[0]
  contentSize, err := strconv.ParseInt(headerParts[1],10, 64) //TODO: Check why it's int64 and why copyN needs int64 below
  if err != nil {
    fmt.Fprintf(os.Stderr, "fatal: Invalid size in header of object file: %s\n", err)
  }
	io.CopyN(os.Stdout, r, contentSize)
	// decompressedGitObject, _ := io.ReadAll(r)
	// data := string(decompressedGitObject[:])
	// parts := strings.Split(data, "\x00")
	// content := parts[1]
	// fmt.Fprintf(os.Stdout, content)
}
