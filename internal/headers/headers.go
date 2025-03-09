package headers

import (
	"bytes"
	"fmt"
	"log"
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

func (h Headers) SetKey(key, val string) {
	h[strings.ToLower(key)] = val
}

func (h Headers) AddKey(key, val string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		h[key] = v + ", " + val
		return
	}
	h[key] = val
	return
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	loc := bytes.Index(data, []byte("\r\n"))
	if loc != -1 {
		log.Print(loc)
		log.Print(string(data[:loc]))
		if loc == 0 {
			return 2, true, nil
		}

		headerKey, headerVal, err := ParseHeaderLine(string(data[:loc]))
		if err != nil {
			return 0, false, fmt.Errorf("Failed to Parse Request line with err %v", err)

		}
		h.AddKey(headerKey, headerVal)

		return loc + 2, false, nil
	}
	log.Print(loc)
	return 0, false, nil

}

func ParseHeaderLine(headerLine string) (string, string, error) {
	log.Printf("Parsing Header line: %s", headerLine)
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
	log.Printf("Parsed Header and added key, value: %s: %s", key, val)
	return key, val, nil
}

func NewHeaders() Headers {
	return Headers{}
}
