package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
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

func CreateTree(dirName string) (string, error) {
	excludeDir := ".git"
	entries, _ := os.ReadDir(dirName)
	var err error
	treeEntries := []TreeEntry{}
	for _, entry := range entries {
		if entry.Name() != excludeDir {
			te, err := createTreeEntryFromDirEntry(dirName, entry)
			if err != nil {
				return "", fmt.Errorf("%s\n", err)
			}
			treeEntries = append(treeEntries, te)

		}
	}
	n, buff, err := TreeEntriesToByteBuffer(treeEntries)
	if err != nil {
		return "", fmt.Errorf("%s\n", err)
	}
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
	io.WriteString(multiWriter, fmt.Sprintf("tree %d\000", n))
	multiWriter.Write(buff.Bytes())

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

func createTreeEntryFromDirEntry(dirName string, entry fs.DirEntry) (TreeEntry, error) {
	var hexDigest string
	var fileMode string
	var err error
	var te TreeEntry

	if entry.Type().IsDir() {
		fileMode = "40000"
		hexDigest, err = CreateTree(path.Join(dirName, entry.Name()))
		if err != nil {
			return te, fmt.Errorf("%s\n", err)
		}

	} else {
		fileMode = "100644"
		hexDigest, err = CreateBlob(path.Join(dirName, entry.Name()), true) // TODO: Consider returning byte array so that consumer can decide to convert or return both.
		if err != nil {
			return te, fmt.Errorf("%s\n", err)
		}
	}
	te = TreeEntry{
		Mode: fileMode,
		Name: entry.Name(),
	}
	err = te.SetByteShaFromHexDigest(hexDigest)
	if err != nil {
		return te, fmt.Errorf("%s\n", err)
	}
	return te, nil
}

func CreateBlob(filename string, write bool) (string, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return "", fmt.Errorf("fatal: Could not find file: %s\n", err)
	}
	hasher := sha1.New()
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("fatal: Could not compute file size: %s\n", err)
	}

	// We want to Write to both hasher and zlib in a stream fashion for optimal memory usage. But we won't know the
	// filename until we compute the hash after writing to hasher. So we write to tempfile first and then rename finally
	tempWriter, err := os.CreateTemp("", "mygithashobject")
	defer tempWriter.Close()
	zlibWriter := zlib.NewWriter(tempWriter)
	defer zlibWriter.Close()

	if err != nil {
		return "", fmt.Errorf("fatal: could not create temp file: %s", err)
	}
	multiWriter := io.MultiWriter(hasher, zlibWriter)
	io.WriteString(multiWriter, fmt.Sprintf("blob %d\000", fileInfo.Size()))
	io.Copy(multiWriter, file)
	checksum := hasher.Sum(nil)
	hash, err := hex.EncodeToString(checksum), nil
	if err != nil {
		return "", fmt.Errorf("error: Could not compute hash: %s", err)
	}
	if write {
		dir := fmt.Sprintf(".git/objects/%s", hash[:2])
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("Error creating directory: %s\n", err)
		}
		if err := os.Rename(tempWriter.Name(), fmt.Sprintf("%s/%s", dir, hash[2:])); err != nil {
			return "", fmt.Errorf(
				"Error renaming tempfile to git object file in git directory: %s",
				err,
			)
		}

	}
	return hash, nil
}

type Object struct {
	Kind   string
	Size   int
	Reader io.Reader
}

type TreeEntry struct {
	Mode string
	Name string
	Sha  []byte
}

func (t *TreeEntry) SetByteShaFromHexDigest(digest string) error {
	sha, err := hex.DecodeString(digest)
	if err != nil {
		return err
	}
	t.Sha = sha
	return nil
}

func (t *TreeEntry) FileModeToString() string {
	//  TODO: Change later
	return t.Mode
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

		n, err = buff.WriteString(entry.Name)
		if err != nil {
			return -1, buff, err
		}
		totalSize += n

		n, err = buff.WriteString("\000")
		if err != nil {
			return -1, buff, err
		}
		totalSize += n

		n, err = buff.Write(entry.Sha)
		if err != nil {
			return -1, buff, err
		}
		totalSize += n

	}
	return totalSize, buff, nil
}

func NewObject(hash string) (Object, error) {
	obj := Object{}
	r, err := GetZlibReaderFromBlob(hash)
	if err != nil {
		return obj, fmt.Errorf("%s", err)
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
		return obj, fmt.Errorf("fatal: Invalid size in header of object file: %s\n", err)
	}
	obj.Size = int(contentSize) // TODO: Lossy conversion. Fix it
	obj.Reader = r
	return obj, nil
}
