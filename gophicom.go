package gophicom

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"

	bcd "github.com/johnsonjh/gobcd"
	"go.bug.st/serial"
)

type IcomRadio struct {
	port               serial.Port
	mode               *serial.Mode
	transceiverAddress byte
	controllerAddress  byte
	reader             *bufio.Reader
}
type SquelchStatus uint8

const (
	Closed SquelchStatus = 0
	Open
)

const preambleByte byte = 0xfe
const endOfMessage byte = 0xfd
const responseOkay byte = 0xfb
const responseNoGood byte = 0xfa

var stdPreamble []byte = bytes.Repeat([]byte{preambleByte}, 2)
var emptyData = []byte{}
var getModeCmd = []byte{0x04}
var setModeCmd = []byte{0x06}

func NewIcomRadio(portName string, transceiverAddress byte, controllerAddress byte) (*IcomRadio, error) {
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Fatal(err)
	}

	radio := new(IcomRadio)
	radio.port = port
	radio.mode = mode
	radio.transceiverAddress = transceiverAddress
	radio.controllerAddress = controllerAddress
	radio.reader = bufio.NewReader(radio.port)
	return radio, nil
}

func NewIcomRadioWithDefaultAddresses(portName string) (*IcomRadio, error) {
	return NewIcomRadio(portName, 0x92, 0xe0)
}

func (radio *IcomRadio) PowerOn() error {
	preambleLen := 8
	switch radio.mode.BaudRate {
	case 19200:
		preambleLen = 27
	case 9600:
		preambleLen = 14
	}

	preamble := bytes.Repeat([]byte{preambleByte}, preambleLen)
	_, err := radio.sendCommand(preamble, []byte{0x18, 0x01}, emptyData, false)
	if err != nil {
		return err
	}
	return nil
}

func (radio *IcomRadio) PowerOff() error {
	_, err := radio.sendCommand(stdPreamble, []byte{0x18, 0x00}, emptyData, false)
	if err != nil {
		return err
	}
	return nil
}

func (radio *IcomRadio) GetFrequency() (uint64, error) {
	data, err := radio.sendCommand(stdPreamble, []byte{0x03}, emptyData, false)
	if err != nil {
		return 0, err
	}
	data = data[1:]
	data = reverse(data)
	freq := bcd.ToUint64(data)
	return freq, nil
}

func (radio *IcomRadio) SetFrequency(frequency uint64) error {
	data := bcd.FromUint64(frequency)
	data = reverse(data)
	data = data[:5]
	data, err := radio.sendCommand(stdPreamble, []byte{0x05}, data, false)
	if data[0] == responseNoGood && err == nil {
		return errors.New("response no good")
	}
	return err
}

func (radio *IcomRadio) GetAudioLevel() (uint8, error) {
	data, err := radio.sendCommand(stdPreamble, []byte{0x14, 0x01}, emptyData, false)
	if err != nil {
		return 0, err
	}
	data = data[2:]
	lvl := bcd.ToUint16(data)
	return uint8(lvl), nil
}

func (radio *IcomRadio) SetAudioLevel(level uint8) error {
	data := bcd.FromUint16(uint16(level))
	_, err := radio.sendCommand(stdPreamble, []byte{0x14, 0x01}, data, false)
	if err != nil {
		return err
	}
	return nil
}

func (radio *IcomRadio) GetSquelchLevel() (uint8, error) {
	data, err := radio.sendCommand(stdPreamble, []byte{0x14, 0x03}, emptyData, false)
	if err != nil {
		return 0, err
	}
	data = data[2:]
	lvl := bcd.ToUint16(data)
	return uint8(lvl), nil
}

func (radio *IcomRadio) SetSquelchLevel(level uint8) error {
	data := bcd.FromUint16(uint16(level))
	_, err := radio.sendCommand(stdPreamble, []byte{0x14, 0x05}, data, false)
	if err != nil {
		return err
	}
	return nil
}

func (radio *IcomRadio) GetSquelchStatus() (SquelchStatus, error) {
	data, err := radio.sendCommand(stdPreamble, []byte{0x15, 0x01}, emptyData, false)
	if err != nil {
		return 0, err
	}
	return SquelchStatus(data[2]), nil
}

func (radio *IcomRadio) GetSquelch2Status() (SquelchStatus, error) {
	data, err := radio.sendCommand(stdPreamble, []byte{0x15, 0x05}, emptyData, false)
	if err != nil {
		return 0, err
	}
	return SquelchStatus(data[2]), nil
}

func (radio *IcomRadio) Close() error {
	return radio.port.Close()
}

func (radio *IcomRadio) sendCommand(preamble []byte, command []byte, data []byte, expectResponse bool) ([]byte, error) {
	msg := []byte{}
	msg = append(msg, preamble...)
	msg = append(msg, radio.transceiverAddress)
	msg = append(msg, radio.controllerAddress)
	msg = append(msg, command...)
	msg = append(msg, data...)
	msg = append(msg, endOfMessage)
	_, err := radio.port.Write(msg)
	if err != nil {
		return nil, err
	}

	response, err := radio.reader.ReadBytes(endOfMessage)
	if err != nil {
		return nil, err
	}
	// fmt.Printf(">> ")
	// dumpByteSlice(response)

	if !bytes.Equal(response, msg) {
		/*	Again due to the nature of the single-wire bus a
			sending device immediately receives the echo of
			the message sent. This can be used as a mechanism
			to detect collision, when the received data is not
			the same as the data sent. In this case the sending
			device must abort transmitting the message and send
			the 'jammer' code as above. After detecting silence
			on the bus the aborted message shall be repeated.
		*/
		return nil, errors.New("message send error")
	}

	response, err = radio.reader.ReadBytes(endOfMessage)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("<< ")
	// dumpByteSlice(response)

	return response[4 : len(response)-1], nil
}

func reverse[T any](original []T) (reversed []T) {
	reversed = make([]T, len(original))
	copy(reversed, original)

	for i := len(reversed)/2 - 1; i >= 0; i-- {
		tmp := len(reversed) - 1 - i
		reversed[i], reversed[tmp] = reversed[tmp], reversed[i]
	}

	return
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
