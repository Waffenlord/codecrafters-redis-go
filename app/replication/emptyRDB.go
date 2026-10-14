package replication

import (
	"encoding/base64"
	"fmt"
)

const emptyRDB = "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="

func GetEmptyRDB() ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(emptyRDB)
	if err != nil {
		return nil, fmt.Errorf("decoding empty RDB: %w", err)
	}

	return data, nil
}
