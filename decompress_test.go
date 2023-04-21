package compress_test

import (
	"bytes"
	"fmt"
	"github.com/andybalholm/brotli"
	compress "github.com/lf4096/gin-compress"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

// disable all algos
var dcOpts = []compress.CompressOption{
	compress.WithAlgo("gzip", false),
	compress.WithAlgo("br", false),
	compress.WithAlgo("zstd", false),
	compress.WithAlgo("deflate", false),
}

var lol = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAH"
var lolLarge = strings.Repeat("A", 600*1024)

func TestDecompressBrotli(t *testing.T) {
	r := setupRouter(dcOpts...)

	b := bytes.NewBuffer(nil)
	bc := brotli.NewWriter(b)
	_, err := bc.Write([]byte(lol))
	assert.NoError(t, err)
	err = bc.Close()
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/echo", b)
	req.Header.Set("Content-Encoding", "br")
	r.ServeHTTP(w, req)

	assert.Equal(t, "200", fmt.Sprintf("%v", w.Code))
	assert.Equal(t, lol, w.Body.String())

}

func TestDecompressGzip(t *testing.T) {
	r := setupRouter(dcOpts...)

	b := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(b)
	_, err := gz.Write([]byte(lol))
	assert.NoError(t, err)
	err = gz.Close()
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/echo", b)
	req.Header.Set("Content-Encoding", "gzip")
	r.ServeHTTP(w, req)

	assert.Equal(t, "", w.Header().Get("X-Request-Content-Encoding"))
	assert.Equal(t, "200", fmt.Sprintf("%v", w.Code))
	assert.Equal(t, lol, w.Body.String())

}

func TestDecompressZstd(t *testing.T) {
	r := setupRouter(dcOpts...)

	b := bytes.NewBuffer(nil)
	z, err := zstd.NewWriter(b)
	assert.NoError(t, err)
	_, err = z.Write([]byte(lol))
	assert.NoError(t, err)
	err = z.Close()
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/echo", b)
	req.Header.Set("Content-Encoding", "zstd")
	r.ServeHTTP(w, req)

	assert.Equal(t, "", w.Header().Get("X-Request-Content-Encoding"))
	assert.Equal(t, "200", fmt.Sprintf("%v", w.Code))
	assert.Equal(t, lol, w.Body.String())

}

func TestDecompressDeflate(t *testing.T) {
	r := setupRouter(dcOpts...)

	b := bytes.NewBuffer(nil)
	z := zlib.NewWriter(b)
	_, err := z.Write([]byte(lol))
	assert.NoError(t, err)
	err = z.Close()
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/echo", b)
	req.Header.Set("Content-Encoding", "deflate")
	r.ServeHTTP(w, req)

	assert.Equal(t, "", w.Header().Get("X-Request-Content-Encoding"))
	assert.Equal(t, "200", fmt.Sprintf("%v", w.Code))
	assert.Equal(t, lol, w.Body.String())

}

func TestMultipleDecompressions1(t *testing.T) {
	// do not alter the limit
	r := setupRouter(dcOpts...)

	b := bytes.NewBuffer(nil)
	z := zlib.NewWriter(b)
	g := gzip.NewWriter(z)
	_, err := g.Write([]byte(lolLarge))
	assert.NoError(t, err)
	err = g.Close()
	assert.NoError(t, err)
	err = z.Close()
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/echo", b)
	req.Header.Set("Content-Encoding", "gzip, deflate")
	r.ServeHTTP(w, req)

	assert.Equal(t, "gzip", w.Header().Get("X-Request-Content-Encoding"))
	assert.Equal(t, "200", fmt.Sprintf("%v", w.Code))
}

func TestMultipleDecompressions2(t *testing.T) {
	var sDcOpts = []compress.CompressOption{
		compress.WithAlgo("gzip", false),
		compress.WithAlgo("br", false),
		compress.WithAlgo("zstd", false),
		compress.WithAlgo("deflate", false),
		compress.WithMaxDecodeSteps(2),
	}
	r := setupRouter(sDcOpts...)

	b := bytes.NewBuffer(nil)
	z := zlib.NewWriter(b)
	g := gzip.NewWriter(z)
	_, err := g.Write([]byte(lolLarge))
	assert.NoError(t, err)
	err = g.Close()
	assert.NoError(t, err)
	err = z.Close()
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/echo", b)
	req.Header.Set("Content-Encoding", "gzip, deflate")
	r.ServeHTTP(w, req)

	assert.Equal(t, "", w.Header().Get("X-Request-Content-Encoding"))
	assert.Equal(t, "200", fmt.Sprintf("%v", w.Code))
	assert.Equal(t, lolLarge, w.Body.String())
}

func TestMultipleDecompressions3(t *testing.T) {
	var sDcOpts = []compress.CompressOption{
		compress.WithAlgo("gzip", false),
		compress.WithAlgo("br", false),
		compress.WithAlgo("zstd", false),
		compress.WithAlgo("deflate", false),
		compress.WithMaxDecodeSteps(4),
	}
	r := setupRouter(sDcOpts...)

	b := bytes.NewBuffer(nil)
	z := zlib.NewWriter(b)
	g := gzip.NewWriter(z)
	_, err := g.Write([]byte(lolLarge))
	assert.NoError(t, err)
	err = g.Close()
	assert.NoError(t, err)
	err = z.Close()
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/echo", b)
	req.Header.Set("Content-Encoding", "gzipButDifferentLol, deflate")
	r.ServeHTTP(w, req)

	assert.Equal(t, "gzipButDifferentLol", w.Header().Get("X-Request-Content-Encoding"))
	assert.Equal(t, "200", fmt.Sprintf("%v", w.Code))
}
