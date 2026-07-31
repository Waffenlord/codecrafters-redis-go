package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

func sendHandshakeCommand(c net.Conn, args ...string) error {
	_, err := fmt.Fprint(c, command.FormatArray(args))
	if err != nil {
		return err
	}
	return nil
}


func readSimpleString(r *bufio.Reader, expected string) error {
	resp, err := r.ReadString('\n')
	if err != nil {
		return err
	}

	formatted := command.FormatSimpleString(expected)

	if resp != formatted {
		return fmt.Errorf("expected %s, got %s", formatted, resp)
	}

	return nil
}