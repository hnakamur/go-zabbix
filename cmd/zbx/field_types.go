package main

import (
	"strconv"
	"time"
)

type Timestamp time.Time

func ParseTimestamp(epochSeconds string) (Timestamp, error) {
	ts, err := strconv.ParseInt(epochSeconds, 10, 64)
	if err != nil {
		return Timestamp{}, err
	}
	return Timestamp(time.Unix(ts, 0)), nil
}

func (t Timestamp) String() string {
	tt := time.Time(t)
	return strconv.FormatInt(tt.Unix(), 10)
}

type Seconds time.Duration

func ParseSeconds(seconds string) (Seconds, error) {
	s, err := strconv.ParseInt(seconds, 10, 64)
	if err != nil {
		return 0, err
	}
	return Seconds(time.Duration(s) * time.Second), nil
}

func (s Seconds) String() string {
	seconds := int64(time.Duration(s) / time.Second)
	return strconv.FormatInt(seconds, 10)
}
