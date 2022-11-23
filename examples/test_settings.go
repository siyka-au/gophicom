package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/siyka-au/gophicom"
)

func main() {

	args := os.Args[1:]

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()

	radio, err := gophicom.NewIcomRadioWithDefaultAddresses(args[0])
	if err != nil {
		log.Fatal(err)
	}

	freq, err := radio.GetFrequency()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Current frequency: %d\n", freq)

	freq, err = strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Setting frequency to %d\n", freq)
	err = radio.SetFrequency(freq)
	if err != nil {
		log.Fatal(err)
	}

	lvl, _ := radio.GetAudioLevel()
	radio.GetSquelchLevel()

	if lvl == 255 {
		lvl = 100
	} else {
		lvl = 255
	}
	radio.SetAudioLevel(lvl)

	for {
		select {
		case x, ok := <-ch:
			fmt.Println("ch1", x, ok)
			if !ok {
				ch = nil
			}
		case x, ok := <-ch2:
			fmt.Println("ch2", x, ok)
			if !ok {
				ch2 = nil
			}
		}

		if ch == nil && ch2 == nil {
			break
		}

		squelch, err := radio.GetSquelchStatus()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Squelch status: %d\r", squelch)

		time.Sleep(100 * time.Millisecond)
	}
}
