package lstree

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/codecrafters-io/git-starter-go/internal/object"
)

func LsTree(gitObjectName string, nameOnly bool) {
	r, err := object.GetZlibReaderFromBlob(gitObjectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
	defer r.Close()
	// Read until first null byte encountered
	_, err = object.SplitBufferByNullByteN(r, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
	// header := parts[0]
	// headerParts := strings.Split(string(header), " ")
	// kind := headerParts[0]
	// entries := object.SplitBufferByNullByte(r)
	hashbuf := make([]byte, 20)
	for {
		parts, err := object.SplitBufferByNullByteN(r, 1)
		if len(parts) == 0 {
			break
		}
		n, err := io.ReadFull(r, hashbuf)
		if n != len(hashbuf) {
			fmt.Fprintf(os.Stderr, "fatal: Could not read %d bytes as per git spec", n)
			return
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal: %s\n", err)
			return
		}

		mode_and_name := parts[0]
		mode_and_name_parts := bytes.Split(mode_and_name, []byte(" "))
		if len(mode_and_name_parts) != 2 {
			fmt.Fprintf(os.Stderr, "fatal: Invalid entry")
			return
		}
		modeStr := string(mode_and_name_parts[0])
		name := string(mode_and_name_parts[1])
		if nameOnly {
			fmt.Println(name)
		} else {
			mode, err := strconv.Atoi(modeStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "fatal: Could not parse mode: %s", modeStr)
			}
			hash := hex.EncodeToString(hashbuf)
			obj, err := object.NewObject(hash)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid hash: %s", hash)
			}

			fmt.Printf("%06d %s %s    %s\n", mode, obj.Kind, hash, name)
		}

	}

	return

	// contentSize, err := strconv.ParseInt(
	// 	headerParts[1],
	// 	10,
	// 	64,
	// ) // TODO: Check why it's int64 and why copyN needs int64 below
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "fatal: Invalid size in header of object file: %s\n", err)
	// }
	// io.CopyN(os.Stdout, r, contentSize)
}
