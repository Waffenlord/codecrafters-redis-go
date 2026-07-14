package config

import (
	"flag"
)

var port = flag.Int("port", 6379, "port to listen on")
var replicaOf = flag.String("replicaof", "", "host:port of the master to replicate from")

type Role string

const (
	masterRole Role = "master"
	slaveRole  Role = "slave"
)

type Config struct {
	Port int
	Role Role
	MasterReplId string
	MasterReplOffset int64
}

func NewConfig() *Config {
	role := masterRole
	if *replicaOf != "" {
		role = slaveRole
	}
	return &Config{
		Port: *port,
		Role: role,
		MasterReplId: "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
		MasterReplOffset: 0,
	}
}
