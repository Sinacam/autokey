package main

import (
	"fmt"
	"os"

	"github.com/Sinacam/autokey"
	"gopkg.in/yaml.v2"
)

func main() {
	f, err := os.Open("input.yml")
	if err != nil {
		fmt.Println(err)
		return
	}

	var yml interface{}
	err = yaml.NewDecoder(f).Decode(&yml)
	if err != nil {
		fmt.Println(err)
		return
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
