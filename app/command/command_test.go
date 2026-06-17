package command

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

func TestPing(t *testing.T) {
	var buf bytes.Buffer
	err := ping(nil, &buf, nil, nil, nil)
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
	err := echo(nil, &buf, []string{"Hola"}, nil, nil)
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
	err := set(nil, &buf, []string{"key1"}, s, nil)
	if err == nil {
		t.Error("should throw an error for invalid number of arguments")
	}
	err = set(nil, &buf, []string{"key2", "test2"}, s, nil)
	if err != nil || buf.String() != FormatSimpleString("OK") {
		t.Error("should set the value and return OK")
	}
	err = set(nil, &buf, []string{"key2", "test2", "testExtension"}, s, nil)
	if err == nil {
		t.Error("should throw an error for invalid extension")
	}
	err = set(nil, &buf, []string{"key2", "test2", "PX"}, s, nil)
	if err == nil {
		t.Error("should throw an error for invalid extension value")
	}
	buf.Reset()
	err = set(nil, &buf, []string{"key2", "test2", "PX", "1000"}, s, nil)
	if err != nil {
		t.Errorf("should set the value with expiration: %s", err)
	}
}

func TestGet(t *testing.T) {
	var buf bytes.Buffer
	s := storage.NewStorage()
	err := get(nil, &buf, []string{}, s, nil)
	if err == nil {
		t.Error("should throw an error for invalid number of arguments")
	}
	err = set(nil, &buf, []string{"key1", "test1"}, s, nil)
	if err != nil || buf.String() != FormatSimpleString("OK") {
		t.Error("should set the value and return OK")
	}
	buf.Reset()
	err = get(nil, &buf, []string{"key1"}, s, nil)
	if err != nil || buf.String() != FormatBulkString("test1") {
		t.Error("should retrieve the value saved")
	}
	buf.Reset()
	err = get(nil, &buf, []string{"key2"}, s, nil)
	if err != nil || buf.String() != FormatNullBulkString() {
		t.Error("should return null string for missing key")
	}
}

func TestRpush(t *testing.T) {
	var buf bytes.Buffer
	s := storage.NewStorage()
	err := rpush(nil, &buf, []string{"key1"}, s, nil)
	if err == nil {
		t.Error("should throw an error for invalid number of arguments")
	}
	err = rpush(nil, &buf, []string{"key2", "value1"}, s, nil)
	if err != nil || buf.String() != FormatInteger(1) {
		t.Error("should create the list and append the value to the list")
	}
	buf.Reset()
	err = rpush(nil, &buf, []string{"key2", "value2", "value3"}, s, nil)
	if err != nil || buf.String() != FormatInteger(3) {
		t.Error("should append multiple values to the list")
	}
}

func TestIsValidRange(t *testing.T) {
	start, end, isValid := isRangeValid(5, 0, 3)
	if start != 0 {
		t.Errorf("should return the same start idx. Instead it is: %d", start)
	}
	if end != 3 {
		t.Errorf("should return the same end idx. Instead it is: %d", end)
	}
	if !isValid {
		t.Errorf("positive interval should be valid")
	}
	start, end, isValid = isRangeValid(5, 4, 3)
	if start != 0 || end != 0 || isValid {
		t.Errorf("should return 0 for indexes and false for validation: start %d end %d isValid %t", start, end, isValid)
	}
	start, end, isValid = isRangeValid(5, -4, -2)
	if start != 1 {
		t.Errorf("should return the correct positive start idx. Instead it is: %d", start)
	}
	if end != 3 {
		t.Errorf("should return the correct positive start idx. Instead it is: %d", end)
	}
	if !isValid {
		t.Errorf("negative interval should be valid")
	}

}

func TestLRange(t *testing.T) {
	var buf bytes.Buffer
	s := storage.NewStorage()
	err := rpush(nil, &buf, []string{"key1", "value1", "value2"}, s, nil)
	if err != nil || buf.String() != FormatInteger(2) {
		t.Error("should create the list and append the values to the list")
	}
	buf.Reset()
	err = lrange(nil, &buf, []string{}, s, nil)
	if err == nil {
		t.Error("should throw an error for invalid number of arguments")
	}
	err = lrange(nil, &buf, []string{"key1", "0", "1"}, s, nil)
	if err != nil || buf.String() != FormatArray([]string{"value1", "value2"}) {
		t.Error("should return the elements of the range provided")
	}
	buf.Reset()
	err = lrange(nil, &buf, []string{"key2", "0", "1"}, s, nil)
	if err != nil || buf.String() != FormatArray(make([]string, 0)) {
		t.Error("should return null for missing key")
	}
	buf.Reset()
}

