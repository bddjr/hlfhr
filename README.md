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
	if err != nil && err != http.ErrServerClosed {
		fmt.Println(err)
	}
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

Request:  
```curl
curl -v http://localhost:5678/foo/bar
*   Trying [::1]:5678...
* Connected to localhost (::1) port 5678
> GET /foo/bar HTTP/1.1
> Host: localhost:5678
> User-Agent: curl/8.4.0
> Accept: */*
>
< HTTP/1.1 302 Found
< Location: https://localhost:5678/foo/bar
< Connection: close
<
Redirect to HTTPS
* Closing connection
```

<br/>

***
## Option Example

Hlfhr_ReadFirstRequestBytesLen
```go
srv.Hlfhr_ReadFirstRequestBytesLen = 4096
```

Hlfhr_HttpOnHttpsPortErrorHandler
```go
srv.Hlfhr_HttpOnHttpsPortErrorHandler = func(rb []byte, conn net.Conn) {
	// 302 Found
	if host, path, ok := hlfhr.ReadReqHostPath(rb); ok {
		fmt.Fprint(conn,
			"HTTP/1.1 302 Found\r\n",
			"Location: https://", host, path, "\r\n",
			"Connection: close\r\n",
			"\r\n",
			"Redirect to HTTPS\n",
		)
		return
	}
	// script
	fmt.Fprint(conn,
		"HTTP/1.1 400 Bad Request\r\n",
		"Content-Type: text/html\r\n",
		"Connection: close\r\n",
		"\r\n",
		"<script> location.protocol = 'https:' </script>\n",
	)
}
```

<br/>

***
## License
[BSD-3-clause license](LICENSE.txt)  
