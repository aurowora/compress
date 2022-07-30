package compress

import (
	"io"
	"io/ioutil"
	"sync"

	"github.com/klauspost/compress/zlib"
)

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

type algorithmDeflate struct {
	compressorPool *sync.Pool
	cfg            algorithmConfig
}

func (a *algorithmDeflate) makeCompressor() interface{} {
	dw, err := zlib.NewWriterLevel(ioutil.Discard, a.cfg.compressLevel)
	if err != nil {
		panic(err)
	}

	return dw
}

/* Implement algorithm */

func (a *algorithmDeflate) getConfig() *algorithmConfig {
	return &a.cfg
}

func (a *algorithmDeflate) getWriter(w io.Writer) io.WriteCloser {
	dw := a.compressorPool.Get().(*zlib.Writer)
	dw.Reset(w)

	return &wrappedWriter{
		p: a.compressorPool,
		w: dw,
	}
}

func (a *algorithmDeflate) getReader(r io.Reader) io.ReadCloser {
	dr, err := zlib.NewReader(r)
	if err != nil {
		panic(err)
	}

	return dr
}

func newAlgorithmDeflate() *algorithmDeflate {
	a := algorithmDeflate{
		cfg: algorithmConfig{
			priority:      200,
			enable:        true,
			compressLevel: GzFlateDefault,
		},
	}

	a.compressorPool = &sync.Pool{
		New: a.makeCompressor,
	}

	return &a
}