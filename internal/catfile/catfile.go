package catfile

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func CatFilePrettyPrint(gitObjectName string) {
	matches := GetObjectsMatchingPrefix(".git/objects", gitObjectName)
	if len(matches) > 1 {
		fmt.Fprintf(os.Stderr, "error: short object ID %s is ambiguous", gitObjectName)
		return
	}
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "fatal: Not a valid object name: %s\n", gitObjectName)
		return
	}
	gitObject, err := os.Open(
		fmt.Sprintf(".git/objects/%s", matches[0]),
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
	contentSize, err := strconv.ParseInt(
		headerParts[1],
		10,
		64,
	) // TODO: Check why it's int64 and why copyN needs int64 below
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: Invalid size in header of object file: %s\n", err)
	}
	io.CopyN(os.Stdout, r, contentSize)
}

func GetObjectsMatchingPrefix(directoryName string, prefix string) []string {
	entries, _ := os.ReadDir(fmt.Sprintf("%s/%s", directoryName, prefix[:2]))
	matches := []string{}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix[2:]) {
			matches = append(matches, fmt.Sprintf("%s/%s", prefix[:2], e.Name()))
		}
	}
	return matches
}
