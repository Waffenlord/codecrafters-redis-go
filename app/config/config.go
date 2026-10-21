package config

import (
	"errors"
	"flag"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var port = flag.Int("port", 6379, "port to listen on")
var replicaOf = flag.String("replicaof", "", "host port of the master to replicate from")

type Role string

const (
	MasterRole Role = "master"
	SlaveRole  Role = "slave"
)

type Config struct {
	Port int
	Role Role
	MasterReplId string
	MasterReplOffset int
	ReplicaOfHost string
	ReplicaOfPort int
	Replicas map[net.Conn]struct{}
	Mu sync.Mutex
}

func NewConfig() *Config {
	role := MasterRole
	replicaValue := *replicaOf
	replicaOfHost := ""
	replicaOfPort := 0
	if replicaValue != "" {
		host, port, err := parseReplicaOf(replicaValue)
		if err != nil {
			log.Fatal(err.Error())
		}
		role = SlaveRole
		replicaOfHost = host
		replicaOfPort = port
	}
	return &Config{
		Port: *port,
		Role: role,
		MasterReplId: "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
		MasterReplOffset: 0,
		ReplicaOfHost: replicaOfHost,
		ReplicaOfPort: replicaOfPort,
		Replicas: make(map[net.Conn]struct{}),
	}
}

func parseReplicaOf(replicaValue string) (string, int, error) {
	parts := strings.Split(replicaValue, " ")
	if len(parts) != 2 {
		return "", 0, errors.New("invalid parameters for replica of")
	}
	port, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, err
	}
	return parts[0], int(port), nil
}


func (sc *Config) RegisterReplica(c net.Conn) {
	sc.Mu.Lock()
	defer sc.Mu.Unlock()
	sc.Replicas[c] = struct{}{}
}

func (sc *Config) PropagateCommand(c string) {
	sc.Mu.Lock()
	defer sc.Mu.Unlock()

	for conn := range sc.Replicas {
		if _, err := conn.Write([]byte(c)); err != nil {
			delete(sc.Replicas, conn)
			continue
		}
	}
}