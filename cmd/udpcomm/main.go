package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/tgreiser/cymapper"
)

/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func main() {
	Conn, err := net.Dial("udp", "192.168.0.113:1331")
	CheckError(err)
	if err != nil {
		os.Exit(0)
	}
	defer Conn.Close()

	c1 := 150
	c2 := 150
	c3 := 720
	c4 := 150
	c5 := 150
	c6 := 0
	c7 := 0
	c8 := 0

	data := cymapper.Handshake(c1, c2, c3, c4, c5, c6, c7, c8)
	Conn.Write(data)
	// wait for response
	status, err := bufio.NewReader(Conn).ReadString('\n')
	CheckError(err)
	if status == "Space" {

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
			time.Sleep(time.Millisecond * 100)
		}
	}
}
