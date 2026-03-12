package command

import (
	"bytes"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func TestPing(t *testing.T) {
	var buf bytes.Buffer
	err := ping(nil, &buf, nil, nil)
	if err != nil {
		t.Errorf("Error executing ping command: %s", err)
	}
	if buf.Len() == 0 {
		t.Error("invalid output length")
	}

	if buf.String() != "+PONG\r\n" {
		t.Error("incorrect output string")
	}
}

func TestEcho(t *testing.T) {
	var buf bytes.Buffer
	err := echo(nil, &buf, []string{"Hola"}, nil)
	if err != nil {
		t.Errorf("Error executing echo command: %s", err)
	}
	if buf.Len() == 0 {
		t.Error("invalid output length")
	}
	if buf.String() != "$4\r\nHola\r\n" {
		t.Error("incorrect output string")
	}
}

func TestSet(t *testing.T) {
	var buf bytes.Buffer
	s := storage.NewStorage()
	err := set(nil, &buf, []string{"key1"}, s)
	if err == nil {
		t.Error("should throw an error for invalid number of arguments")
	}
	err = set(nil, &buf, []string{"key2", "test2"}, s)
	if err != nil || buf.String() != FormatSimpleString("OK") {
		t.Error("should set the value and return OK")
	}
	err = set(nil, &buf, []string{"key2", "test2", "testExtension"}, s)
	if err == nil {
		t.Error("should throw an error for invalid extension")
	}
	err = set(nil, &buf, []string{"key2", "test2", "PX"}, s)
	if err == nil {
		t.Error("should throw an error for invalid extension value")
	}
	buf.Reset()
	err = set(nil, &buf, []string{"key2", "test2", "PX", "1000"}, s)
	if err != nil {
		t.Errorf("should set the value with expiration: %s", err)
	}
}

func TestGet(t *testing.T) {
	var buf bytes.Buffer
	s := storage.NewStorage()
	err := get(nil, &buf, []string{}, s)
	if err == nil {
		t.Error("should throw an error for invalid number of arguments")
	}
	err = set(nil, &buf, []string{"key1", "test1"}, s)
	if err != nil || buf.String() != FormatSimpleString("OK") {
		t.Error("should set the value and return OK")
	}
	buf.Reset()
	err = get(nil, &buf, []string{"key1"}, s)
	if err != nil || buf.String() != FormatBulkString("test1") {
		t.Error("should retrieve the value saved")
	}
	buf.Reset()
	err = get(nil, &buf, []string{"key2"}, s)
	if err != nil || buf.String() != FormatNullBulkString() {
		t.Error("should return null string for missing key")
	}
}

func TestRpush(t *testing.T) {
	var buf bytes.Buffer
	s := storage.NewStorage()
	err := rpush(nil, &buf, []string{"key1"}, s)
	if err == nil {
		t.Error("should throw an error for invalid number of arguments")
	}
	err = rpush(nil, &buf, []string{"key2", "value1"}, s)
	if err != nil || buf.String() != FormatInteger(1) {
		t.Error("should create the list and append the value to the list")
	}
	buf.Reset()
	err = rpush(nil, &buf, []string{"key2", "value2", "value3"}, s)
	if err != nil || buf.String() != FormatInteger(3) {
		t.Error("should append multiple values to the list")
	}
}
