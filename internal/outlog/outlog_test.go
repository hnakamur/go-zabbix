package outlog

import (
	"log"
	"testing"
)

func TestParseLogFlags(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testCases := []struct {
			input string
			want  int
		}{
			{input: "", want: 0},
			{input: "stdFlags", want: log.LstdFlags},
			{input: "date | time", want: log.LstdFlags},
			{input: "stdFlags | microseconds", want: log.LstdFlags | log.Lmicroseconds},
		}
		for _, c := range testCases {
			got, err := ParseLogFlags(c.input)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				log.Printf("result mismatch, input=%s, got=0x%x, want=0x%x",
					c.input, got, c.want)
			}
		}
	})
	t.Run("error", func(t *testing.T) {
		testCases := []string{
			"stsFlags2",
			"date | | time",
		}
		for _, c := range testCases {
			if _, err := ParseLogFlags(c); err == nil {
				t.Errorf("want error but got no error, input=%s", c)
			}
		}
	})
}

func TestLogFlagsString(t *testing.T) {
	testCases := []struct {
		input int
		want  string
	}{
		{input: 0, want: ""},
		{input: log.LstdFlags, want: "stdFlags"},
		{input: log.Ldate | log.Ltime, want: "stdFlags"},
		{input: log.Ldate | log.Llongfile, want: "date|longfile"},
	}
	for _, c := range testCases {
		got := LogFlags(c.input).String()
		if got != c.want {
			t.Errorf("result mismatch, input=0x%x, got=%s, want=%s", c.input, got, c.want)
		}
	}
}
