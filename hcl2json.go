package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hashicorp/hcl"
)

func main() {
	var err error

	// Use io.ReadAll instead of ioutil.ReadAll
	buffer, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var converted interface{}
	if err := hcl.Unmarshal(buffer, &converted); err != nil {
		log.Fatal(err)
	}

	// MarshalIndent for pretty JSON output
	output, err := json.MarshalIndent(&converted, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
}