func TestLpush(t *testing.T) {
	var buf bytes.Buffer
	s := storage.NewStorage()
	err := lpush(nil, &buf, []string{"key1"}, s, nil)
	if err == nil {
		t.Error("should throw an error for invalid number of arguments")
	}
	err = lpush(nil, &buf, []string{"key2", "value1"}, s, nil)
	if err != nil || buf.String() != FormatInteger(1) {
		t.Error("should create the list and append the value to the list")
	}
	buf.Reset()
	err = lpush(nil, &buf, []string{"key2", "value2", "value3"}, s, nil)
	if err != nil || buf.String() != FormatInteger(3) {
		t.Error("should append multiple values to the list")
	}
}

func TestLlen(t *testing.T) {
	var buf bytes.Buffer
	s := storage.NewStorage()
	err := llen(nil, &buf, []string{}, s, nil)
	if err == nil {
		t.Error("should throw an error for invalid number of arguments")
	}
	err = llen(nil, &buf, []string{"missing"}, s, nil)
	if err != nil || buf.String() != FormatInteger(0) {
		t.Error("should return 0 for missing key")
	}
	buf.Reset()
	err = rpush(nil, &buf, []string{"key1", "value1", "value2"}, s, nil)
	if err != nil || buf.String() != FormatInteger(2) {
		t.Error("should create the list and append the values to the list")
	}
	buf.Reset()
	err = llen(nil, &buf, []string{"key1"}, s, nil)
	if err != nil || buf.String() != FormatInteger(2) {
		t.Error("should return 2 as the length of the list")
	}

}

func TestLpop(t *testing.T) {
	var buf bytes.Buffer
	s := storage.NewStorage()
	err := rpush(nil, &buf, []string{"key1", "value1", "value2"}, s, nil)
	if err != nil || buf.String() != FormatInteger(2) {
		t.Error("should create the list and append the values to the list")
	}
	buf.Reset()
	err = lpop(nil, &buf, []string{"key1"}, s, nil)
	if err != nil {
		t.Error("error while removing first item")
	}
	if buf.String() != FormatBulkString("value1") {
		t.Error("invalid format for removed value")
	}
}

func TestBlpop(t *testing.T) {
	var listBuf bytes.Buffer
	s := storage.NewStorage()
	wg := sync.WaitGroup{}
	var resultWithoutTimeout string
	var resultWithTimeout string

	wg.Go(func() {
		var blpopBuf bytes.Buffer
		err := blpop(nil, &blpopBuf, []string{"key1", "0"}, s, nil)
		if err != nil {
			t.Error("error ocurred while removing item with 0 timeout")
		}
		resultWithoutTimeout = blpopBuf.String()
	})
	time.Sleep(time.Second * 1)
	err := rpush(nil, &listBuf, []string{"key1", "value1"}, s, nil)
	if err != nil || listBuf.String() != FormatInteger(1) {
		t.Errorf("error while pushing a new item. err: %s, value: %s", err, listBuf.String())
	}
	wg.Wait()
	if resultWithoutTimeout != FormatArray([]string{"key1", "value1"}) {
		t.Errorf("should return the value recently pushed. value: %s", resultWithoutTimeout)
	}
	listBuf.Reset()

	wg.Go(func() {
		var blpopBuf2 bytes.Buffer
		err := blpop(nil, &blpopBuf2, []string{"key1", "4"}, s, nil)
		if err != nil {
			t.Error("error ocurred while removing item with 4 sec timeout")
		}
		resultWithTimeout = blpopBuf2.String()
	})
	time.Sleep(time.Second * 2)
	err = rpush(nil, &listBuf, []string{"key1", "value2"}, s, nil)
	if err != nil || listBuf.String() != FormatInteger(1) {
		t.Errorf("error while pushing a new item. err: %s, value: %s", err, listBuf.String())
	}
	wg.Wait()
	if resultWithTimeout != FormatArray([]string{"key1", "value2"}) {
		t.Errorf("should return the value recently pushed after the timeout. value: %s", resultWithTimeout)
	}
}

func TestType(t *testing.T) {
	s := storage.NewStorage()
	var resultBuf bytes.Buffer

	err := set(nil, &resultBuf, []string{"key1", "value1"}, s, nil)
	if err != nil || resultBuf.String() != FormatSimpleString("OK") {
		t.Error("error ocurred while setting a string value")
	}
	resultBuf.Reset()
	err = typeCmd(nil, &resultBuf, []string{"key1"}, s, nil)
	if err != nil || resultBuf.String() != FormatSimpleString("string") {
		t.Error("invalid type for string")
	}
	resultBuf.Reset()
	err = rpush(nil, &resultBuf, []string{"key2", "value2"}, s, nil)
	if err != nil || resultBuf.String() != FormatInteger(1) {
		t.Error("error ocurred while appending a value to list")
	}
	resultBuf.Reset()
	err = typeCmd(nil, &resultBuf, []string{"key2"}, s, nil)
	if err != nil || resultBuf.String() != FormatSimpleString("list") {
		t.Error("invalid type for list")
	}
}
