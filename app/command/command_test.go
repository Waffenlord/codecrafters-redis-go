package command

import (
	"bytes"
	"testing"
)


func TestPing(t *testing.T) {
	var buf bytes.Buffer
	err := ping(nil, &buf, nil)
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
	err := echo(nil, &buf, []string{"Hola"})
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