package grpc

import (
	"encoding/base64"
	"strconv"
)

func encodeOffset(offset int) string {
	s := strconv.Itoa(offset)
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func decodeOffset(token string) (int, error) {
	if token == "" {
		return 0, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0, err
	}

	n, err := strconv.Atoi(string(decoded))
	if err != nil {
		return 0, err
	}

	return n, nil
}
