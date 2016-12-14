package main

import (
	"fmt"
	"net"
)

func pingHandler(conn net.Conn, args []string) {
	code := args[1]

	fmt.Fprintf(conn, "%s %s\r\n", "PONG", code)
	fmt.Println("Pinged and Ponged, code is " + code)
}
