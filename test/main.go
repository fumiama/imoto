package main

import (
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
	imoto.API = "http://127.0.0.1:8000/"
	token := "0000000000000000000000000000000000000000000000000000000000000000"
	u, _, k, err := imoto.Bed(token, data)
	if err != nil {
		panic(err)
	}
	isexist := imoto.Live(u)
	if !isexist {
		panic("HEAD")
	}
	req, err := http.NewRequest("GET", u, nil)
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
		panic("GET error: " + imoto.BytesToString(msg))
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
	_, err = imoto.Use(token, u, k)
	if err != nil {
		panic(err)
	}
}
