package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/hcl"
)

func main() {
	var err error

	buffer, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var converted interface{}
	if err := hcl.Unmarshal(buffer, &converted); err != nil {
		log.Fatal(err)
	}

	output, err := json.MarshalIndent(&converted, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
}
