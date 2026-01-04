package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// Get a *net.UDPAddr to use
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("Couldn't resolve the address")
	}
	// Establishes UDP connection (no handshake) + will only allow incoming from raddr and will only send to raddr (client-style)
	// To accept info incoming from anywhere (server-style) use ListenUDP
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal("Couldn't dial UDP address")
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">	")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("reader error")
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Fatal("connection couldn't write user input")
		}
	}
}
