package poloniex

import (
	"fmt"
	"strings"
)

type convertibleBool bool

func (bit *convertibleBool) UnmarshalJSON(data []byte) error {
	asString := string(data)
	asString = strings.Trim(asString, `"`)
	if asString == "1" || asString == "true" {
		*bit = true
	} else if asString == "0" || asString == "false" {
		*bit = false
	} else {
		return fmt.Errorf("boolean unmarshal error: invalid input %s", asString)
	}
	return nil
}
