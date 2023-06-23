package zabbix

import (
	"strconv"
	"strings"
)

// ID is an alias to uint64. ID is encoded as a string in JSON RPC
// requests and responses.
type ID uint64

// MarshalJSON returns a string representation.
func (i ID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatUint(uint64(i), 10) + `"`), nil
}

// UnmarshalJSON parses a string representation.
func (i *ID) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*i = ID(v)
	return nil
}

func (i ID) String() string {
	return strconv.FormatUint(uint64(i), 10)
}
