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

func rpush(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 2 {
		return errors.New("invalid number of arguments for rpush command")
	}

	key := args[0]
	values := args[1:]

	v, found := s.Get(key)
	if !found {
		first := values[0]
		newList := storage.NewList(first)
		for i := 1; i < len(values); i++ {
			newList.AppendR(values[i])
		}
		s.Set(key, newList)
		fmt.Fprint(out, FormatInteger(newList.Len))
		return nil
	}
	switch data := v.(type) {
	case *storage.ListType:
		newLength := 0
		for i := range values {
			newLength = data.AppendR(values[i])
		}
		fmt.Fprint(out, FormatInteger(newLength))
		return nil
	default:
		return errors.New("value stored should be a list")
	}
}

func isRangeValid(length int, startIdx int, endIdx int) (int, int, bool) {
	if startIdx < 0 {
		startIdx = max(length+startIdx, 0)
	}
	if endIdx < 0 {
		endIdx = max(length+endIdx, 0)
	}
	if startIdx >= length {
		return 0, 0, false
	}
	if startIdx > endIdx {
		return 0, 0, false
	}
	if endIdx >= length {
		return startIdx, length - 1, true
	}

	return startIdx, endIdx, true
}

func lrange(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 3 {
		return errors.New("invalid number of arguments for lrange command")
	}

	key := args[0]
	startIdx := args[1]
	endIdx := args[2]

	startIdxInt, err := strconv.Atoi(startIdx)
	if err != nil {
		return errors.New("invalid value for start index")
	}
	endIdxInt, err := strconv.Atoi(endIdx)
	if err != nil {
		return errors.New("invalid value for end index")
	}

	v, found := s.Get(key)
	if !found {
		fmt.Fprint(out, FormatArray(make([]string, 0)))
		return nil
	}
	switch data := v.(type) {
	case *storage.ListType:
		cleanedStartIdx, cleanedEndIdx, isValid := isRangeValid(data.Len, startIdxInt, endIdxInt)
		if !isValid {
			fmt.Fprint(out, FormatArray(make([]string, 0)))
			return nil
		}
		result := data.LRange(cleanedStartIdx, cleanedEndIdx)
		fmt.Fprint(out, FormatArray(result))
		return nil
	default:
		return errors.New("value stored should be a list")
	}
}

func lpush(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 2 {
		return errors.New("invalid number of arguments for lpush command")
	}

	key := args[0]
	values := args[1:]

	v, found := s.Get(key)
	if !found {
		first := values[0]
		newList := storage.NewList(first)
		for i := 1; i < len(values); i++ {
			newList.AppendL(values[i])
		}
		s.Set(key, newList)
		fmt.Fprint(out, FormatInteger(newList.Len))
		return nil
	}
	switch data := v.(type) {
	case *storage.ListType:
		newLength := 0
		for i := range values {
			newLength = data.AppendL(values[i])
		}
		fmt.Fprint(out, FormatInteger(newLength))
		return nil
	default:
		return errors.New("value stored should be a list")
	}
}

func llen(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 1 {
		return errors.New("invalid number of arguments for llen command")
	}

	key := args[0]

	v, found := s.Get(key)
	if !found {
		fmt.Fprint(out, FormatInteger(0))
		return nil
	}

	switch data := v.(type) {
	case *storage.ListType:
		data.Mux.RLock()
		defer data.Mux.RUnlock()
		fmt.Fprint(out, FormatInteger(data.Len))
		return nil
	default:
		return errors.New("value stored should be a list")
	}
}

var CommandMenu = map[string]Builtin{
	"echo":   echo,
	"ping":   ping,
	"set":    set,
	"get":    get,
	"rpush":  rpush,
	"lrange": lrange,
	"lpush":  lpush,
	"llen":   llen,
}
