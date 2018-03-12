package main

import (
	"bufio"
	"flag"
	"log"

	"github.com/tarm/serial"
)

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

	data := []byte{'C', 'y', 'm', 'a',
		(byte)((c1) >> 8),
		(byte)(c1) & 0xFF,
		0,
		(byte)((c2) >> 8),
		(byte)(c2) & 0xFF,
		0,
		(byte)((c3) >> 8),
		(byte)(c3) & 0xFF,
		0,
		(byte)((c4) >> 8),
		(byte)(c4) & 0xFF,
		0,
		(byte)((c5) >> 8),
		(byte)(c5) & 0xFF,
		0,
		(byte)((c6) >> 8),
		(byte)(c6) & 0xFF,
		0,
		(byte)((c7) >> 8),
		(byte)(c7) & 0xFF,
		0,
		(byte)((c8) >> 8),
		(byte)(c8) & 0xFF,
		0,
	}
	data[6] = (byte)(data[4] ^ data[5] ^ 0x55)
	data[9] = (byte)(data[7] ^ data[8] ^ 0x55)
	data[12] = (byte)(data[10] ^ data[11] ^ 0x55)
	data[15] = (byte)(data[13] ^ data[14] ^ 0x55)
	data[18] = (byte)(data[16] ^ data[17] ^ 0x55)
	s.Write(data)

	reader := bufio.NewReader(s)
	for {
		data, err := reader.ReadBytes('\x0a')
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s\n", data)
	}
	/*
		b := make([]byte, 256)
		for {
			_, err := s.Read(b)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%s\n", b)
		}
	*/
}
