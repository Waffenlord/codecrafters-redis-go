package main

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/evaluator"
	"github.com/codecrafters-io/redis-starter-go/app/lexer"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

func main() {	
	/*
		test := []byte("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n")
		lex := lexer.NewLexer(test)
		par := parser.New(lex)
		result, err := par.ParseProgram()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(result)
		encoded, err := evaluator.EvalProgram(result)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(encoded)
	*/

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		fmt.Println("Listening for connections on port 6379")
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}

}

func handleConnection(c net.Conn) {
	defer c.Close()
	buf := make([]byte, 1024)
	for {
		n, err := c.Read(buf)
		if err != nil {
			break
		}
		if n > 0 {
			lex := lexer.NewLexer(buf)
			par := parser.New(lex)
			result, err := par.ParseProgram()
			if err != nil {
				fmt.Println(err)
				fmt.Fprintf(c, "Error ocurred while parsing: %s", err)
				return
			}
			encoded, err := evaluator.EvalProgram(result)
			if err != nil {
				fmt.Println(err)
				fmt.Fprintf(c, "Error ocurred while evaluating: %s", err)
				return
			}
			c.Write([]byte(encoded))
		}
	}

}
