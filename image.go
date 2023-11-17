package imoto

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/fumiama/imgsz"
	"github.com/pkg/errors"
)

var (
	API = "https://imoto.seku.su/"
)

var (
	ErrInvalidURL = errors.New("invalid URL")
)

// Live judge if the image is exist
func Live(u string) bool {
	req, err := http.NewRequest("HEAD", u, nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

// Bed image to server
func Bed(t string, b []byte) (string, uint64, uint64, error) {
	_, _, err := imgsz.DecodeSize(bytes.NewReader(b))
	if err != nil {
		return "", 0, 0, errors.Wrap(err, getThisFuncName())
	}
	m := md5.Sum(b)
	u := API + hex.EncodeToString(m[:])
	req, err := http.NewRequest("PUT", u, bytes.NewReader(b))
	if err != nil {
		return "", 0, 0, errors.Wrap(err, getThisFuncName())
	}
	req.Header.Add("Authorization", t)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, 0, errors.Wrap(err, getThisFuncName())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return "", 0, 0, errors.New(getThisFuncName() + ": " + BytesToString(msg))
	}
	p, k := SplitMD5(m)
	return u[:len(u)-16], p, k, nil
}

// Use a URL once and delete it immediately
func Use(t string, u string, k uint64) ([]byte, error) {
	i := strings.LastIndex(u, "/")
	if i < 0 {
		return nil, errors.Wrap(ErrInvalidURL, getThisFuncName())
	}
	ms := u[i+1:]
	var m [md5.Size]byte
	switch len(ms) {
	case 32:
		n, err := hex.Decode(m[:], StringToBytes(ms))
		if err != nil {
			return nil, errors.Wrap(err, getThisFuncName())
		}
		if n != md5.Size {
			return nil, errors.Wrap(ErrInvalidURL, getThisFuncName())
		}
	case 16:
		n, err := hex.Decode(m[:8], StringToBytes(ms))
		if err != nil {
			return nil, errors.Wrap(err, getThisFuncName())
		}
		if n != 8 {
			return nil, errors.Wrap(ErrInvalidURL, getThisFuncName())
		}
		binary.LittleEndian.PutUint64(m[8:], k)
		u += Uint64String(k)
	default:
		return nil, errors.Wrap(ErrInvalidURL, getThisFuncName())
	}
	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, errors.Wrap(err, getThisFuncName())
	}
	req.Header.Add("Authorization", t)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, getThisFuncName())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, errors.New(getThisFuncName() + ": " + BytesToString(msg))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, getThisFuncName())
	}
	m2 := md5.Sum(data)
	if m2 != m {
		return nil, errors.New(getThisFuncName() + ": expect " + hex.EncodeToString(m[:]) + " but got " + hex.EncodeToString(m2[:]))
	}
	return data, nil
}
