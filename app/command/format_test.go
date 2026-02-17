package command

import "testing"


func TestFormatSimpleString(t *testing.T) {
	result := FormatSimpleString("OK")
	if result != "+OK\r\n" {
		t.Errorf("incorrect format: %s", result)
	}
}

func TestFormatBulkString(t *testing.T) {
	result := FormatBulkString("golang")
	if result != "$6\r\ngolang\r\n" {
		t.Errorf("incorrect format: %s", result)
	}
}