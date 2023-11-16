package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"io"
	"net/http"
	"time"

	"github.com/FloatTech/ttl"
	"github.com/fumiama/imgsz"
	"github.com/sirupsen/logrus"

	"github.com/fumiama/imoto"
)

var imgcache *ttl.Cache[uint64, *imagebody]

type imagebody struct {
	key uint64
	typ string
	dat []byte
}

func main() {
	cachetime := flag.Uint("t", 60, "cache time (s)")
	endpoint := flag.String("e", "127.0.0.1:8000", "listening endpoint")
	flag.Parse()
	imgcache = ttl.NewCache[uint64, *imagebody](time.Second * time.Duration(*cachetime))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		m, err := imoto.GetMD5(r.URL.Path)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "400 Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}
		p, k := imoto.SplitMD5(m)
		logrus.Infoln("[handle]", r.Method, r.URL.Path)
		switch r.Method {
		case http.MethodHead:
			if imgcache.Get(p) == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			img := imgcache.Get(p)
			if img == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "image/"+img.typ)
			_, _ = w.Write(img.dat)
		case http.MethodPut:
			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			realm := md5.Sum(data)
			if m != realm {
				http.Error(w, "400 Bad Request: file md5 mismatch", http.StatusBadRequest)
				return
			}
			_, typ, err := imgsz.DecodeSize(bytes.NewReader(data))
			if err != nil {
				http.Error(w, "400 Bad Request: "+err.Error(), http.StatusBadRequest)
				return
			}
			imgcache.Set(p, &imagebody{
				key: k,
				typ: typ,
				dat: data,
			})
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			img := imgcache.Get(p)
			if img == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if k != img.key {
				http.Error(w, "403 Forbidden", http.StatusForbidden)
				return
			}
			imgcache.Delete(p)
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "image/"+img.typ)
			_, _ = w.Write(img.dat)
		default:
			http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	logrus.Errorln(http.ListenAndServe(*endpoint, nil))
}
