package compress

import (
	"io"
	"io/ioutil"
	"sync"

	"github.com/klauspost/compress/zstd"
)

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

const (
	ZstdSpeedFastest           = int(zstd.SpeedFastest)
	ZstdSpeedDefault           = int(zstd.SpeedDefault)
	ZstdSpeedBetterCompression = int(zstd.SpeedBetterCompression)
	ZstdSpeedBestCompression   = int(zstd.SpeedBestCompression)
)

type algorithmZstd struct {
	compressorPool   *sync.Pool
	decompressorPool *sync.Pool
	cfg              algorithmConfig
}

// makeCompressor allocates a new ZSTD encoder (not for direct use)
func (a *algorithmZstd) makeCompressor() interface{} {
	z, err := zstd.NewWriter(ioutil.Discard, zstd.WithEncoderLevel(zstd.EncoderLevel(a.cfg.compressLevel)))
	if err != nil {
		panic(err)
	}

	return z
}

/* Implement algorithm */

func (a *algorithmZstd) getConfig() *algorithmConfig {
	return &a.cfg
}

func (a *algorithmZstd) getWriter(w io.Writer) io.WriteCloser {
	zw := a.compressorPool.Get().(*zstd.Encoder)
	zw.Reset(w)

	return &wrappedWriter{
		p: a.compressorPool,
		w: zw,
	}
}

func (a *algorithmZstd) getReader(r io.Reader) io.ReadCloser {
	zr := a.decompressorPool.Get().(*zstd.Decoder)
	if err := zr.Reset(r); err != nil {
		panic(err)
	}

	return &wrappedReader{
		p: a.decompressorPool,
		r: zr,
	}
}

func newAlgorithmZstd() *algorithmZstd {
	a := algorithmZstd{
		cfg: algorithmConfig{
			priority:      100,
			enable:        true,
			compressLevel: ZstdSpeedDefault,
		},
		decompressorPool: &sync.Pool{
			New: func() interface{} {
				z, err := zstd.NewReader(nil)
				if err != nil {
					panic(err)
				}

				return z
			},
		},
	}

	a.compressorPool = &sync.Pool{
		New: a.makeCompressor,
	}

	return &a
}
