package main

import (
	"fmt"
	"net"
	"os"
	"flag"
	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/evaluator"
	"github.com/codecrafters-io/redis-starter-go/app/lexer"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

var port = flag.Int("port", 6379, "port to listen on")

func main() {
	storage := storage.NewStorage()
	flag.Parse()

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		fmt.Println("Failed to bind to port", *port)
		os.Exit(1)
	}

	for {
		fmt.Println("Listening for connections on port", *port)
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, storage)
	}

}

func handleConnection(c net.Conn, s *storage.Storage) {
	defer c.Close()
	client := command.RedisClient{
		InTransaction: false,
		TxQueue:       nil,
	}
	for {
		buf := make([]byte, 1024)
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
			encoded, err := evaluator.EvalProgram(result, s, &client)
			if err != nil {
				fmt.Println(err)
				fmt.Fprintf(c, "Error ocurred while evaluating: %s", err)
				return
			}
			c.Write([]byte(encoded))
		}
	}

}
