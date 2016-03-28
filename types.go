package zabbix

import (
	"strconv"
	"strings"
	"time"
)

type Timestamp time.Time

// MarshalJSON returns a string of seconds from the Unix Epoch time.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(time.Time(t).Unix(), 10) + `"`), nil
}

// UnmarshalJSON reads a string of seconds from the Unix Epoch time.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*t = Timestamp(time.Unix(ts, 0))
	return nil
}

func (t Timestamp) String() string {
	return time.Time(t).Format(time.RFC3339)
}
