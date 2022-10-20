package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/siyka-au/gophicom"
)

func main() {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()

	radio, err := gophicom.NewIcomRadioWithDefaultAddresses("/dev/ttyUSB1")
	if err != nil {
		log.Fatal(err)
	}

	freq, err := radio.GetFrequency()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d\n", freq)

	err = radio.SetFrequency(119000000)
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
		// radio.GetSquelchStatus()
		// fmt.Printf("%d\r", v)

		time.Sleep(100 * time.Millisecond) // or runtime.Gosched() or similar per @misterbee
	}

	// Open the first serial port detected at 9600bps N81

	// // Send the string "10,20,30\n\r" to the serial port
	// n, err := port.Write([]byte("10,20,30\n\r"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Sent %v bytes\n", n)

	// // Read and print the response

	// buff := make([]byte, 100)
	// for {
	// 	// Reads up to 100 bytes
	// 	n, err := port.Read(buff)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if n == 0 {
	// 		fmt.Println("\nEOF")
	// 		break
	// 	}

	// 	fmt.Printf("%s", string(buff[:n]))

	// 	// If we receive a newline stop reading
	// 	if strings.Contains(string(buff[:n]), "\n") {
	// 		break
	// 	}
	// }
}

func dumpByteSlice(b []byte) {
	var a [16]byte
	n := (len(b) + 15) &^ 15
	for i := 0; i < n; i++ {
		if i%16 == 0 {
			fmt.Printf("%4d", i)
		}
		if i%8 == 0 {
			fmt.Print(" ")
		}
		if i < len(b) {
			fmt.Printf(" %02X", b[i])
		} else {
			fmt.Print("   ")
		}
		if i >= len(b) {
			a[i%16] = ' '
		} else if b[i] < 32 || b[i] > 126 {
			a[i%16] = '.'
		} else {
			a[i%16] = b[i]
		}
		if i%16 == 15 {
			fmt.Printf("  %s\n", string(a[:]))
		}
	}
}
