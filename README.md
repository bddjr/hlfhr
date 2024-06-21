# HTTPS Listener For HTTP Redirect

这个 mod 通过劫持 `net.Conn` 实现了一个功能:   
当用户使用 http 协议访问 https 端口时，服务器返回302重定向。  
改编自 `net/http` 。  

This mod implements a feature by hijacking `net.Conn` :  
If a user accesses an https port using http, the server returns 302 redirection.  
Adapted from `net/http`  and `crypto/tls` .  

Related Issue:  
[net/http: configurable error message for Client sent an HTTP request to an HTTPS server. #49310](https://github.com/golang/go/issues/49310)  


***
## Get
```
go get github.com/bddjr/hlfhr
```


***
## Example
[See example/main.go](example/main.go)  

Example:  
```go
var srv *hlfhr.Server

func main() {
	// Use hlfhr.New
	srv = hlfhr.New(&http.Server{
		Addr: ":5678",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Write something...
		}),
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

```

  test:
  cmd /C curl -v -k -L http://localhost:5678/

2024/06/20 11:50:09 http: TLS handshake error from [::1]:60470: hlfhr: Client sent an HTTP request to an HTTPS server
```

[See request](README_curl.md)  


***
## Logic

[See request](README_curl.md)  

HTTPS Server Start -> Hijacking net.Listener.Accept  

### Client HTTPS 
Accept hijacking net.Conn.Read -> First byte 0x16 looks like TLS handshake -> ✅Continue...  

### Client HTTP/1.1
Accept hijacking net.Conn.Read -> First byte looks like HTTP -> HttpOnHttpsPortErrorHandler

If handler nil -> Read Host header and path -> 🔄302 Redirect.  

### Client HTTP/???
Accept hijacking net.Conn.Read -> First byte looks like HTTP -> HttpOnHttpsPortErrorHandler

If handler nil -> Missing Host header -> ❌400 ScriptRedirect.  

### Client ???
Accept hijacking net.Conn.Read -> First byte does not looks like TLS handshake or HTTP request -> Close.

***
## Option Example

#### Hlfhr_ReadFirstRequestBytesLen
```go
srv.Hlfhr_ReadFirstRequestBytesLen = 4096
```

#### Hlfhr_HttpOnHttpsPortErrorHandler
```go
// Default
srv.Hlfhr_HttpOnHttpsPortErrorHandler = func(rb []byte, conn net.Conn) {
	resp := hlfhr.NewResponse(conn)
	// 302 Found
	if host, path, ok := hlfhr.ReadReqHostPath(rb); ok {
		resp.Redirect(302, fmt.Sprint("https://", host, path))
		return
	}
	// script
	resp.ScriptRedirect()
}
```
```go
// Check Host header
srv.Hlfhr_HttpOnHttpsPortErrorHandler = func(rb []byte, conn net.Conn) {
	resp := hlfhr.NewResponse(conn)
	if host, path, ok := hlfhr.ReadReqHostPath(rb); ok {
		// Check Host header
		hostname, port := hlfhr.ReadHostnamePort(host)
		switch hostname {
		case "localhost":
			resp.Redirect(302, fmt.Sprint("https://", host, path))
		case "www.localhost", "127.0.0.1":
			resp.Redirect(302, fmt.Sprint("https://localhost:", port, path))
		default:
			resp.StatusCode = 421
			resp.Write()
		}
		return
	}
	resp.StatusCode = 400
	resp.Write()
}
```
```go
// Script only
srv.Hlfhr_HttpOnHttpsPortErrorHandler = func(rb []byte, conn net.Conn) {
	hlfhr.NewResponse(conn).ScriptRedirect()
}
```
```go
// Custom script only
srv.Hlfhr_HttpOnHttpsPortErrorHandler = func(rb []byte, conn net.Conn) {
	resp := hlfhr.NewResponse(conn)
	resp.StatusCode = 400
	resp.SetContentType("text/html")
	resp.Write(
		"<script> location.protocol = 'https:' </script>\n",
	)
}
```
```go
// Custom script only, not use Response
srv.Hlfhr_HttpOnHttpsPortErrorHandler = func(rb []byte, conn net.Conn) {
	io.WriteString(conn, "HTTP/1.1 400 Bad Request\r\nConnection: close\r\nContent-Type: text/html\r\n\r\n<script> location.protocol = 'https:' </script>\n")
}
```
```go
// Close conn
srv.Hlfhr_HttpOnHttpsPortErrorHandler = func(rb []byte, conn net.Conn) {}
```


***
## Feature Example

#### New  
```go
srv := hlfhr.New(&http.Server{})
```

#### NewServer  
```go
srv := hlfhr.NewServer(&http.Server{})
```

#### Server.Hlfhr_IsShuttingDown
```go
var srv *hlfhr.Server
isShuttingDown := srv.Hlfhr_IsShuttingDown()
```

#### ReadReqHostPath
```go
var rb []byte
host, path, ok := hlfhr.ReadReqHostPath(rb)
```

#### ReadReqMethodHostPath
```go
var rb []byte
method, host, path, ok := hlfhr.ReadReqMethodHostPath(rb)
```

#### ReadReq
```go
var rb []byte
req, err := hlfhr.ReadReq(rb)
```

#### ReadHostnamePort
```go
hostname, port := hlfhr.ReadHostnamePort("localhost:5678")
// hostname: "localhost"
// port: "5678"
```

#### NewResponse
```go
var conn net.Conn
resp := hlfhr.NewResponse(conn)
```

#### Response.SetContentType
```go
var resp *hlfhr.Response
resp.SetContentType("text/html")
```

#### Response.Write
```go
var resp *hlfhr.Response
resp.Write(
	"Hello world!\n",
	"Hello hlfhr!\n",
)
```

#### Response.Redirect
```go
var resp *hlfhr.Response
resp.Redirect(302, "https://example.com")
```

#### Response.ScriptRedirect
```go
var resp *hlfhr.Response
resp.ScriptRedirect()
```

#### NewListener
```go
var l net.Listener
var h HttpOnHttpsPortErrorHandler
l = hlfhr.NewListener(l, 4096, h)
```

#### IsMyListener
```go
var l net.Listener
isHlfhrListener := hlfhr.IsMyListener(l)
```

#### NewConn
```go
var c net.Conn
var h HttpOnHttpsPortErrorHandler
c = hlfhr.NewConn(c, 4096, h)
```

#### IsMyConn
```go
var c net.Conn
isHlfhrConn := hlfhr.IsMyConn(c)
```


***
## License
[BSD-3-clause license](LICENSE.txt)  
