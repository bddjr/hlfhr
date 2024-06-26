# HTTPS Listener For HTTP Redirect

这个 mod 通过劫持 `net.Conn` 实现了一个功能:   
当用户使用 http 协议访问 https 端口时，服务器返回302重定向。  

This mod implements a feature by hijacking `net.Conn` :  
If a user accesses an https port using http, the server returns 302 redirection.  

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
```go
var l net.Listener
var srv *http.Server

func main() {
	srv = &http.Server{
		Addr: ":5678",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Write something...
		}),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       10 * time.Second,
	}

	var err error
	l, err = net.Listen("tcp", srv.Addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	// Use hlfhr.NewListener
	l = hlfhr.NewListener(l, srv, nil)

	err = srv.ServeTLS(l, "localhost.crt", "localhost.key")
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


***
## Logic

[See request](README_curl.md)  

HTTPS Server Start -> Hijacking net.Listener.Accept  

### Client HTTPS 
Accept hijacking net.Conn.Read -> First byte not looks like HTTP -> ✅Continue...  

### Client HTTP/1.1
Accept hijacking net.Conn.Read -> First byte looks like HTTP -> Read request -> Found Host header -> HttpOnHttpsPortErrorHandler -> Close connect.

If handler nil -> 🔄302 Redirect -> Close connect.  

### Client HTTP/???
Accept hijacking net.Conn.Read -> First byte looks like HTTP -> Read request -> Missing Host header -> ❌400 Bad Request -> Close connect.  


***
## Option Example

#### HttpOnHttpsPortErrorHandler
```go
// Default
srv.HttpOnHttpsPortErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	hlfhr.RedirectToHttps(w, r, 302)
})
```
```go
// Check Host Header
srv.HttpOnHttpsPortErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	switch hlfhr.Hostname(r.Host) {
	case "localhost":
		hlfhr.RedirectToHttps(w, r, 302)
	case "www.localhost", "127.0.0.1":
		hlfhr.Redirect(w, 302, "https://"+hlfhr.ReplaceHostname(r.Host, "localhost")+r.URL.RequestURI())
	default:
		w.WriteHeader(421)
	}
})
```
```go
// Script Redirect
srv.HttpOnHttpsPortErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(400)
	io.WriteString(w, "<script> location.protocol = 'https:' </script>")
})
```
```go
// Keep Alive
srv.HttpOnHttpsPortErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(400)
	io.WriteString(w, "Hello hlfhr")
})
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

#### Server.IsShuttingDown
```go
var srv *hlfhr.Server
isShuttingDown := srv.IsShuttingDown()
```

#### Server.NewListener
```go
var l net.Listener
var srv *http.Server
l = hlfhr.New(srv).NewListener(l)
```

#### NewListener
```go
var l net.Listener
var srv *http.Server
var h http.Handler
l = hlfhr.NewListener(c, srv, h)
```

#### IsMyListener
```go
var l net.Listener
isHlfhrListener := hlfhr.IsMyListener(l)
```

#### NewConn
```go
var c net.Conn
var srv *http.Server
var h http.Handler
c = hlfhr.NewConn(c, srv, h)
```

#### IsMyConn
```go
var c net.Conn
isHlfhrConn := hlfhr.IsMyConn(c)
```

#### NewResponse
```go
resp := hlfhr.NewResponse()
```

#### NewResponseWriter
```go
var c net.Conn
w := hlfhr.NewResponseWriter(c, nil)
hw := http.ResponseWriter(w)
```

#### ResponseWriter.Finish
```go
var w *hlfhr.ResponseWriter
w.Finish()
```

#### Redirect
```go
var w http.ResponseWriter
hlfhr.Redirect(w, "https://example.com/", 302)
```

#### RedirectToHttps
```go
var w http.ResponseWriter
var r *http.Request
hlfhr.RedirectToHttps(w, r, 302)
```

#### SplitHostnamePort
```go
hostname, port := hlfhr.SplitHostnamePort("[::1]:5678")
// hostname: [::1]
// port: 5678
```

#### Hostname
```go
hostname := hlfhr.Hostname("[::1]:5678")
// hostname: [::1]
```

#### Port
```go
port := hlfhr.Port("[::1]:5678")
// port: 5678
```

#### HostnameAppendPort
```go
Host := hlfhr.HostnameAppendPort("[::1]", "5678")
// Host: [::1]:5678
```

#### ReplaceHostname
```go
Host := hlfhr.ReplaceHostname("[::1]:5678", "localhost")
// Host: localhost:5678
```

#### ReplacePort
```go
Host := hlfhr.ReplacePort("[::1]:5678", "7890")
// Host: [::1]:7890
```

#### Ipv6CutPrefixSuffix
```go
v6 := hlfhr.Ipv6CutPrefixSuffix("[::1]")
// v6: ::1
```


***
## License
[BSD-3-clause license](LICENSE.txt)  
