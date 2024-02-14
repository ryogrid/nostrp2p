package buzz_util

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

var DebugMode = false
var DenyWriteMode = false

type Stringset map[string]struct{}

func (ss Stringset) Set(value string) error {
	ss[value] = struct{}{}
	return nil
}

func (ss Stringset) String() string {
	return strings.Join(ss.Slice(), ",")
}

func (ss Stringset) Slice() []string {
	slice := make([]string, 0, len(ss))
	for k := range ss {
		slice = append(slice, k)
	}
	sort.Strings(slice)
	return slice
}

func OSInterrupt() {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	<-quit
}

func BuzzDbgPrintln(a ...interface{}) {
	if DebugMode {
		fmt.Fprintln(os.Stderr, a...)
	}
}

// Assert fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		tb.Errorf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}
