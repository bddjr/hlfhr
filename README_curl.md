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
< Content-Length: 46
< Content-Type: text/html; charset=utf-8
< Date: Tue, 25 Jun 2024 10:10:11 GMT
< Location: https://localhost:5678/
< X-Powered-By: github.com/bddjr/hlfhr
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
* Request completely sent off
< HTTP/1.1 200 OK
< Accept-Ranges: bytes
< Content-Length: 320
< Content-Type: text/html; charset=utf-8
< Last-Modified: Wed, 19 Jun 2024 10:52:39 GMT
< Date: Tue, 25 Jun 2024 10:10:11 GMT
<
<html>

<head>
    <meta name="robots" content="noindex, nofollow">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * {
            color-scheme: light dark;
        }
    </style>
</head>

<body>
    <h1>Hello HTTPS!</h1>
    <p>hlfhr</p>
</body>

</html>* Connection #1 to host localhost left intact
```

[<= Back](README.md)
