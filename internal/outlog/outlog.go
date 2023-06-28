package outlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

func Print(v ...any) {
	logger.Output(2, fmt.Sprint(v...))
}

func Printf(format string, v ...any) {
	logger.Output(2, fmt.Sprintf(format, v...))
}

func SetFlags(flag int) {
	logger.SetFlags(flag)
}

func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}

func ParseLogFlags(s string) (int, error) {
	var flags int
	words := strings.Split(s, "|")
	for _, word := range words {
		word = strings.TrimSpace(word)
		switch word {
		case "date":
			flags |= log.Ldate
		case "time":
			flags |= log.Ltime
		case "microseconds":
			flags |= log.Lmicroseconds
		case "longfile":
			flags |= log.Llongfile
		case "shortfile":
			flags |= log.Lshortfile
		case "UTC":
			flags |= log.LUTC
		case "msgprefix":
			flags |= log.Lmsgprefix
		case "stdFlags":
			flags |= log.LstdFlags
		}
	}
	return flags, nil
}

type LogFlags int

func (f LogFlags) String() string {
	var b strings.Builder

	flagsTable := []struct {
		flag int
		name string
	}{
		// LstdFlags must be the first element since LstdFlags = Ldate | Ltime
		{flag: log.LstdFlags, name: "stdFlags"},
		{flag: log.Ldate, name: "date"},
		{flag: log.Ltime, name: "time"},
		{flag: log.Lmicroseconds, name: "microseconds"},
		{flag: log.Llongfile, name: "longfile"},
		{flag: log.LUTC, name: "UTC"},
		{flag: log.Lmsgprefix, name: "msgprefix"},
	}
	rest := int(f)
	for _, t := range flagsTable {
		if rest&t.flag != 0 {
			if b.Len() > 0 {
				b.WriteByte('|')
			}
			b.WriteString(t.name)
		}
	}
	return b.String()
}
