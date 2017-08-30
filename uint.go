package poloniex

import (
	"strconv"
	"strings"
)

type convertibleUint uint64

func (cs *convertibleUint) UnmarshalJSON(data []byte) error {
	asString := string(data)
	asString = strings.Replace(asString, "\"", "", 2)
	intVal, err := strconv.Atoi(asString)
	if err != nil {
		return err
	}

	*cs = convertibleUint(intVal)
	return nil
}
