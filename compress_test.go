package compress_test

import (
	"bytes"
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/zstd"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	compress "github.com/lf4096/gin-compress"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

var smallBody = "SMALL BODY"
var largeBody = strings.Repeat("LARGE BODY", 256)

func setupRouter(opts ...compress.CompressOption) *gin.Engine {
	r := gin.Default()
	r.Use(compress.Compress(opts...))

	r.GET("/small", func(c *gin.Context) {
		c.String(200, smallBody)
	})
	r.GET("/large", func(c *gin.Context) {
		c.String(200, largeBody)
	})
	r.POST("/echo", func(c *gin.Context) {
		c.Header("X-Request-Content-Encoding", c.GetHeader("Content-Encoding"))

		b := bytes.NewBuffer(nil)
		if _, err := io.Copy(b, c.Request.Body); err != nil {
			panic(err)
		}

		c.Data(200, "text/plain", b.Bytes())
	})

	return r
}

func checkNoop(t *testing.T, w *httptest.ResponseRecorder) {
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "", w.Header().Get("Content-Encoding"))
	assert.Equal(t, "", w.Header().Get("Vary"))
}

func checkCompress(t *testing.T, w *httptest.ResponseRecorder, expectedAlgo string) {
	assert.Equal(t, w.Code, 200)
	assert.Equal(t, expectedAlgo, w.Header().Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", w.Header().Get("Vary"))
}

func TestCompressNoopSmall(t *testing.T) {
	req, _ := http.NewRequest("GET", "/small", nil)
	req.Header.Add("Accept-Encoding", "gzip, zstd, br")
	r := setupRouter()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	checkNoop(t, w)

	assert.Equal(t, w.Body.String(), smallBody)
}

func TestCompressNoopNoneAcceptable(t *testing.T) {
	req, _ := http.NewRequest("GET", "/large", nil)
	r := setupRouter()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	checkNoop(t, w)

	assert.Equal(t, w.Body.String(), largeBody)
}

func TestCompressNoopNoneAcceptable2(t *testing.T) {
	req, _ := http.NewRequest("GET", "/large", nil)
	req.Header.Set("Accept-Encoding", "doesnotexist")

	r := setupRouter()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	checkNoop(t, w)

	assert.Equal(t, w.Body.String(), largeBody)
}

func TestCompressGzip(t *testing.T) {
	req, _ := http.NewRequest("GET", "/large", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	r := setupRouter(compress.WithCompressLevel("gzip", compress.GzFlateBestCompression))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	checkCompress(t, w, "gzip")

	gz, err := gzip.NewReader(w.Body)
	assert.NoError(t, err)
	defer gz.Close()

	b := bytes.NewBuffer(nil)
	_, err = gz.WriteTo(b)
	assert.NoError(t, err)
	assert.Equal(t, b.String(), largeBody)
}

func TestCompressBrotli(t *testing.T) {
	req, _ := http.NewRequest("GET", "/large", nil)
	req.Header.Add("Accept-Encoding", "br")
	r := setupRouter()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	checkCompress(t, w, "br")

	br := brotli.NewReader(w.Body)

	b := bytes.NewBuffer(nil)
	if _, err := io.Copy(b, br); err != nil {
		t.Errorf("Decompression failed: %v\n", err)
	}

	assert.Equal(t, b.String(), largeBody)
}

func TestCompressZstd(t *testing.T) {
	req, _ := http.NewRequest("GET", "/large", nil)
	req.Header.Add("Accept-Encoding", "zstd")
	r := setupRouter()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	checkCompress(t, w, "zstd")

	z, err := zstd.NewReader(w.Body)
	assert.NoError(t, err)
	defer z.Close()

	b := bytes.NewBuffer(nil)
	if _, err := io.Copy(b, z); err != nil {
		t.Errorf("Decompression failed: %v\n", err)
	}

	assert.Equal(t, b.String(), largeBody)
}

func TestCompressDeflate(t *testing.T) {
	req, _ := http.NewRequest("GET", "/large", nil)
	req.Header.Add("Accept-Encoding", "deflate")
	r := setupRouter()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	checkCompress(t, w, "deflate")

	z, err := zlib.NewReader(w.Body)
	assert.NoError(t, err)
	defer z.Close()

	b := bytes.NewBuffer(nil)
	if _, err := io.Copy(b, z); err != nil {
		t.Errorf("Decompression failed: %v\n", err)
	}

	assert.Equal(t, b.String(), largeBody)
}

func TestQ(t *testing.T) {
	req, _ := http.NewRequest("GET", "/large", nil)
	req.Header.Add("Accept-Encoding", "br;q=0.5, gzip;q=0.7, deflate;q=0.3")
	r := setupRouter()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	checkCompress(t, w, "gzip")

	gz, err := gzip.NewReader(w.Body)
	assert.NoError(t, err)
	defer gz.Close()

	b := bytes.NewBuffer(nil)
	_, err = gz.WriteTo(b)
	assert.NoError(t, err)
	assert.Equal(t, b.String(), largeBody)
}
