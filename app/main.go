package main

import (
	"fmt"
	"os"

	"github.com/Sinacam/autokey"
)

func main() {
	ymlfile := "input.yml"
	if len(os.Args) > 1 {
		ymlfile = os.Args[1]
	}

	yml := map[interface{}]interface{}{
		"file": ymlfile,
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
