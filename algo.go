package compress

import (
	"io"
	"io/ioutil"
	"sync"
)

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

// algorithmConfig specifies options for a given compression algorithm
type algorithmConfig struct {
	// enable indicates whether or not this compressor should be used
	enable bool
	// compressLevel is passed to the encoder object
	compressLevel int
	// priority indicates which algorithm will be selected when the client accepts multiple algorithms with equal q values
	priority int
}

type algorithm interface {
	// returns a compressor for this algorithm
	getWriter(w io.Writer) io.WriteCloser
	// returns a decompressor for this algorithm
	getReader(r io.Reader) io.ReadCloser
	// returns a pointer to the configuration struct
	getConfig() *algorithmConfig
}

// algoritms contains the supported algorithms
var algorithms = map[string]algorithm{
	ZSTD:    newAlgorithmZstd(),
	GZIP:    newAlgorithmGzip(),
	BROTLI:  newAlgorithmBrotli(),
	DEFLATE: newAlgorithmDeflate(),
}

func getEnabledAlgorithms() map[string]algorithm {
	algos := make(map[string]algorithm, len(algorithms))

	for k, v := range algorithms {
		if v.getConfig().enable {
			algos[k] = v
		}
	}

	return algos
}

/*
	The resettable* and wrapped* types are used to handle re-using writers/readers that support it
	while still presenting a WriteCloser/ReadCloser interface
*/

type resettableCompressor interface {
	io.WriteCloser
	Reset(w io.Writer)
}

type resettableDecompressor interface {
	io.Reader
	Reset(r io.Reader) error
}

type wrappedWriter struct {
	w resettableCompressor
	p *sync.Pool
	c bool
}

func (w *wrappedWriter) Write(b []byte) (int, error) {
	if w.c {
		panic("attempted to write to a closed writer")
	}

	return w.w.Write(b)
}

func (w *wrappedWriter) Close() error {
	if w.c {
		panic("attempted to close a compressor that has already been closed")
	}

	w.c = true

	defer w.p.Put(w.w)
	defer w.w.Reset(ioutil.Discard)
	return w.w.Close()
}

type wrappedReader struct {
	r resettableDecompressor
	p *sync.Pool
	c bool
}

func (r *wrappedReader) Read(b []byte) (int, error) {
	if r.c {
		panic("attempted to read from a closed reader")
	}

	return r.r.Read(b)
}

func (r *wrappedReader) Close() error {
	if r.c {
		return nil
	}

	r.c = true
	defer r.p.Put(r.r)
	r.r.Reset(nil)
	return nil
}
