package command

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

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

type setOptionalArgument struct {
	acceptsValue bool
	method       func(int, *storage.StringType) error
}

var setOptionalArguments = map[string]setOptionalArgument{
	"ex": {
		acceptsValue: true,
		method: func(v int, s *storage.StringType) error {
			u, err := time.ParseDuration(fmt.Sprintf("%ds", v))
			if err != nil {
				return errors.New("invalid EX value")
			}
			s.ExpMil = u.Milliseconds()
			return nil
		},
	},
	"px": {
		acceptsValue: true,
		method: func(v int, s *storage.StringType) error {
			u, err := time.ParseDuration(fmt.Sprintf("%dms", v))
			if err != nil {
				return errors.New("invalid PX value")
			}
			s.ExpMil = u.Milliseconds()
			return nil
		},
	},
}

func set(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 2 {
		return errors.New("invalid number of arguments for set command")
	}
	key := args[0]
	value := args[1]
	data := storage.StringType{
		Value:     value,
		CreatedAt: time.Now(),
	}

	if len(args) > 2 {
		currentIdx := 2
		for currentIdx < len(args) {
			cmdArg := args[currentIdx]
			cmd, ok := setOptionalArguments[strings.ToLower(cmdArg)]
			if !ok {
				return errors.New("invalid set optional argument")
			}
			if cmd.acceptsValue {
				currentIdx++
				if currentIdx >= len(args) {
					return errors.New("missing value for set optional argument")
				}
				arg := args[currentIdx]
				value, err := strconv.Atoi(arg)
				if err != nil {
					return fmt.Errorf("error converting arg value: %s", err)
				}
				err = cmd.method(value, &data)
			}
			currentIdx++
		}
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
		if data.ExpMil > 0 {
			diff := time.Since(data.CreatedAt)
			if diff.Milliseconds() >= data.ExpMil {
				s.Delete(key)
				fmt.Fprint(out, FormatNullBulkString())
				return nil
			}
		}
		fmt.Fprint(out, FormatBulkString(data.Value))
		return nil
	}
	return nil
}

var CommandMenu = map[string]Builtin{
	"echo": echo,
	"ping": ping,
	"set":  set,
	"get":  get,
}
