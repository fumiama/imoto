package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os"

	"github.com/fumiama/imoto"
)

func main() {
	data, err := os.ReadFile("test.jpeg")
	if err != nil {
		panic(err)
	}
	m := md5.Sum(data)
	p, _ := imoto.SplitMD5(m)
	req, err := http.NewRequest("PUT", "http://127.0.0.1:8000/"+hex.EncodeToString(m[:]), bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		panic("PUT error: " + imoto.BytesToString(msg))
	}
	req, err = http.NewRequest("HEAD", "http://127.0.0.1:8000/"+imoto.Uint64String(p), nil)
	if err != nil {
		panic(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		panic("HEAD error: " + imoto.BytesToString(msg))
	}
	req, err = http.NewRequest("GET", "http://127.0.0.1:8000/"+imoto.Uint64String(p), nil)
	if err != nil {
		panic(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		panic("HEAD error: " + imoto.BytesToString(msg))
	}
	h := md5.New()
	_, err = io.Copy(h, resp.Body)
	if err != nil {
		panic(err)
	}
	var m2 [md5.Size]byte
	h.Sum(m2[:0])
	if m2 != m {
		panic("GET error: expected " + hex.EncodeToString(m[:]) + " but got " + hex.EncodeToString(m2[:]))
	}
	req, err = http.NewRequest("DELETE", "http://127.0.0.1:8000/"+hex.EncodeToString(m[:]), nil)
	if err != nil {
		panic(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		panic("HEAD error: " + imoto.BytesToString(msg))
	}
	h = md5.New()
	_, err = io.Copy(h, resp.Body)
	if err != nil {
		panic(err)
	}
	h.Sum(m2[:0])
	if m2 != m {
		panic("DELETE error: expected " + hex.EncodeToString(m[:]) + " but got " + hex.EncodeToString(m2[:]))
	}
}
