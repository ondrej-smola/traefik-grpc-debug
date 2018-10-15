package util

import (
	"fmt"
	"os"
)

func Exit(msg interface{}) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
