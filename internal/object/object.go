package object

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strings"
)

func GetZlibReaderFromBlob(gitObjectName string) (io.ReadCloser, error) {
	matches := GetObjectsMatchingPrefix(".git/objects", gitObjectName)
	if len(matches) > 1 {
		return nil, fmt.Errorf("error: short object ID %s is ambiguous", gitObjectName)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("fatal: Not a valid object name: %s\n", gitObjectName)
	}
	gitObject, err := os.Open(
		fmt.Sprintf(".git/objects/%s", matches[0]),
	)
	if err != nil {
		return nil, fmt.Errorf("fatal: Not a valid object name: %s\n", err)
	}
	r, err := zlib.NewReader(gitObject)
	if err != nil {
		return nil, fmt.Errorf("fatal: Unable to decompress the object file: %s\n", err)
	}
	return r, nil
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

func SplitBufferByNullByte(r io.Reader) [][]byte {
	return SplitBufferByNullByteN(r, -1)
}

func SplitBufferByNullByteN(r io.Reader, count int) [][]byte {
	buf := make([]byte, 1)
	var part []byte
	var parts [][]byte
	counter := 0

	for {
		n, err := r.Read(buf)
		if n > 0 && buf[0] == 0 {
			counter++
			parts = append(parts, part)
			part = []byte{}
			if counter == count {
				break
			}
		}
		if err != nil {
			if err == io.EOF {
				part = []byte{}
				break
			}
			fmt.Fprintf(os.Stderr, "fatal: Could not read object file: %s\n", err)
		}

		part = append(part, buf[0])
	}
	return parts
}
