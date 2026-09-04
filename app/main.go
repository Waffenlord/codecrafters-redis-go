package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/evaluator"
	"github.com/codecrafters-io/redis-starter-go/app/lexer"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func main() {
	storage := storage.NewStorage()
	flag.Parse()

	cfg := config.NewConfig()

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	if err != nil {
		fmt.Println("Failed to bind to port", cfg.Port)
		os.Exit(1)
	}

	if cfg.Role == config.SlaveRole {
		conn, err := net.Dial(
			"tcp",
			net.JoinHostPort(cfg.ReplicaOfHost, strconv.Itoa(cfg.ReplicaOfPort)),
		)
		if err != nil {
			fmt.Println("Error connecting to master: ", err.Error())
			os.Exit(1)
		}

		fmt.Println("Connected to master on port", cfg.ReplicaOfPort)
		handleClientHandshake(conn, storage, cfg)
	}

	for {
		fmt.Println("Listening for connections on port", cfg.Port)
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleServerConnection(conn, storage, cfg)
	}

}

func handleServerConnection(c net.Conn, s *storage.Storage, config *config.Config) {
	defer c.Close()
	client := command.RedisClient{
		InTransaction: false,
		TxQueue:       nil,
	}
	buf := make([]byte, 4096)
	for {
		n, err := c.Read(buf)
		if err != nil {
			break
		}
		if n > 0 {
			lex := lexer.NewLexer(buf[:n])
			par := parser.New(lex)
			result, err := par.ParseProgram()
			if err != nil {
				log.Printf("error parsing: %s", err)
				fmt.Fprintf(c, "Error ocurred while parsing: %s", err)
				return
			}
			encoded, err := evaluator.EvalProgram(result, s, &client, config)
			if err != nil {
				log.Printf("error evaluating: %s", err)
				fmt.Fprintf(c, "Error ocurred while evaluating: %s", err)
				return
			}
			c.Write([]byte(encoded))
		}
	}
}

func handleClientHandshake(c net.Conn, s *storage.Storage, cfg *config.Config) {
	defer c.Close()
	r := bufio.NewReader(c)
	// Ping
	if err := sendHandshakeCommand(c, "PING"); err != nil {
		log.Printf("error sending PING: %s", err)
		return
	}
	if err := readSimpleString(r, "PONG"); err != nil {
		log.Printf("error reading PONG: %s", err)
		return
	}

	// replconf
	if err := sendHandshakeCommand(c, "replconf", "listening-port", fmt.Sprintf("%d", cfg.Port)); err != nil {
		log.Printf("error sending first replconf: %s", err)
		return
	}
	if err := readSimpleString(r, "OK"); err != nil {
		log.Printf("error reading OK for first replconf: %s", err)
		return
	}
	if err := sendHandshakeCommand(c, "replconf", "capa", "psync2"); err != nil {
		log.Printf("error sending second replconf: %s", err)
		return
	}
	if err := readSimpleString(r, "OK"); err != nil {
		log.Printf("error reading OK for second replconf: %s", err)
		return
	}

	// psync
	if err := sendHandshakeCommand(c, "psync", "?", "-1"); err != nil {
		log.Printf("error sending psync: %s", err)
		return
	}

	if err := readSimpleString(r, fmt.Sprintf("FULLRESYNC %s 0", cfg.MasterReplId)); err != nil {
		log.Printf("error reading FULLRESYNC for psync: %s", err)
		return
	}
}
