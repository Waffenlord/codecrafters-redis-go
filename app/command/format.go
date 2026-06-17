package command

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

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
		return fmt.Sprintf(":%d\r\n", i)
	}
	return fmt.Sprintf(":%d\r\n", i)
}

func FormatArray(l []string) string {
	var result strings.Builder
	fmt.Fprintf(&result, "*%d\r\n", len(l))
	for i := range l {
		fmt.Fprintf(&result, "$%d\r\n%s\r\n", len(l[i]), l[i])
	}
	return result.String()
}

func FormatNullArray() string {
	return "*-1\r\n"
}

type ErrorType string

const (
	genericError ErrorType = "ERR"
)

func FormatSimpleError(errorType ErrorType, message string) string {
	return fmt.Sprintf("-%s %s\r\n", errorType, message)
}

func FormatStreamEntries(entries []storage.NodeResult) string {
	var result strings.Builder
	fmt.Fprintf(&result, "*%d\r\n", len(entries))
	for i := range entries {
		fmt.Fprint(&result, "*2\r\n")
		fmt.Fprint(&result, FormatBulkString(entries[i].Id))
		fmt.Fprint(&result, FormatArray(entries[i].Values))
	}
	return result.String()
}

func FormatXReadEntries(entries []storage.XReadResult) string {
	var result strings.Builder
	fmt.Fprintf(&result, "*%d\r\n", len(entries))
	for i := range entries {
		fmt.Fprint(&result, "*2\r\n")
		fmt.Fprint(&result, FormatBulkString(entries[i].StreamKey))
		fmt.Fprint(&result, FormatStreamEntries(entries[i].Results))
	}
	return result.String()
}

func FormatSimpleStringArray(l []string) string {
	var result strings.Builder
	fmt.Fprintf(&result, "*%d\r\n", len(l))
	for i := range l {
		fmt.Fprintf(&result, "%s", l[i])
	}
	return result.String()
}
