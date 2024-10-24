# Simple HTTPS Listener For HTTP Redirect

Full version: [hlfhr](../)

---

## Get

```
go get github.com/bddjr/hlfhr
```

---

## Logic

Hijacking net.Listener.Accept -> Hijacking net.Conn.Read

### Client HTTPS

First byte not looks like HTTP -> ✅Continue...

### Client HTTP/1.1

First byte looks like HTTP -> Read bytes until encountering "\n\r\n" or "\n\n" -> 🟡300 Script Redirect -> Close.

### See

- [simple.go](simple.go)

---

## Example

```go
// Use ListenAndServeTLS
func main() {
	srv := &http.Server{
		Addr:              ":5677",
		Handler:           http.HandlerFunc(httpResponseHandle),
		ReadHeaderTimeout: 10 * time.Second,
	}

	err := simplehlfhr.ListenAndServeTLS(srv, "../../test/localhost.crt", "../../test/localhost.key")
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
```

```go
// Use NewListener
func main() {
	srv := &http.Server{
		Addr:              ":5677",
		Handler:           http.HandlerFunc(httpResponseHandle),
		ReadHeaderTimeout: 10 * time.Second,
	}

	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	// Use NewListener
	l = simplehlfhr.NewListener(l, srv)

	// Must use ServeTLS! For issue https://github.com/bddjr/hlfhr/issues/4
	err = srv.ServeTLS(l, "localhost.crt", "localhost.key")
	fmt.Println(err)
}
```

---

## Test

```
git clone https://github.com/bddjr/hlfhr
cd hlfhr
cd simplehlfhr
./run.sh
```

---

## Reference

https://developer.mozilla.org/docs/Web/HTTP/Status/300

[hlfhr](../)

---

## License

[BSD-3-clause license](../LICENSE.txt)
