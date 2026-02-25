package command

import "fmt"


func FormatSimpleString(i string) string {
	return fmt.Sprintf("+%s\r\n", i)
}

func FormatBulkString(i string) string {
	l := len(i)
	return fmt.Sprintf("$%d\r\n%s\r\n", l, i)
}

func FormatNullBulkString() string {
	return fmt.Sprint("$-1\r\n")
}