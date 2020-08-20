# fasttrie
Go Trie from [fasthttp/router](https://github.com/fasthttp/router).

## What Change

* replace handler(fasthttp.RequestHandler) to value(interface{})
* replace ctx(fasthttp.RequestCtx) to params(map\[string\]string)
