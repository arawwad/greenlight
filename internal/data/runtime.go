package data

import (
	"fmt"
	"strconv"
)

type Runtime int64

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	quotedJson := strconv.Quote(jsonValue)

	return []byte(quotedJson), nil
}
