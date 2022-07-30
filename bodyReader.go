package compress

import "io"

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

// compressedBodyReader wraps the decompressors so that they all are properly closed
type compressedBodyReader struct {
	decomps []io.ReadCloser // must be ordered such that the last item in the slice is the last reader
}

func (c *compressedBodyReader) Read(b []byte) (int, error) {
	return c.decomps[len(c.decomps)-1].Read(b)
}

func (c *compressedBodyReader) Close() error {
	for _, dc := range c.decomps {
		if err := dc.Close(); err != nil {
			return err
		}
	}
	return nil
}
