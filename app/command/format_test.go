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

func TestFormatInteger(t *testing.T) {
	result := FormatInteger(-5)
	if result != ":-5\r\n" {
		t.Errorf("incorret negative int format: %s", result)
	}
	result = FormatInteger(4)
	if result != ":4\r\n" {
		t.Errorf("incorrect positive int format: %s", result)
	}
}

func TestFormatArray(t *testing.T) {
	result := FormatArray([]string{})
	if result != "*0\r\n" {
		t.Errorf("incorrect empty array format: %s", result)
	}
	result = FormatArray([]string{"hola", "mundo"})
	if result != "*2\r\n$4\r\nhola\r\n$5\r\nmundo\r\n" {
		t.Errorf("incorrect array format: %s", result)
	}

}
