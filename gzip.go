package compress

import (
	"io"
	"io/ioutil"
	"sync"

	"github.com/klauspost/compress/gzip"
)

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

type algorithmGzip struct {
	compressorPool *sync.Pool
	cfg            algorithmConfig
}

func (a *algorithmGzip) makeCompressor() interface{} {
	gz, err := gzip.NewWriterLevel(ioutil.Discard, a.cfg.compressLevel)
	if err != nil {
		panic(err)
	}

	return gz
}

/* Implement algorithm */

func (a *algorithmGzip) getConfig() *algorithmConfig {
	return &a.cfg
}

func (a *algorithmGzip) getWriter(w io.Writer) io.WriteCloser {
	gw := a.compressorPool.Get().(*gzip.Writer)
	gw.Reset(w)

	return &wrappedWriter{
		p: a.compressorPool,
		w: gw,
	}
}

func (a *algorithmGzip) getReader(r io.Reader) io.ReadCloser {
	// gzip reader implements Reset, but isn't usable with a sync pool
	// because it'll panic when you try to construct one with a nil
	// reader
	gr, err := gzip.NewReader(r)
	if err != nil {
		panic(err)
	}

	return gr
}

func newAlgorithmGzip() *algorithmGzip {
	a := algorithmGzip{
		cfg: algorithmConfig{
			priority:      300,
			enable:        true,
			compressLevel: GzFlateDefault,
		},
	}

	a.compressorPool = &sync.Pool{
		New: a.makeCompressor,
	}

	return &a
}
