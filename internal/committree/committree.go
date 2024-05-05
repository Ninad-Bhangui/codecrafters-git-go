package committree

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

func InvokeCommitTree(treeHash, commitMsg string, parentTreeHashes []string) {
	hash, err := CommitTree(treeHash, commitMsg, parentTreeHashes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: Could not commit tree: %s", err)
	}
	fmt.Println(hash)
}

func CommitTree(treeHash, commitMsg string, parentTreeHashes []string) (string, error) {
	tempTreeFile, err := os.CreateTemp("", "mygittreeobject")
	if err != nil {
		return "", fmt.Errorf("%s\n", err)
	}

	defer tempTreeFile.Close()

	zlibWriter := zlib.NewWriter(tempTreeFile)
	defer zlibWriter.Close()
	hasher := sha1.New()
	multiWriter := io.MultiWriter(hasher, zlibWriter)

	// Actual Write
	n, content, err := commitContent(treeHash, commitMsg, parentTreeHashes)
	if err != nil {
		return "", err
	}
	io.WriteString(multiWriter, fmt.Sprintf("commit %d\000", n))
	multiWriter.Write(content.Bytes())

	checksum := hasher.Sum(nil)
	hash, err := hex.EncodeToString(checksum), nil
	if err != nil {
		return "", fmt.Errorf("error: Could not compute hash: %s", err)
	}
	dir := fmt.Sprintf(".git/objects/%s", hash[:2])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("Error creating directory: %s\n", err)
	}
	if err := os.Rename(tempTreeFile.Name(), fmt.Sprintf("%s/%s", dir, hash[2:])); err != nil {
		return "", fmt.Errorf(
			"Error renaming tempfile to git object file in git directory: %s",
			err,
		)
	}
	return hash, nil
}

func commitContent(
	treeHash, commitMsg string,
	parentTreeHashes []string,
) (int, bytes.Buffer, error) {
	// TODO: Replace somehow maybe with env
	authorName := os.Getenv("GIT_AUTHOR_NAME")
	authorEmail := os.Getenv("GIT_AUTHOR_EMAIL")
	now := time.Now()
	_, offset := now.Zone()

	offsetHours := offset / 3600
	offsetSeconds := (offset % 3600) / 60

	var offsetSymbol string
	offsetSymbol = "+"
	if offset < 0 {
		offsetSymbol = "-"
	}
	authorDateSeconds := now.Unix()
	authorDateTimezone := fmt.Sprintf("%s%02d%02d", offsetSymbol, offsetHours, offsetSeconds)
	buff := bytes.Buffer{}
	size := 0
	n, _ := buff.WriteString(fmt.Sprintf("tree %s", treeHash))
	size += n
	for _, hash := range parentTreeHashes {
		n, _ := buff.WriteString(fmt.Sprintf("%s\n", hash))
		size += n
	}
	n, _ = buff.WriteString(
		fmt.Sprintf(
			"author %s <%s> %d %s\n",
			authorName,
			authorEmail,
			authorDateSeconds,
			authorDateTimezone,
		),
	)

	size += n
	n, _ = buff.WriteString(
		fmt.Sprintf(
			"committer %s <%s> %d %s\n",
			authorName,
			authorEmail,
			authorDateSeconds,
			authorDateTimezone,
		),
	)

	size += n
	n, _ = buff.WriteString("\n")
	size += n
	n, _ = buff.WriteString(commitMsg)
	size += n

	n, _ = buff.WriteString("\n")
  size += n
	return size, buff, nil
}
