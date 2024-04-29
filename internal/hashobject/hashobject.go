package hashobject

import (
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func HashObject(filename string, write bool) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: Could not find file: %s\n", err)
	}
	hasher := sha1.New()
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: Could not compute file size: %s\n", err)
		return
	}

	// We want to Write to both hasher and zlib in a stream fashion for optimal memory usage. But we won't know the
	// filename until we compute the hash after writing to hasher. So we write to tempfile first and then rename finally
	tempWriter, err := os.CreateTemp("", "mygithashobject")
	defer tempWriter.Close()
	zlibWriter := zlib.NewWriter(tempWriter)
	defer zlibWriter.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: could not create temp file: %s", err)
		return
	}
	multiWriter := io.MultiWriter(hasher, zlibWriter)
	io.WriteString(multiWriter, fmt.Sprintf("blob %d\000", fileInfo.Size()))
	io.Copy(multiWriter, file)
	checksum := hasher.Sum(nil)
	hash, err := hex.EncodeToString(checksum), nil
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: Could not compute hash: %s", err)
	}
	if write {
		dir := fmt.Sprintf(".git/objects/%s", hash[:2])
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			return
		}
		if err := os.Rename(tempWriter.Name(), fmt.Sprintf("%s/%s", dir, hash[2:])); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Error renaming tempfile to git object file in git directory: %s",
				err,
			)
			return
		}
	}

	fmt.Println(hash)
}
