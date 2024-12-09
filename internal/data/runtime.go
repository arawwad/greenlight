package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int64

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

func (r *Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	quotedJson := strconv.Quote(jsonValue)

	return []byte(quotedJson), nil
}

func (r *Runtime) UnmarshalJSON(input []byte) error {
	unquotedValue, err := strconv.Unquote(string(input))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	parts := strings.Split(unquotedValue, " ")

	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(i)

	return nil
}
