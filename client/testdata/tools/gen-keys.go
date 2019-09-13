// This helper files generates a bunch of ed25519 keys to be used by the test
// runners. This is done such that the signatures stay stable when the metadata
// is regenerated.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	sign "github.com/flynn/go-tuf/sign"
)

var expirationDate = time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC)

func main() {
	rolenames := []string{
		"root",
		"snapshot",
		"targets",
		"timestamp",
	}

	roles := make(map[string][][]*sign.PrivateKey)

	for _, name := range rolenames {
		keys := [][]*sign.PrivateKey{}

		for i := 0; i < 2; i++ {
			key, err := sign.GenerateEd25519Key()
			assertNotNil(err)
			keys = append(keys, []*sign.PrivateKey{key})
		}

		roles[name] = keys
	}

	s, err := json.MarshalIndent(&roles, "", "    ")
	assertNotNil(err)

	ioutil.WriteFile("keys.json", []byte(s), 0644)
}

func assertNotNil(err error) {
	if err != nil {
		panic(fmt.Sprintf("assertion failed: %s", err))
	}
}
