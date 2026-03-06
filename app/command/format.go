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
	return "$-1\r\n"
}

func FormatInteger(i int) string {
	if i < 0 {
		return fmt.Sprintf(":%s%d\r\n", "-", i)
	}
	return fmt.Sprintf(":%d\r\n", i)
}
