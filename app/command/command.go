package command

import (
	"fmt"
	"io"
	"strings"
)


type Builtin func(in io.Reader, out io.Writer, args []string) error

func ping(_ io.Reader, out io.Writer, _ []string) error {
	fmt.Fprint(out, FormatSimpleString("PONG"))
	return nil
}

func echo(_ io.Reader, out io.Writer, args []string) error {
	result := strings.Join(args, "")
	fmt.Fprint(out, FormatBulkString(result))
	return nil
}

var CommandMenu = map[string]Builtin{
	"echo": echo,
	"ping": ping,
}