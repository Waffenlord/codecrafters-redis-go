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
		newList := storage.NewList()
		listLength := 0
		for i := range values {
			if s.Blocker.NotifyClient(key, values[i]) {
				listLength++
				continue
			}
			listLength = newList.AppendR(values[i])
		}
		s.Set(key, newList)
		fmt.Fprint(out, FormatInteger(listLength))
		return nil
	}
	switch data := v.(type) {
	case *storage.ListType:
		newLength := 0
		for i := range values {
			if s.Blocker.NotifyClient(key, values[i]) {
				newLength++
				continue
			}
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
		newList := storage.NewList()
		listLength := 0
		for i := range values {
			if s.Blocker.NotifyClient(key, values[i]) {
				listLength++
				continue
			}
			listLength = newList.AppendL(values[i])
		}
		s.Set(key, newList)
		fmt.Fprint(out, FormatInteger(listLength))
		return nil
	}
	switch data := v.(type) {
	case *storage.ListType:
		newLength := 0
		for i := range values {
			if s.Blocker.NotifyClient(key, values[i]) {
				newLength++
				continue
			}
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

func lpop(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 1 {
		return errors.New("invalid number of arguments for lpop command")
	}

	key := args[0]

	v, found := s.Get(key)
	if !found {
		fmt.Fprint(out, FormatNullBulkString())
		return nil
	}
	switch data := v.(type) {
	case *storage.ListType:
		if data.Len == 0 {
			fmt.Fprint(out, FormatNullBulkString())
			return nil
		}
		n := 1
		var err error
		if len(args) > 1 {
			n, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid value for number of items to extract: %s", err)
			}
		}
		result := data.Lpop(n)
		if len(result) > 1 {
			fmt.Fprint(out, FormatArray(result))
		} else {
			fmt.Fprint(out, FormatBulkString(result[0]))
		}
		return nil
	default:
		return errors.New("value stored should be a list")
	}
}

func blpop(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 2 {
		return errors.New("invalid number of arguments for blpop command")
	}
	key := args[0]
	timeout := args[1]

	timeoutFloat, err := strconv.ParseFloat(timeout, 32)
	if err != nil {
		return fmt.Errorf("invalid value for timeout: %s", err)
	}

	data, found := s.Get(key)
	if found {
		if list, ok := data.(*storage.ListType); ok && list.Len > 0 {
			values := list.Lpop(1)
			fmt.Fprint(out, FormatArray([]string{key, values[0]}))
			return nil
		}
	}

	ch := s.Blocker.BlockedByKey(key)

	if timeout == "0" {
		v := <-ch
		fmt.Fprint(out, FormatArray([]string{key, v}))
		return nil
	} else {
		select {
		case v := <-ch:
			fmt.Fprint(out, FormatArray([]string{key, v}))
			return nil
		case <-time.After(time.Duration(timeoutFloat * float64(time.Second))):
			fmt.Fprint(out, FormatNullArray())
			return nil
		}
	}
}

func typeCmd(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 1 {
		return errors.New("invalid number of arguments for type command")
	}

	key := args[0]
	v, found := s.Get(key)
	if !found {
		fmt.Fprint(out, FormatSimpleString("none"))
		return nil
	}

	switch v.(type) {
	case *storage.StringType:
		fmt.Fprint(out, FormatSimpleString("string"))
	case *storage.ListType:
		fmt.Fprint(out, FormatSimpleString("list"))
	case *storage.StreamType:
		fmt.Fprint(out, FormatSimpleString("stream"))
	default:
		fmt.Fprint(out, FormatSimpleString("none"))
	}
	return nil
}

func xadd(_ io.Reader, out io.Writer, args []string, s *storage.Storage) error {
	if len(args) < 4 {
		return errors.New("invalid number of arguments for xadd command")
	}

	key := args[0]
	id := args[1]
	values := args[2:]

	if len(values)%2 != 0 {
		return errors.New("values need to be key value pairs")
	}

	v, found := s.Get(key)
	if !found {
		st := storage.NewStreamType(&storage.RadixNode{
			Key:   id,
			IsEnd: true,
			Value: values,
			Edges: nil,
		})
		s.Set(key, st)
		fmt.Fprint(out, FormatBulkString(id))
		return nil
	}

	switch data := v.(type) {
	case *storage.StreamType:
		data.Insert(data.Tree, id, values)
		fmt.Fprint(out, FormatBulkString(id))
		return nil
	default:
		return errors.New("value stored should be a stream")
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
	"lpop":   lpop,
	"blpop":  blpop,
	"type":   typeCmd,
	"xadd":   xadd,
}
