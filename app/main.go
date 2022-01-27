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
	time.Sleep(5 * time.Second)
}

func test() {
	autokey.Init()
	defer autokey.Teardown()

	ch := make(chan autokey.Input)
	go func() {
		for i := range ch {
			fmt.Printf("%c", i.Key)
			autokey.Send(autokey.Input{Key: 'B', Flag: autokey.KeyDown})
			autokey.Send(autokey.Input{Key: 'B', Flag: autokey.KeyUp})
		}
	}()

	autokey.NotifyOn(ch, autokey.Keys("qwerty")...)
	fmt.Println("Installed")

	time.Sleep(5 * time.Second)
	fmt.Println("\nUninstalled")
}
