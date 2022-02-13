package main

import (
	"fmt"

	"github.com/Sinacam/autokey"
)

func main() {
	yml := map[interface{}]interface{}{
		"file": "input.yml",
	}

	expr, err := autokey.Compile(yml)
	if err != nil {
		fmt.Println(err)
		return
	}

	autokey.Init()
	defer autokey.Teardown()
	fmt.Println("Installed, press enter to exit")
	expr.Eval()
	// time.Sleep(10 * time.Second)
	fmt.Scanln()
}
