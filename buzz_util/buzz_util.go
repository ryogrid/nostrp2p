package buzz_util

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const ServerImplVersion uint16 = 1

// ratio of post event attached profile update time
const AttachProfileUpdateProb = 0.2 // 1post / 5posts

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

var randGen *rand.Rand

func InitializeRandGen(seed int64) {
	seed_ := seed + time.Now().UnixNano()
	randGen = rand.New(rand.NewSource(seed_))
}

func GetRandUint64() uint64 {
	return randGen.Uint64()
}

// return true with given probability
func IsHit(prob float64) bool {
	return randGen.Float64() < prob
}

func GetLower64bitUint(bytes [32]byte) uint64 {
	return binary.LittleEndian.Uint64(bytes[:8])
}

func GzipCompless(data []byte) []byte {
	fmt.Println("GzipCompless:" + strconv.Itoa(len(data)))
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(data)
	if err != nil {
		panic(err)
	}

	if err2 := zw.Close(); err2 != nil {
		panic(err2)
	}

	retBuf := buf.Bytes()
	fmt.Println("GzipCompless:" + strconv.Itoa(len(retBuf)))
	return retBuf
}

func GzipDecompless(data []byte) []byte {
	buf := bytes.NewBuffer(data)
	zr, err := gzip.NewReader(buf)
	if err != nil {
		panic(err)
	}

	buf2 := new(bytes.Buffer)
	io.Copy(buf2, zr)
	if err2 := zr.Close(); err2 != nil {
		panic(err2)
	}

	retBuf := buf2.Bytes()
	return retBuf
}
