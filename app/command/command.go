package command

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

type CmdToExecute struct {
	Name string
	Cmd  Builtin
	Args []string
}

type RedisClient struct {
	InTransaction bool
	TxQueue       []*CmdToExecute
}

type CommandContext struct {
	In           io.Reader
	Out          io.Writer
	Args         []string
	C            *RedisClient
	ServerConfig *config.Config
}

type Builtin func(ctx *CommandContext, s *storage.Storage) error

func ping(ctx *CommandContext, s *storage.Storage) error {
	fmt.Fprint(ctx.Out, FormatSimpleString("PONG"))
	return nil
}

func echo(ctx *CommandContext, s *storage.Storage) error {
	result := strings.Join(ctx.Args, "")
	fmt.Fprint(ctx.Out, FormatBulkString(result))
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

func set(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 2 {
		return errors.New("invalid number of arguments for set command")
	}
	key := ctx.Args[0]
	value := ctx.Args[1]
	data := storage.StringType{
		Value:     value,
		CreatedAt: time.Now(),
	}

	if len(ctx.Args) > 2 {
		currentIdx := 2
		for currentIdx < len(ctx.Args) {
			cmdArg := ctx.Args[currentIdx]
			cmd, ok := setOptionalArguments[strings.ToLower(cmdArg)]
			if !ok {
				return errors.New("invalid set optional argument")
			}
			if cmd.acceptsValue {
				currentIdx++
				if currentIdx >= len(ctx.Args) {
					return errors.New("missing value for set optional argument")
				}
				arg := ctx.Args[currentIdx]
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
	fmt.Fprint(ctx.Out, FormatSimpleString("OK"))
	return nil
}

func get(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 1 {
		return errors.New("key must be provided")
	}
	key := ctx.Args[0]

	v, found := s.Get(key)
	if !found {
		fmt.Fprint(ctx.Out, FormatNullBulkString())
		return nil
	}

	switch data := v.(type) {
	case *storage.StringType:
		if data.ExpMil > 0 {
			diff := time.Since(data.CreatedAt)
			if diff.Milliseconds() >= data.ExpMil {
				s.Delete(key)
				fmt.Fprint(ctx.Out, FormatNullBulkString())
				return nil
			}
		}
		fmt.Fprint(ctx.Out, FormatBulkString(data.Value))
		return nil
	}
	return nil
}

func rpush(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 2 {
		return errors.New("invalid number of arguments for rpush command")
	}

	key := ctx.Args[0]
	values := ctx.Args[1:]

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
		fmt.Fprint(ctx.Out, FormatInteger(listLength))
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
		fmt.Fprint(ctx.Out, FormatInteger(newLength))
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

func lrange(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 3 {
		return errors.New("invalid number of arguments for lrange command")
	}

	key := ctx.Args[0]
	startIdx := ctx.Args[1]
	endIdx := ctx.Args[2]

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
		fmt.Fprint(ctx.Out, FormatArray(make([]string, 0)))
		return nil
	}
	switch data := v.(type) {
	case *storage.ListType:
		cleanedStartIdx, cleanedEndIdx, isValid := isRangeValid(data.Len, startIdxInt, endIdxInt)
		if !isValid {
			fmt.Fprint(ctx.Out, FormatArray(make([]string, 0)))
			return nil
		}
		result := data.LRange(cleanedStartIdx, cleanedEndIdx)
		fmt.Fprint(ctx.Out, FormatArray(result))
		return nil
	default:
		return errors.New("value stored should be a list")
	}
}

func lpush(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 2 {
		return errors.New("invalid number of arguments for lpush command")
	}

	key := ctx.Args[0]
	values := ctx.Args[1:]

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
		fmt.Fprint(ctx.Out, FormatInteger(listLength))
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
		fmt.Fprint(ctx.Out, FormatInteger(newLength))
		return nil
	default:
		return errors.New("value stored should be a list")
	}
}

func llen(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 1 {
		return errors.New("invalid number of arguments for llen command")
	}

	key := ctx.Args[0]

	v, found := s.Get(key)
	if !found {
		fmt.Fprint(ctx.Out, FormatInteger(0))
		return nil
	}

	switch data := v.(type) {
	case *storage.ListType:
		data.Mux.RLock()
		defer data.Mux.RUnlock()
		fmt.Fprint(ctx.Out, FormatInteger(data.Len))
		return nil
	default:
		return errors.New("value stored should be a list")
	}
}

func lpop(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 1 {
		return errors.New("invalid number of arguments for lpop command")
	}

	key := ctx.Args[0]

	v, found := s.Get(key)
	if !found {
		fmt.Fprint(ctx.Out, FormatNullBulkString())
		return nil
	}
	switch data := v.(type) {
	case *storage.ListType:
		if data.Len == 0 {
			fmt.Fprint(ctx.Out, FormatNullBulkString())
			return nil
		}
		n := 1
		var err error
		if len(ctx.Args) > 1 {
			n, err = strconv.Atoi(ctx.Args[1])
			if err != nil {
				return fmt.Errorf("invalid value for number of items to extract: %s", err)
			}
		}
		result := data.Lpop(n)
		if len(result) > 1 {
			fmt.Fprint(ctx.Out, FormatArray(result))
		} else {
			fmt.Fprint(ctx.Out, FormatBulkString(result[0]))
		}
		return nil
	default:
		return errors.New("value stored should be a list")
	}
}

func blpop(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 2 {
		return errors.New("invalid number of arguments for blpop command")
	}
	key := ctx.Args[0]
	timeout := ctx.Args[1]

	timeoutFloat, err := strconv.ParseFloat(timeout, 32)
	if err != nil {
		return fmt.Errorf("invalid value for timeout: %s", err)
	}

	data, found := s.Get(key)
	if found {
		if list, ok := data.(*storage.ListType); ok && list.Len > 0 {
			values := list.Lpop(1)
			if len(values) > 0 {
				fmt.Fprint(ctx.Out, FormatArray([]string{key, values[0]}))
				return nil
			}

		}
	}

	ch := s.Blocker.BlockedByKey(key)

	if timeout == "0" {
		v := <-ch
		fmt.Fprint(ctx.Out, FormatArray([]string{key, v}))
		return nil
	} else {
		select {
		case v := <-ch:
			fmt.Fprint(ctx.Out, FormatArray([]string{key, v}))
			return nil
		case <-time.After(time.Duration(timeoutFloat * float64(time.Second))):
			fmt.Fprint(ctx.Out, FormatNullArray())
			return nil
		}
	}
}

func typeCmd(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 1 {
		return errors.New("invalid number of arguments for type command")
	}

	key := ctx.Args[0]
	v, found := s.Get(key)
	if !found {
		fmt.Fprint(ctx.Out, FormatSimpleString("none"))
		return nil
	}

	switch v.(type) {
	case *storage.StringType:
		fmt.Fprint(ctx.Out, FormatSimpleString("string"))
	case *storage.ListType:
		fmt.Fprint(ctx.Out, FormatSimpleString("list"))
	case *storage.StreamType:
		fmt.Fprint(ctx.Out, FormatSimpleString("stream"))
	default:
		fmt.Fprint(ctx.Out, FormatSimpleString("none"))
	}
	return nil
}

func xadd(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 4 {
		return errors.New("invalid number of arguments for xadd command")
	}

	key := ctx.Args[0]
	id := ctx.Args[1]
	values := ctx.Args[2:]

	if len(values)%2 != 0 {
		return errors.New("values need to be key value pairs")
	}

	v, found := s.Get(key)
	if !found {
		st := storage.NewStreamType()
		s.Set(key, st)
		v = st
	}

	switch data := v.(type) {
	case *storage.StreamType:
		id, err := data.Add(id, values)
		if err != nil {
			fmt.Fprint(ctx.Out, FormatSimpleError(genericError, err.Error()))
			return nil
		}
		s.Blocker.NotifyClient(key, "OK")
		fmt.Fprint(ctx.Out, FormatBulkString(id))
		return nil
	default:
		return errors.New("value stored should be a stream")
	}
}

func xrange(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 3 {
		return errors.New("invalid number of arguments for xrange command")
	}

	key := ctx.Args[0]
	startId := ctx.Args[1]
	endId := ctx.Args[2]

	v, found := s.Get(key)
	if !found {
		return errors.New("stream with the especified key not found")
	}

	switch data := v.(type) {
	case *storage.StreamType:
		results, err := data.XRange(startId, endId)
		if err != nil {
			return err
		}
		fmt.Fprint(ctx.Out, FormatStreamEntries(results))
		return nil
	default:
		return errors.New("value stored should be a stream")
	}
}

func xread(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 3 {
		return errors.New("invalid number of arguments for xread command")
	}

	parsedArgs, err := storage.ParseXReadArgs(ctx.Args)
	if err != nil {
		return err
	}
	var results []storage.XReadResult
	for i, key := range parsedArgs.StreamKeys {
		var ch chan string
		currentId := parsedArgs.StreamIds[i]
		v, found := s.Get(key)
		if !found {
			if parsedArgs.IsBlocked {
				ch = s.Blocker.BlockedByKey(key)
				if parsedArgs.BlockTime == 0 {
					<-ch
				} else {
					select {
					case <-ch:
					case <-time.After(time.Duration(parsedArgs.BlockTime) * time.Millisecond):
						fmt.Fprint(ctx.Out, FormatNullArray())
						return nil
					}
				}
				v, found = s.Get(key)
				if !found {
					return errors.New("stream with the especified key not found")
				}
			} else {
				return errors.New("stream with the especified key not found")
			}
		}
		switch data := v.(type) {
		case *storage.StreamType:
			var finalResult storage.XReadResult
			if currentId == "$" {
				currentId = data.LastEntryId
			}
			if parsedArgs.IsBlocked {
				ch = s.Blocker.BlockedByKey(key)
				var currentResult storage.XReadResult
				var err error
				if parsedArgs.BlockTime == 0 {
					<-ch
					currentResult, err = data.XRead(currentId, key)
					if err != nil {
						return err
					}
				} else {
					select {
					case <-ch:
						currentResult, err = data.XRead(currentId, key)
						if err != nil {
							return err
						}
					case <-time.After(time.Duration(parsedArgs.BlockTime) * time.Millisecond):
						fmt.Fprint(ctx.Out, FormatNullArray())
						return nil
					}
				}
				finalResult = currentResult
			} else {
				result, err := data.XRead(currentId, key)
				if err != nil {
					return err
				}
				finalResult = result
			}

			results = append(results, finalResult)
		default:
			return errors.New("value stored should be a stream")
		}
	}
	fmt.Fprint(ctx.Out, FormatXReadEntries(results))
	return nil
}

func incr(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) != 1 {
		return errors.New("invalid number of arguments for incr command")
	}

	key := ctx.Args[0]
	v, found := s.Get(key)
	if !found {
		data := storage.StringType{
			Value:     "1",
			CreatedAt: time.Now(),
		}
		s.Set(key, &data)
		fmt.Fprint(ctx.Out, FormatInteger(1))
		return nil
	}

	switch data := v.(type) {
	case *storage.StringType:
		parsed, err := strconv.Atoi(data.Value)
		if err != nil {
			fmt.Fprint(ctx.Out, FormatSimpleError(genericError, "value is not an integer or out of range"))
			return nil
		}
		parsed++
		data.Value = strconv.Itoa(parsed)
		fmt.Fprint(ctx.Out, FormatInteger(parsed))
		return nil
	default:
		return errors.New("value stored should be a string type")
	}
}

func multi(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) != 0 {
		return errors.New("multi command should not have any arguments")
	}
	if ctx.C.InTransaction {
		fmt.Fprint(ctx.Out, FormatSimpleError(genericError, "multi command already called"))
		return nil
	}
	ctx.C.InTransaction = true
	fmt.Fprint(ctx.Out, FormatSimpleString("OK"))
	return nil
}

func exec(ctx *CommandContext, s *storage.Storage) error {
	if !ctx.C.InTransaction {
		fmt.Fprint(ctx.Out, FormatSimpleError(genericError, "EXEC without MULTI"))
		return nil
	}
	if len(ctx.C.TxQueue) == 0 {
		ctx.C.InTransaction = false
		fmt.Fprint(ctx.Out, FormatArray(nil))
		return nil
	}
	fmt.Println("Executing transaction...")
	var row strings.Builder
	var results []string
	for _, cmd := range ctx.C.TxQueue {
		currentContext := CommandContext{
			In:   ctx.In,
			Out:  &row,
			Args: cmd.Args,
			C:    ctx.C,
		}
		err := cmd.Cmd(&currentContext, s)
		if err != nil {
			fmt.Fprint(currentContext.Out, FormatSimpleError(genericError, err.Error()))
		}
		results = append(results, row.String())
		row.Reset()
	}
	ctx.C.InTransaction = false
	fmt.Fprint(ctx.Out, FormatSimpleStringArray(results))
	return nil
}

func discard(ctx *CommandContext, s *storage.Storage) error {
	if !ctx.C.InTransaction {
		fmt.Fprint(ctx.Out, FormatSimpleError(genericError, "DISCARD without MULTI"))
		return nil
	}
	ctx.C.InTransaction = false
	ctx.C.TxQueue = nil
	fmt.Fprint(ctx.Out, FormatSimpleString("OK"))
	return nil
}

func info(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) > 0 {
		currentArg := ctx.Args[0]
		switch currentArg {
		case "replication":
			fmt.Fprint(
				ctx.Out, FormatBulkString(
					fmt.Sprintf(
						"role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d\r\n",
						ctx.ServerConfig.Role,
						ctx.ServerConfig.MasterReplId,
						ctx.ServerConfig.MasterReplOffset,
					)))
		}
	}
	return nil
}

func replconf(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 2 {
		return errors.New("invalid number of arguments for REPLCONF")
	}

	fmt.Fprint(ctx.Out, FormatSimpleString("OK"))
	return nil
}

func psync(ctx *CommandContext, s *storage.Storage) error {
	if len(ctx.Args) < 2 {
		return errors.New("invalid number of arguments for PSYNC")
	}

	replicationId := ctx.ServerConfig.MasterReplId

	fmt.Fprint(ctx.Out, FormatSimpleString(fmt.Sprintf("FULLRESYNC %s 0", replicationId)))
	return nil
}

var CommandMenu = map[string]Builtin{
	"echo":     echo,
	"ping":     ping,
	"set":      set,
	"get":      get,
	"rpush":    rpush,
	"lrange":   lrange,
	"lpush":    lpush,
	"llen":     llen,
	"lpop":     lpop,
	"blpop":    blpop,
	"type":     typeCmd,
	"xadd":     xadd,
	"xrange":   xrange,
	"xread":    xread,
	"incr":     incr,
	"multi":    multi,
	"exec":     exec,
	"discard":  discard,
	"info":     info,
	"replconf": replconf,
	"psync":    psync,
}
