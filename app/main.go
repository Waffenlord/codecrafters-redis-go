package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

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
		log.Fatal("Failed to bind to port", cfg.Port)
	}
	defer l.Close()

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

		r := bufio.NewReader(conn)
		if err := handleClientHandshake(conn, storage, cfg, r); err != nil {
			conn.Close()
			log.Fatal(err)
		}
		go handleReplication(conn, storage, cfg, r)
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
	r := bufio.NewReader(c)
	handleCommand(c, r, s, config)
}

func handleClientHandshake(
	c net.Conn,
	s *storage.Storage,
	cfg *config.Config,
	r *bufio.Reader,
) error {
	// Ping
	if err := sendHandshakeCommand(c, "PING"); err != nil {
		return fmt.Errorf("error sending PING: %w", err)

	}
	if err := readSimpleString(r, "PONG"); err != nil {
		return fmt.Errorf("error reading PONG: %w", err)

	}
	// replconf
	if err := sendHandshakeCommand(c, "replconf", "listening-port", fmt.Sprintf("%d", cfg.Port)); err != nil {
		return fmt.Errorf("error sending first replconf: %w", err)

	}
	if err := readSimpleString(r, "OK"); err != nil {
		return fmt.Errorf("error reading OK for first replconf: %w", err)

	}
	if err := sendHandshakeCommand(c, "replconf", "capa", "psync2"); err != nil {
		return fmt.Errorf("error sending second replconf: %w", err)
	}
	if err := readSimpleString(r, "OK"); err != nil {
		return fmt.Errorf("error reading OK for second replconf: %w", err)
	}

	// psync
	if err := sendHandshakeCommand(c, "psync", "?", "-1"); err != nil {
		return fmt.Errorf("error sending psync: %w", err)
	}

	if err := readSimpleString(r, fmt.Sprintf("FULLRESYNC %s 0", cfg.MasterReplId)); err != nil {
		log.Printf("error reading FULLRESYNC for psync: %w", err)
	}

	header, err := r.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading RDB length header: %w", err)
	}

	if len(header) < 3 || header[0] != '$' {
		return fmt.Errorf("invalid RDB length header: %q", header)
	}

	size, err := strconv.Atoi(strings.TrimSpace(header[1:]))
	if err != nil {
		return fmt.Errorf("parsing RDB length: %w", err)
	}

	if size < 0 {
		return fmt.Errorf("invalid RDB length: %d", size)
	}

	if _, err := io.CopyN(io.Discard, r, int64(size)); err != nil {
		return fmt.Errorf("reading RDB payload: %w", err)
	}
	return nil
}

func handleReplication(
	c net.Conn,
	s *storage.Storage,
	cfg *config.Config,
	r *bufio.Reader,
) {
	defer c.Close()
	handleCommand(c, r, s, cfg)
}

func handleCommand(
	c net.Conn,
	r *bufio.Reader,
	s *storage.Storage,
	cfg *config.Config,
) {
	client := command.RedisClient{
		InTransaction: false,
		TxQueue:       nil,
		ClientConn:    c,
	}
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
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
			encoded, err := evaluator.EvalProgram(result, s, &client, cfg)
			if err != nil {
				log.Printf("error evaluating: %s", err)
				fmt.Fprintf(c, "Error ocurred while evaluating: %s", err)
				return
			}
			c.Write([]byte(encoded))
		}
	}
}
