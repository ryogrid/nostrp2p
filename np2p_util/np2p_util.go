package np2p_util

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/ryogrid/nostrp2p/np2p_const"
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

var DebugMode = false

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

func Np2pDbgPrintln(a ...interface{}) {
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

func GetLower64bitUint(bytes [np2p_const.PubkeySize]byte) uint64 {
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

func Gen256bitHash(data []byte) [32]byte {
	hf := sha256.New()
	hf.Write(data)
	var ret [32]byte
	copy(ret[:], hf.Sum(nil)[:32])
	return ret
}

func ExtractUint64FromBytes(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

func Get6ByteUint64FromHexPubKeyStr(pubKeyStr string) uint64 {
	pubKeyBytes, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		fmt.Printf("public key: %s: %v\n", pubKeyStr, err)
	}

	return binary.BigEndian.Uint64(pubKeyBytes[len(pubKeyBytes)-8:]) & 0x0000ffffffffffff
}

func GetUint64FromHexPubKeyStr(pubKeyStr string) uint64 {
	pubKeyBytes, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		fmt.Printf("public key: %s: %v\n", pubKeyStr, err)
	}

	return binary.BigEndian.Uint64(pubKeyBytes[len(pubKeyBytes)-8:])
}

func StrTo32BytesArr(pubKeyStr string) [32]byte {
	bytes, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		panic(err)
	}
	byteArr32 := [32]byte{}
	copy(byteArr32[:], bytes)

	return byteArr32
}
