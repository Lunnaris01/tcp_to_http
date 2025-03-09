package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func (h Headers) Get(key string) (string, error) {
	val, ok := h[strings.ToLower(key)]
	if !ok {
		return "", fmt.Errorf("Key not found!")
	}
	return val, nil
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if loc := bytes.Index(data, []byte("\r\n")); loc != -1 {
		if loc == 0 {
			return 2, true, nil
		}

		headerKey, headerVal, err := ParseHeaderLine(string(data[:loc]))
		if err != nil {
			return 0, false, fmt.Errorf("Failed to Parse Request line with err %v", err)

		}

		val, ok := h[headerKey]
		if ok {
			headerVal = val + ", " + headerVal
		}
		h[headerKey] = headerVal

		return loc + 2, false, nil
	}
	return 0, false, nil

}

func ParseHeaderLine(headerLine string) (string, string, error) {
	trimmedline := strings.TrimSpace(headerLine)
	splitLine := strings.SplitN(trimmedline, ":", 2)
	if len(splitLine) != 2 {
		return "", "", fmt.Errorf("Failed to split headerline")
	}
	key := splitLine[0]
	if key != strings.TrimSpace(key) {
		return "", "", fmt.Errorf("Malformed key (maybe a whitespace before the :?)")
	}
	key = strings.ToLower(key)
	for _, c := range key {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyz1234567890!#$%&'*+-.^_`|~", c) {
			return "", "", fmt.Errorf("Illegal Character in Keystring")
		}
	}
	val := strings.TrimSpace(splitLine[1])

	return key, val, nil
}

func NewHeaders() Headers {
	return Headers{}
}
