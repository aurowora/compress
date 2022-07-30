package compress

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type compressMiddleware struct {
	cfg *compressOptions
}

func newCompressMiddleware(opts *compressOptions) (cm *compressMiddleware) {
	cm = &compressMiddleware{
		cfg: opts,
	}

	return
}

func (cm *compressMiddleware) Handler(c *gin.Context) {
	if cf, err := cm.decompressRequest(c); err != nil {
		_ = c.AbortWithError(400, err)
		return
	} else if cf != nil {
		defer cf()
	}

	algo := cm.selectAlgorithm(c)

	if algo == "" || !cm.shouldCompress(c) {
		c.Next()
		c.Header("Content-Length", fmt.Sprintf("%v", c.Writer.Size()))
		return
	}

	rw := newResponseWriter(c, cm.cfg.minCompressBytes, algorithms[algo])
	c.Writer = rw
	c.Next()

	_ = rw.Close()

	if rw.Swapped() {
		c.Header("Vary", "Accept-Encoding")
		c.Header("Content-Encoding", algo)
	}
	c.Header("Content-Length", fmt.Sprintf("%v", c.Writer.Size()))
}

// decompresses the request body, if one exists and Content-Encoding is specified
func (cm *compressMiddleware) decompressRequest(c *gin.Context) (func() error, error) {
	if cm.cfg.skipDecompressRequest {
		return nil, nil
	}

	encodings := strings.Split(strings.ReplaceAll(c.GetHeader("Content-Encoding"), " ", ""), ",")
	if len(encodings) == 0 || c.Request.Body == nil {
		// nothing to do
		return nil, nil
	}

	// Content-Encodings are specified in the order they were applied,
	// so we need to unapply them in the reverse order
	readers := make([]io.ReadCloser, 0, len(encodings))
	i := len(encodings) - 1
	for ; i >= (len(encodings)-cm.cfg.maxDecodeSteps) && i >= 0; i-- {
		enc := encodings[i]

		w := c.Request.Body
		if len(readers) > 0 {
			w = readers[len(readers)-1]
		}

		if algo, ok := algorithms[enc]; ok {
			r := algo.getReader(w)
			readers = append(readers, r)
		} else {
			break
		}
	}
	if len(readers) == 0 {
		return nil, nil
	}

	c.Request.Header.Del("Content-Length")
	if i <= -1 {
		c.Request.Header.Del("Content-Encoding")
	} else {
		c.Request.Header.Set("Content-Encoding", strings.Join(encodings[:i+1], ", "))
	}

	br := &compressedBodyReader{
		decomps: readers,
	}
	c.Request.Body = br

	return br.Close, nil
}

type acceptableEncoding struct {
	encoding string
	q        int
}

func (cm *compressMiddleware) selectAlgorithm(c *gin.Context) string {
	acceptEncodings := strings.ToLower(strings.ReplaceAll(c.GetHeader("Accept-Encoding"), " ", ""))
	if acceptEncodings == "" {
		return ""
	}

	allowedEncodings := getEnabledAlgorithms()

	// parse the Accept-Encoding header
	encodings := strings.Split(acceptEncodings, ",")
	acceptableEncodings := make([]acceptableEncoding, 0, len(encodings))
	for _, encoding := range encodings {
		parts := strings.Split(encoding, ";")
		acc := acceptableEncoding{
			encoding: parts[0],
			q:        1000,
		}

		if len(parts) > 1 && strings.HasPrefix(parts[1], "q=") {
			q, err := strconv.ParseFloat(parts[1][2:], 64)
			if err != nil {
				_ = c.Error(err)
			} else {
				acc.q = int(q * 1000)
			}
		}

		// exclude any encodings that are not supported
		if _, ok := allowedEncodings[acc.encoding]; ok && acc.q > 0 {
			acceptableEncodings = append(acceptableEncodings, acc)
		}
	}
	if len(acceptableEncodings) == 0 {
		// could not agree upon an algo
		return ""
	}

	// sort the encodings by q-value first, then their priorities
	sort.Slice(acceptableEncodings, func(i int, j int) bool {
		a, b := acceptableEncodings[i], acceptableEncodings[j]

		if a.q == b.q {
			alA, alB := allowedEncodings[a.encoding], allowedEncodings[b.encoding]
			return alA.getConfig().priority < alB.getConfig().priority

		} else {
			return a.q < b.q
		}
	})

	return acceptableEncodings[len(acceptableEncodings)-1].encoding
}

func (cm *compressMiddleware) shouldCompress(c *gin.Context) bool {
	if strings.Contains(c.GetHeader("Accept"), "text/event-stream") ||
		strings.Contains(c.GetHeader("Connection"), "Upgrade") {

		return false
	}

	if cm.cfg.excludeFunc != nil && cm.cfg.excludeFunc(c) {
		return false
	}

	return len(getEnabledAlgorithms()) > 0
}
