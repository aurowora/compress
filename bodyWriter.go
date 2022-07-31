package compress

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
)

// respWriter wraps the default request writer to allow for compressing the request contents. It uses an internal buffer
// until threshold is hit, at which point it switches to the compressor.
// If threshold is never hit, calling Close() will copy the buffer contents to the response writer
type respWriter struct {
	gin.ResponseWriter
	threshold    int
	algo         algorithm
	buf          *bytes.Buffer
	bytesWritten int
	compressor   io.WriteCloser
}

func newResponseWriter(c *gin.Context, swapSize int, algo algorithm) *respWriter {
	return &respWriter{
		c.Writer,
		swapSize,
		algo,
		bytes.NewBuffer(nil),
		0,
		nil,
	}
}

func (rw *respWriter) WriteString(s string) (int, error) {
	return rw.Write([]byte(s))
}

func (rw *respWriter) Write(b []byte) (int, error) {
	rw.Header().Del("Content-Length")

	if !rw.Swapped() && rw.buf.Len()+len(b) >= rw.threshold {
		rw.compressor = rw.algo.getWriter(rw.ResponseWriter)
		if copied, err := io.Copy(rw.compressor, rw.buf); err != nil {
			return int(copied), err
		}
		rw.buf = nil
	}

	var w io.Writer
	if rw.Swapped() {
		w = rw.compressor
	} else {
		w = rw.buf
	}

	if n, err := w.Write(b); err != nil {
		return n, err
	} else {
		rw.bytesWritten += n
		return n, err
	}
}

func (rw *respWriter) Size() int {
	return rw.bytesWritten
}

func (rw *respWriter) Written() bool {
	return rw.bytesWritten > 0
}

func (rw *respWriter) Close() error {
	if !rw.Swapped() {
		// buf was never switched...
		if _, err := io.Copy(rw.ResponseWriter, rw.buf); err != nil {
			return err
		}
	} else {
		return rw.compressor.Close()
	}

	return nil
}

func (rw *respWriter) Swapped() bool {
	return rw.buf == nil
}

func (rw *respWriter) WriteHeader(code int) {
	rw.Header().Del("Content-Length")
	rw.ResponseWriter.WriteHeader(code)
}
