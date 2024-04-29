package catfile

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/git-starter-go/internal/object"
)

func CatFilePrettyPrint(gitObjectName string) {
	r, err := object.GetZlibReaderFromBlob(gitObjectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
	defer r.Close()
	// Read until null byte encountered
	parts, err := object.SplitBufferByNullByteN(r, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
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
