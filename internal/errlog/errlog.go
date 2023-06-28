package errlog

import (
	"fmt"
	"io"
	"log"
	"os"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

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
