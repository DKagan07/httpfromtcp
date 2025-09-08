package headers

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}

	// If the data starts with a CRLF, then we return true...
	if idx == 0 {
		// ...and we want to consume the CRLF, so we return 2
		return 2, true, nil
	}

	fields := bytes.SplitN(data[:idx], []byte(":"), 2)
	if len(fields) != 2 {
		return 0, false, errors.New("malformed field-line")
	}

	fn := fields[0]
	if bytes.HasSuffix(fn, []byte(" ")) {
		return 0, false, errors.New("misformatted field-line")
	}
	fieldName := strings.ToLower(string(fn))

	isValid := isValidFieldName(fieldName)
	if !isValid {
		return 0, false, errors.New("invalid character in field value")
	}

	fv := bytes.TrimSpace(fields[1])
	fieldValue := string(fv)

	if a, ok := h[fieldName]; ok {
		fieldValue = fmt.Sprintf("%s, %s", a, fieldValue)
	}
	h[fieldName] = fieldValue

	// Need to add 2 for CRLF
	return idx + 2, false, nil
}

func isValidFieldName(fieldName string) bool {
	if len(fieldName) < 1 {
		return false
	}

	// all lower case, all upper case, all numeric, and specific special characters
	regex := regexp.MustCompile(`^[a-zA-Z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`)
	return regex.MatchString(fieldName)
}

// Get accepts a case-insensitive key to a header and returns the value
func (h Headers) Get(key string) string {
	l_key := strings.ToLower(key)
	v, ok := h[l_key]
	if !ok {
		return ""
	}

	return v
}
