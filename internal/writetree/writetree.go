package writetree

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/codecrafters-io/git-starter-go/internal/object"
)

var octalMode = "%06o"

type TreeEntry struct {
	mode fs.FileMode
	name string
	sha  []byte
}

func (t *TreeEntry) SetByteShaFromHexDigest(digest string) error {
	sha, err := hex.DecodeString(digest)
	if err != nil {
		return err
	}
	t.sha = sha
	return nil
}

func (t *TreeEntry) FileModeToString() string {
	//  TODO: Change later
	return "100644"
	// return fmt.Sprintf(octalMode, t.mode)
}

func TreeEntriesToByteBuffer(entries []TreeEntry) (int, bytes.Buffer, error) {
	buff := bytes.Buffer{}

	totalSize := 0
	for _, entry := range entries {
		n, err := buff.WriteString(entry.FileModeToString())
		if err != nil {
			return totalSize, buff, err
		}
		totalSize += n

		n, err = buff.WriteString(" ")
		if err != nil {
			return -1, buff, err
		}
		totalSize += n

		n, err = buff.WriteString(entry.name)
		if err != nil {
			return -1, buff, err
		}
		totalSize += n

		n, err = buff.WriteString("\000")
		if err != nil {
			return -1, buff, err
		}
		totalSize += n

		n, err = buff.Write(entry.sha)
		if err != nil {
			return -1, buff, err
		}
		totalSize += n

	}
	return totalSize, buff, nil
}

func WriteTree() {
	excludeDir := ".git"
	entries, _ := os.ReadDir(".")

	treeEntries := []TreeEntry{}
	for _, entry := range entries {
		if entry.Name() != excludeDir {
			if entry.Type().IsDir() {
				fmt.Printf("%s is a directory\n", entry.Name())
			} else {
				fileInfo, _ := entry.Info()
				fileMode := fileInfo.Mode().Perm()
				fmt.Printf("%s is a file with mode %s.\n", entry.Name(), fmt.Sprintf(octalMode, fileMode))

				hexDigest, err := object.CreateBlob(entry.Name(), true) // TODO: Consider returning byte array so that consumer can decide to convert or return both.
				if err != nil {
					fmt.Fprint(os.Stderr, "%s\n", err)
				}
				te := TreeEntry{
					mode: fileMode,
					name: entry.Name(),
				}
				err = te.SetByteShaFromHexDigest(hexDigest)
				if err != nil {
					fmt.Fprint(os.Stderr, "%s\n", err)
				}
				// TODO: Change it to stream mode to avoid in memory collection
				treeEntries = append(treeEntries, te)
			}
		}
	}

	n, buff, err := TreeEntriesToByteBuffer(treeEntries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	tempTreeFile, err := os.CreateTemp("", "mygittreeobject")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	defer tempTreeFile.Close()
	zlibWriter := zlib.NewWriter(tempTreeFile)
	defer zlibWriter.Close()
	hasher := sha1.New()
	multiWriter := io.MultiWriter(hasher, zlibWriter)
	// // TESTING IF ZLIB IS WORKING FINE BY WRITING SIMPLE string
	// io.WriteString(multiWriter, "teststring")

	// Actual Write
	io.WriteString(multiWriter, fmt.Sprintf("tree %d\000", n))
	multiWriter.Write(buff.Bytes())

	checksum := hasher.Sum(nil)
	hash, err := hex.EncodeToString(checksum), nil
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: Could not compute hash: %s", err)
	}
	dir := fmt.Sprintf(".git/objects/%s", hash[:2])
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
	}
	if err := os.Rename(tempTreeFile.Name(), fmt.Sprintf("%s/%s", dir, hash[2:])); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Error renaming tempfile to git object file in git directory: %s",
			err,
		)
	}
	fmt.Println(hash)

	//
}
