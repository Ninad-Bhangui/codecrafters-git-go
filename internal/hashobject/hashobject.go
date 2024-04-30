package hashobject

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/git-starter-go/internal/object"
)

func HashObject(filename string, write bool) {
  hash, err := object.CreateBlob(filename, write)
  if err != nil {
    fmt.Fprintf(os.Stderr, "%s\n", err)
  }
	fmt.Println(hash)
}
