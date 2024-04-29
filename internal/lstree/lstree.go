package lstree

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/git-starter-go/internal/object"
)

func LsTree(gitObjectName string, nameOnly bool) {
	r, err := object.GetZlibReaderFromBlob(gitObjectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
	defer r.Close()
	// Read until first null byte encountered
	parts := object.SplitBufferByNullByteN(r, 1)
	header := parts[0]
	headerParts := strings.Split(string(header), " ")
	// kind := headerParts[0]
	if nameOnly {
		entries := object.SplitBufferByNullByte(r)
		for _, entry := range entries {
			// fmt.Println("debug: ", string(entry))
			entryParts := strings.Split(string(entry), " ")
			if len(entryParts) == 2 {
				fileName := entryParts[1]
				fmt.Println(fileName)

			}
		}
		return
	}

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
