package object

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strconv"
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

func SplitBufferByNullByte(r io.Reader) ([][]byte, error) {
	return SplitBufferByNullByteN(r, -1)
}

func SplitBufferByNullByteN(r io.Reader, count int) ([][]byte, error) {
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
			return parts, fmt.Errorf("fatal: Could not read object file: %s\n", err)
		}

		part = append(part, buf[0])
	}
	return parts, nil
}

type Object struct {
	Kind   string
	Size   int
	Reader io.Reader
}

func NewObject(hash string) (Object, error) {
	obj := Object{}
	r, err := GetZlibReaderFromBlob(hash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
	defer r.Close()
	// Read until null byte encountered
	parts, err := SplitBufferByNullByteN(r, 1)
	if err != nil {
		return obj, fmt.Errorf("%s\n", err)
	}
	header := parts[0]
	headerParts := strings.Split(string(header), " ")
	obj.Kind = headerParts[0]
	contentSize, err := strconv.ParseInt(
		headerParts[1],
		10,
		64,
	) // TODO: Check why it's int64 and why copyN needs int64 below
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: Invalid size in header of object file: %s\n", err)
	}
	obj.Size = int(contentSize) // TODO: Lossy conversion. Fix it
	obj.Reader = r
	return obj, nil
}
