package main

import (
	"os"

	"github.com/progrium/go-basher"
)

func main() {
	bash := basher.NewContext()

	bash.Source("./test.bash")
	bash.Run("main", os.Args[1:])
}
