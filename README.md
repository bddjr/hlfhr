# HTTPS Listener For HTTP Redirect

This mod implements a feature by hijacking `net.Conn` :  
If a user accesses an https port using http, the server returns 302 redirection.  
The principle is similar to reverse proxy.

这个 mod 通过劫持 `net.Conn` 实现了一个功能:  
当用户使用 http 协议访问 https 端口时，服务器返回 302 重定向。  
原理与反向代理相似。

Related Issue:  
[net/http: configurable error message for Client sent an HTTP request to an HTTPS server. #49310](https://github.com/golang/go/issues/49310)

---

## Get

```
go get github.com/bddjr/hlfhr
```

---

## Example

[See example/main.go](example/main.go)

Example:

```go
// Use srv.ListenAndServeTLS
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
// Use srv.ServeTLS
var l net.Listener
var srv *hlfhr.Server

func main() {
	srv = hlfhr.New(&http.Server{
		Addr: ":5678",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Write something...
		}),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       10 * time.Second,
	})

	var err error
	l, err = net.Listen("tcp", srv.Addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	// Must use ServeTLS! For issue https://github.com/bddjr/hlfhr/issues/4
	err = srv.ServeTLS(l, "localhost.crt", "localhost.key")
	fmt.Println(err)
}
```

```go
// Use hlfhr.NewListener
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
	var httpOnHttpsPortErrorHandler http.Handler = nil
	l = hlfhr.NewListener(l, srv, httpOnHttpsPortErrorHandler)

	// Must use ServeTLS! For issue https://github.com/bddjr/hlfhr/issues/4
	err = srv.ServeTLS(l, "localhost.crt", "localhost.key")
	fmt.Println(err)
}
```

Run:

```
git clone https://github.com/bddjr/hlfhr
cd hlfhr
./run.sh
```

```

  test:
  curl -v -k -L http://localhost:5678/

2024/06/20 11:50:09 http: TLS handshake error from [::1]:60470: hlfhr: Client sent an HTTP request to an HTTPS server
```

---

## Logic

HTTPS Server Start -> Hijacking net.Listener.Accept -> Hijacking net.Conn

### Client HTTPS

First byte not looks like HTTP -> ✅Continue...

### Client HTTP/1.1

First byte looks like HTTP -> Read request -> Found Host header -> ⚪HttpOnHttpsPortErrorHandler or 🟡302 Redirect -> Close.

### Client HTTP/???

First byte looks like HTTP -> Read request -> ⛔Missing Host header -> Close.

### See

- [curl](README_curl.md)
- [conn.go](conn.go)
- [conn-http-header-reader.go](conn-http-header-reader.go)

---

## Option Example

#### HttpOnHttpsPortErrorHandler

```go
// 307 Temporary Redirect
srv.HttpOnHttpsPortErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	hlfhr.RedirectToHttps(w, r, 307)
})
```

```go
// Check Host Header
srv.HttpOnHttpsPortErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	hostname, port := hlfhr.SplitHostnamePort(r.Host)
	switch hostname {
	case "localhost":
		//
	case "www.localhost", "127.0.0.1":
		r.Host = hlfhr.HostnameAppendPort("localhost", port)
	default:
		w.WriteHeader(421)
		return
	}
	hlfhr.RedirectToHttps(w, r, 302)
})
```

```go
// Script Redirect
srv.HttpOnHttpsPortErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(400)
	io.WriteString(w, "<script>location.protocol='https:'</script>")
})
```

---

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
```

#### Redirect

```go
var w http.ResponseWriter
hlfhr.Redirect(w, 302, "https://example.com/")
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

---

## License

[BSD-3-clause license](LICENSE.txt)
