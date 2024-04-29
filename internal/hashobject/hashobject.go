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
	io.WriteString(hasher, fmt.Sprintf("blob %d\000", fileInfo.Size()))
	io.Copy(hasher, file)
	checksum := hasher.Sum(nil)
	hash, err := hex.EncodeToString(checksum), nil
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: Could not compute hash: %s", err)
	}
	if write {
		dir := fmt.Sprintf(".git/objects/%s", hash[:2])
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		}
		outputFile, err := os.Create(fmt.Sprintf("%s/%s", dir, hash[2:]))
		defer outputFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal: Could not write: %s\n", err)
		}
		file.Seek(0, io.SeekStart)
		zlibWriter := zlib.NewWriter(outputFile)
		defer zlibWriter.Close()
		io.WriteString(zlibWriter, fmt.Sprintf("blob %d\000", fileInfo.Size()))
		io.Copy(zlibWriter, file)
	}

	fmt.Println(hash)
}
