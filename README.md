# HTTPS Listener For HTTP Redirect

这个 mod 通过劫持 `net.Conn` 实现了一个功能:   
当用户使用 http 协议访问 https 端口时，服务器返回302重定向。  
改编自 `net/http` 。  

This mod implements a feature by hijacking `net.Conn` :  
If a user accesses an https port using http, the server returns 302 redirection.  
Adapted from `net/http` .  

Related Issue:  
[net/http: configurable error message for Client sent an HTTP request to an HTTPS server. #49310](https://github.com/golang/go/issues/49310)  


***
## Get
```
go get github.com/bddjr/hlfhr
```


***
## Logic

HTTPS Server Start -> Hijacking net.Listener.Accept  

Client HTTPS -> Accept hijacking net.Conn.Read -> Not looks like HTTP -> ✅Continue...  

Client HTTP/1.1 -> Accept hijacking net.Conn.Read -> Looks like HTTP -> 🔄302 Redirect.  

Client HTTP/??? -> Accept hijacking net.Conn.Read -> Looks like HTTP -> Missing Host header -> ❌400 Script.  

[See request](README_curl.md)  

***
## Example
[See example/main.go](example/main.go)  

Example:  
```go
var srv *hlfhr.Server

func main() {
	srv = hlfhr.New(&http.Server{
		Addr:              ":5678",
		Handler:           http.HandlerFunc(httpResponseHandle),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       10 * time.Second,
	})
	// Then just use it like http.Server .

	err := srv.ListenAndServeTLS("localhost.crt", "localhost.key")
	fmt.Println(err)
}
```

Run:  
```
git clone https://github.com/bddjr/hlfhr
cd hlfhr
cd example

go build
./example
```

***
## Option Example

Hlfhr_ReadFirstRequestBytesLen
```go
srv.Hlfhr_ReadFirstRequestBytesLen = 4096
```

Hlfhr_HttpOnHttpsPortErrorHandler
```go
srv.Hlfhr_HttpOnHttpsPortErrorHandler = func(rb []byte, conn net.Conn) {
	resp := hlfhr.NewResponse(conn)
	// 302 Found
	if host, path, ok := hlfhr.ReadReqHostPath(rb); ok {
		resp.Redirect(302, fmt.Sprint("https://", host, path))
		return
	}
	// script
	resp.StatusCode = 400
	resp.SetContentType("text/html")
	resp.Write(
		"<!-- ", hlfhr.ErrHttpOnHttpsPort, " -->\n",
		"<script> location.protocol = 'https:' </script>\n",
	)
}
```


***
## Feature Example

New  
```go
srv := hlfhr.New(&http.Server{})
```

NewServer  
```go
srv := hlfhr.NewServer(&http.Server{})
```

ReadReqHostPath
```go
var rb []byte
host, path, ok := hlfhr.ReadReqHostPath(rb)
```

ReadReq
```go
var rb []byte
req, err := hlfhr.ReadReq(rb)
```

NewResponse
```go
var conn net.Conn
resp := hlfhr.NewResponse(conn)
```

Response.SetContentType
```go
var resp *hlfhr.Response
resp.SetContentType("text/html")
```

Response.Write
```go
var resp *hlfhr.Response
resp.Write(
	"Hello world!\n",
	"Hello hlfhr!\n",
)
```

Response.Redirect
```go
var resp *hlfhr.Response
resp.Redirect(302, "https://example.com")
```

***
## License
[BSD-3-clause license](LICENSE.txt)  
