package main

import (
	"bufio"
	"flag"
	"log"

	"github.com/tarm/serial"
	"github.com/tgreiser/cymapper"
)

/*
Example of a more optimized serial protocol.
*/

var comPort = flag.String("com", "COM8", "COM port for teensy")

func main() {
	flag.Parse()

	c := &serial.Config{Name: *comPort, Baud: 500000}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("When connecting to port: %v: %v", *comPort, err)
	}
	defer s.Close()

	c1 := 150
	c2 := 150
	c3 := 720
	c4 := 150
	c5 := 150
	c6 := 0
	c7 := 0
	c8 := 0

	data := cymapper.Handshake(c1, c2, c3, c4, c5, c6, c7, c8)
	s.Write(data)

	reader := bufio.NewReader(s)
	for {
		data, err := reader.ReadBytes('\x0a')
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s\n", data)
	}
}
