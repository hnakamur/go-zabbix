package zabbix

import (
	"testing"
	"time"
)

func TestTimestampMarshalJSON(t *testing.T) {
	ts := Timestamp(time.Date(2016, 3, 28, 8, 7, 23, 0, time.UTC))
	actual, err := ts.MarshalJSON()
	if err != nil {
		t.Errorf("Timestamp.MarshalJSON failed with err %s", err)
	}
	expected := `"1459152443"`
	if string(actual) != expected {
		t.Errorf("Timestamp.MarshalJSON got %s; wanted %s", string(actual), expected)
	}
}

func TestTimestampUnmarshalJSON(t *testing.T) {
	var ts Timestamp
	err := ts.UnmarshalJSON([]byte(`"1459152443"`))
	if err != nil {
		t.Errorf("Timestamp.UnrshalJSON failed with err %s", err)
	}
	expected := time.Date(2016, 3, 28, 8, 7, 23, 0, time.UTC)
	if !time.Time(ts).Equal(expected) {
		t.Errorf("Timestamp.UnmarshalJSON got %v; wanted %v", ts, expected)
	}
}
