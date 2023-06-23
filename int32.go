package zabbix

import (
	"strconv"
	"strings"
)

// Int32 is an alias to int32. Int32 is encoded as a string in JSON RPC
// requests and responses.
type Int32 int32

// MarshalJSON returns a string representation.
func (i Int32) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(int64(i), 10) + `"`), nil
}

// UnmarshalJSON parses a string representation.
func (i *Int32) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*i = Int32(v)
	return nil
}

func (i Int32) String() string {
	return strconv.FormatInt(int64(i), 10)
}
