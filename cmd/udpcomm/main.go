package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", "192.168.0.113:9999")
	CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "192.168.0.100:0")
	CheckError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)
	if err != nil {
		os.Exit(0)
	}

	defer Conn.Close()
	i := 1000000000
	for {
		msg := "Other kernel code can be swapped out, in principle. However this can be problematic for several reasons. Hardware drivers that aren't involved in swap could be swapped " + strconv.Itoa(i)
		i++
		buf := []byte(msg)
		fmt.Println(msg)
		_, err := Conn.Write(buf)
		if err != nil {
			fmt.Println(msg, err)
		}
		time.Sleep(time.Millisecond * 10)
	}
}
