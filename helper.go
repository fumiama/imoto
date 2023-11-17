package imoto

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

func getFuncAndFileNameWithSkip(n int) (string, string) {
	pc, fn, _, ok := runtime.Caller(n)
	if !ok {
		return "", ""
	}
	i := strings.LastIndex(fn, "/") + 1
	if i > 0 {
		fn = strings.TrimSuffix(fn[i:], ".go")
	}
	fullname := runtime.FuncForPC(pc).Name()
	i = strings.LastIndex(fullname, ".") + 1
	if i <= 0 || i >= len(fullname) {
		return fullname, fn
	}
	return fullname[i:], fn
}

// getThisFuncName 获取正在执行的函数名
func getThisFuncName() string {
	x, _ := getFuncAndFileNameWithSkip(2)
	return x
}

func GetMD5(u string) (m [md5.Size]byte, err error) {
	u = strings.Trim(u, "/ ?&\n\t")
	if len(u) != md5.Size*2 && len(u) != md5.Size {
		err = errors.New(getThisFuncName() + ": invalid path len: " + strconv.Itoa(len(u)))
		return
	}
	_, err = hex.Decode(m[:], StringToBytes(u))
	return
}

func SplitMD5(m [md5.Size]byte) (path uint64, key uint64) {
	path = binary.LittleEndian.Uint64(m[:8])
	key = binary.LittleEndian.Uint64(m[8:])
	return
}

func Uint64String(x uint64) string {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], x)
	return hex.EncodeToString(buf[:])
}

// slice is the runtime representation of a slice.
// It cannot be used safely or portably and its representation may
// change in a later release.
//
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type slice struct {
	data unsafe.Pointer
	len  int
	cap  int
}

// BytesToString 没有内存开销的转换
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes 没有内存开销的转换
func StringToBytes(s string) (b []byte) {
	bh := (*slice)(unsafe.Pointer(&b))
	sh := (*slice)(unsafe.Pointer(&s))
	bh.data = sh.data
	bh.len = sh.len
	bh.cap = sh.len
	return b
}
