package compress

import (
	"io"
	"io/ioutil"
	"sync"

	"github.com/andybalholm/brotli"
)

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

const (
	BrotliBestSpeed          = brotli.BestSpeed
	BrotliBestCompression    = brotli.BestCompression
	BrotliDefaultCompression = brotli.DefaultCompression
)

type algorithmBrotli struct {
	compressorPool   *sync.Pool
	decompressorPool *sync.Pool
	cfg              algorithmConfig
}

func (a *algorithmBrotli) makeCompressor() interface{} {
	return brotli.NewWriterLevel(ioutil.Discard, a.cfg.compressLevel)
}

/* Implement algorithm */

func (a *algorithmBrotli) getConfig() *algorithmConfig {
	return &a.cfg
}

func (a *algorithmBrotli) getWriter(w io.Writer) io.WriteCloser {
	bw := a.compressorPool.Get().(*brotli.Writer)
	bw.Reset(w)

	return &wrappedWriter{
		p: a.compressorPool,
		w: bw,
	}
}

func (a *algorithmBrotli) getReader(r io.Reader) io.ReadCloser {
	br := a.decompressorPool.Get().(*brotli.Reader)
	if err := br.Reset(r); err != nil {
		panic(err)
	}

	return &wrappedReader{
		p: a.decompressorPool,
		r: br,
	}
}

func newAlgorithmBrotli() *algorithmBrotli {
	a := algorithmBrotli{
		cfg: algorithmConfig{
			priority:      400,
			enable:        true,
			compressLevel: BrotliDefaultCompression,
		},
		decompressorPool: &sync.Pool{
			New: func() interface{} {
				return brotli.NewReader(nil)
			},
		},
	}

	a.compressorPool = &sync.Pool{
		New: a.makeCompressor,
	}

	return &a
}
