package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/theupdateframework/go-tuf/client/testdata/go-tuf/generator"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	generator.Generate(filepath.Join(cwd, "consistent-snapshot-false"), "../keys.json", false)
	generator.Generate(filepath.Join(cwd, "consistent-snapshot-true"), "../keys.json", true)
}
