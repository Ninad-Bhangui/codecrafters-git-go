package writetree

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/git-starter-go/internal/object"
)

var octalMode = "%06o"

func WriteTree() {
	hash, err := object.CreateTree(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	fmt.Println(hash)
}
