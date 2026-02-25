package command

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/storage"
)


type Builtin func(in io.Reader, out io.Writer, args []string, s *storage.Storage) error

func ping(_ io.Reader, out io.Writer, _ []string, _ *storage.Storage) error {
	fmt.Fprint(out, FormatSimpleString("PONG"))
	return nil
}

func echo(_ io.Reader, out io.Writer, args []string, _ *storage.Storage) error {
	result := strings.Join(args, "")
	fmt.Fprint(out, FormatBulkString(result))
	return nil
}

func set(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 2 {
		return errors.New("invalid number of arguments for set command")
	}
	key := args[0]
	value := args[1]

	data := storage.StringType{
		Value: value,
	}
	s.Set(key, &data)
	fmt.Fprint(out, FormatSimpleString("OK"))
	return nil
}

func get(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 1 {
		return errors.New("key must be provided")
	}
	key := args[0]

	v, found := s.Get(key)
	if !found {
		fmt.Fprint(out, FormatNullBulkString())
		return nil
	}

	switch data := v.(type) {
	case *storage.StringType:
		fmt.Fprint(out, FormatBulkString(data.Value))
		return nil
	}
	return nil
}

var CommandMenu = map[string]Builtin{
	"echo": echo,
	"ping": ping,
	"set": set,
	"get": get,
}