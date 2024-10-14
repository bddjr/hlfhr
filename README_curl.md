[<= Back](README.md)

```curl
curl -v -k -L http://localhost:5678/
* Host localhost:5678 was resolved.
* IPv6: ::1
* IPv4: 127.0.0.1
*   Trying [::1]:5678...
* Connected to localhost (::1) port 5678
> GET / HTTP/1.1
> Host: localhost:5678
> User-Agent: curl/8.7.1
> Accept: */*
>
< HTTP/1.1 302 Found
< Connection: close
< Date: Mon, 14 Oct 2024 10:34:55 GMT
< Location: https://localhost:5678/
< Content-Length: 0
<
* Request completely sent off
* Closing connection
* Clear auth, redirects scheme from HTTP to https
* Issue another request to this URL: 'https://localhost:5678/'
* Hostname localhost was found in DNS cache
*   Trying [::1]:5678...
* Connected to localhost (::1) port 5678
* schannel: disabled automatic use of client certificate
* ALPN: curl offers http/1.1
* ALPN: server accepted http/1.1
* using HTTP/1.x
> GET / HTTP/1.1
> Host: localhost:5678
> User-Agent: curl/8.7.1
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: text/plain
< Date: Mon, 14 Oct 2024 10:34:55 GMT
< Content-Length: 14
<
Hello hlfhr!

* Request completely sent off
* Connection #1 to host localhost left intact
```

[<= Back](README.md)
