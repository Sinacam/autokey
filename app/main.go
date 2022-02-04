package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Sinacam/autokey"
	"gopkg.in/yaml.v2"
)

func main() {
	f, err := os.Open("short.yml")
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

	fn, err := Compile(yml)
	if err != nil {
		fmt.Println(err)
		return
	}

	autokey.Init()
	defer autokey.Teardown()
	fmt.Println("Installed")
	fn()
	time.Sleep(10 * time.Second)
}
