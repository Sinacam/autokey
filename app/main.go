package main

import (
	"fmt"
	"time"

	"github.com/Sinacam/autokey"
)

func main() {
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

func BuildEvent() {

}
