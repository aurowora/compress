# gin-compress

Middleware for [Gin Gonic](https://github.com/gin-gonic/gin) for compressing HTTP responses
and decompressing HTTP requests.

Currently, this package supports Brotli, GZIP, Deflate, and ZSTD for both compressing and decompressing.

This was created for my own purposes. If you're looking for something that is perhaps a bit more actively supported,
check out the official [Gin GZIP Middleware](https://github.com/gin-contrib/gzip).

### Usage

If nothing special is desired, one can add the Compress middleware with the default 
config to your Gin router like so:

```go
package main

import "log"
import "github.com/gin-gonic/gin" 
import "github.com/gin-contrib/size"
import "github.com/aurowora/compress"

func main() {
	r := gin.Default()
	r.Use(compress.Compress())
	// Limit payload to 10 MB, notice how it follows Compress() MW.
	// See the "Security" section below...
	r.Use(size.RequestSizeLimiter(10 * 1024 * 1024)) 
	
	// Declare routes and do whatever else here...
	
	log.Fatalln(r.Run())
}
```

Despite the name, the Compress middleware handles compressing both the response body and decompressing the request body.

To configure the middleware, pass the return value of the functions beginning with `With` 
to the middleware's constructors, like so:

```go
// disables Brotli
r.Use(compress.Compress(compress.WithAlgo(compress.BROTLI, false)))
```

#### Configuration

The following configuration options are available for the Compress middleware:

| Function Signature                                                           | Default                              | Description                                                                                                                                                           |
|------------------------------------------------------------------------------|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| WithAlgo(algo CompressAlgorithm, enable bool) | All enabled                          | Allows enabling/disabling any of the supported algorithms. Valid algorithms are currently `compress.ZSTD`, `compress.BROTLI`, `compress.GZIP`, and `compress.DEFLATE` |
| WithCompressLevel(algo CompressAlgorithm, level int)                         | Default for all algorithms           | Allows setting the compression level for any supported algorithm. See the Brotli*, GzFlate*, and Zstd* constants.                                                     |
| WithPriority(algo CompressAlgorithm)                                         | Order is Brotli, GZIP, Deflate, ZSTD | Specify the priority of an algorithm when the client will accept multiple. Higher priorities win.                                                                     |
| WithExcludeFunc(f func(c *gin.Context) bool)                                 | Not Set                              | Specify a function to be called to determine if the compressor should run. Note that response headers/body is not available at this point.                            |
| WithMinCompressBytes(numBytes int)                                           | 512                                  | Do not invoke the compressor unless the response body is at least this many bytes                                                                                     |
| WithMaxDecodeSteps(steps int)                                                | 1                                    | Determines how many rounds of decompression to perform if Content-Encoding includes multiple decompression algorithms.                                                |
| WithDecompressBody(decompress bool)                                          | true                                 | Specifies whether the request body should be decompressed at all.                                                                                                     |

### Security

Bugs/design flaws in the underlying compression algorithm implementations could allow for "zip bombs" that, when
decompressed, expand to massive payloads. This results in high resource usage
and potentially even denial of service. To mitigate this, applications making use of the request body decompression feature
(i.e. WithDecompressBody(true), which is the default), should also use [gin-contrib/size](https://github.com/gin-contrib/size)
(or an equivalent) to limit the size of request bodies.

It is important that any payload size limiter middleware used **come after the compress middleware** as it will be useless otherwise.

An example of correct usage is shown above.

### Tests

Tests cover most functionality in this package. The built-in tests can be run using `go test`.

### License Notice

```
gin-compress Copyright (C) 2022 Aurora McGinnis

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
```

The full terms of the Mozilla Public License 2.0 can also be 
found in the LICENSE.txt file within this repository.

The author of this package is not associated with the authors of [Gin](https://github.com/gin-gonic/gin) in any way.