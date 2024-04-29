package catfile

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/git-starter-go/internal/object"
)

func CatFilePrettyPrint(gitObjectName string) {
	matches := object.GetObjectsMatchingPrefix(".git/objects", gitObjectName)
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
	// Read until null byte encountered
	parts := object.SplitBufferByNullByteN(r, 1)
	header := parts[0]
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
