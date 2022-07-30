package compress

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

import (
	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/flate"
)

const (
	DEFLATE = "deflate"
	GZIP    = "gzip"
	ZSTD    = "zstd"
	BROTLI  = "br"
)

// ExcludeFunc should return true if compression should be skipped for the current request
type ExcludeFunc func(c *gin.Context) bool

// compressOptions is used to configure the Compress middleware. Use NewCompressOptionsBuilder() to create this.
type compressOptions struct {
	excludeFunc ExcludeFunc

	// minCompressBytes specifies the minimum size a response must be to justify compressing.
	minCompressBytes int
	// maxDecodeSteps specifies how many layers of compression we will attempt to undo for Content-Encoding headers
	// that specify multiple
	maxDecodeSteps int
	// skipDecompressRequest can be used to skip decompression of the body
	skipDecompressRequest bool
}

type CompressOption func(opts *compressOptions)

// newCompressOptions creates a new compressOptions with defaults applied
func newCompressOptions() *compressOptions {
	return &compressOptions{
		excludeFunc: func(c *gin.Context) bool {
			return false
		},
		minCompressBytes:      512,
		maxDecodeSteps:        1,
		skipDecompressRequest: false,
	}
}

// WithAlgo specifies whether algo should be enabled for both compression and decompression
func WithAlgo(algo string, enable bool) CompressOption {
	return func(opts *compressOptions) {
		algorithms[algo].getConfig().enable = enable
	}
}

// GzFlate* constants are suitable for both Deflate and GZIP
const (
	GzFlateDefault             = flate.DefaultCompression
	GzFlateNoCompression       = flate.NoCompression
	GzFlateBestSpeed           = flate.BestSpeed
	GzFlateBestCompression     = flate.BestCompression
	GzFlateConstantCompression = flate.ConstantCompression
	GzFlateHuffmanOnly         = flate.HuffmanOnly
)

// WithCompressLevel specifies what level to use for compression. Take care that the specified level
// is valid for the algorithm you've selected.
func WithCompressLevel(algo string, level int) CompressOption {
	return func(opts *compressOptions) {
		algorithms[algo].getConfig().compressLevel = level
	}
}

// WithPriority specifies which compression algo to use when the client can accept multiple algorithms.
// The highest priority algorithm that the client will accept wins.
func WithPriority(algo string, priority int) CompressOption {
	return func(opts *compressOptions) {
		algorithms[algo].getConfig().priority = priority
	}
}

// WithExcludeFunc specifies a function that is called before compression to determine if the response to the current request shouldn't
// be compressed.
func WithExcludeFunc(f func(c *gin.Context) bool) CompressOption {
	return func(opts *compressOptions) {
		opts.excludeFunc = f
	}
}

// WithMinCompressBytes specifies the minimum size a response must be before compressing.
// Using a value <= 0 will always compress.
func WithMinCompressBytes(numBytes int) CompressOption {
	return func(opts *compressOptions) {
		opts.minCompressBytes = numBytes
	}
}

// WithMaxDecodeSteps specifies how many layers of request body compression to undo if multiple.
func WithMaxDecodeSteps(steps int) CompressOption {
	if steps < 1 {
		panic("steps < 1, if you want to disable decompression of the request body, see WithDecompressBody")
	}

	return func(opts *compressOptions) {
		opts.maxDecodeSteps = steps
	}
}

// WithDecompressBody specifies whether to decompress the request body
func WithDecompressBody(decompress bool) CompressOption {
	return func(opts *compressOptions) {
		opts.skipDecompressRequest = !decompress
	}
}
