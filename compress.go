package compress

/*
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/

import "github.com/gin-gonic/gin"

// Compress creates the Compress middleware with the default options
func Compress(opts ...CompressOption) gin.HandlerFunc {
	co := newCompressOptions()

	for _, opt := range opts {
		opt(co)
	}

	return newCompressMiddleware(co).Handler
}
