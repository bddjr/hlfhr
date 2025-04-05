// HTTPS Listener For HTTP Redirect
//
// https://github.com/bddjr/hlfhr
package hlfhr

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync/atomic"
	"unsafe"
)

type Server struct {
	*http.Server

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	// See https://github.com/bddjr/hlfhr#httponhttpsporterrorhandler-example
	HttpOnHttpsPortErrorHandler http.Handler

	// Port 80 redirects to port 443.
	// This option only takes effect when listening on port 443.
	// HttpOnHttpsPortErrorHandler is also using on port 80.
	Listen80RedirectTo443 bool
}

// New hlfhr Server
func New(s *http.Server) *Server {
	if s == nil {
		s = new(http.Server)
	}
	return &Server{Server: s}
}

// New hlfhr Server
func NewServer(s *http.Server) *Server {
	return New(s)
}

// ServeTLS accepts incoming connections on the Listener l, creating a
// new service goroutine for each. The service goroutines perform TLS
// setup and then read requests, calling srv.Handler to reply to them.
//
// Files containing a certificate and matching private key for the
// server must be provided if neither the [Server]'s
// TLSConfig.Certificates nor TLSConfig.GetCertificate are populated.
// If the certificate is signed by a certificate authority, the
// certFile should be the concatenation of the server's certificate,
// any intermediates, and the CA's certificate.
//
// ServeTLS always returns a non-nil error. After [Server.Shutdown] or [Server.Close], the
// returned error is [http.ErrServerClosed].
func (s *Server) ServeTLS(l net.Listener, certFile string, keyFile string) error {
	// Setup HTTP/2
	if s.TLSConfig == nil {
		s.TLSConfig = &tls.Config{}
	}
	if len(s.TLSConfig.NextProtos) == 0 {
		s.TLSConfig.NextProtos = []string{"h2", "http/1.1"}
	}

	// clone tls config
	config := s.TLSConfig.Clone()

	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil || config.GetConfigForClient != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}

	// listen 80
	if s.Listen80RedirectTo443 && strings.HasPrefix(l.Addr().Network(), "tcp") {
		addr := l.Addr().String()
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			s.log("hlfhr: listen 80 error: net.SplitHostPort: ", err)
		} else if port == "443" {
			addr = net.JoinHostPort(host, "80")
			l80, err := net.Listen(l.Addr().Network(), addr)
			if err != nil {
				s.log("hlfhr: listen 80 error: net.Listen: ", err)
			} else {
				defer l80.Close()
				go func() {
					for {
						c, err := l80.Accept()
						if err != nil {
							return
						}
						go func(c net.Conn) {
							defer c.Close()
							(&conn{
								Conn: c,
								srv:  s,
							}).serve(nil, 0)
						}(c)
					}
				}()
			}
		}
	}

	// serve
	return s.Server.Serve(&TLSListener{
		Listener: l,
		TLSConf:  config,
		Server:   s,
	})
}

// ServeTLS accepts incoming HTTPS connections on the listener l,
// creating a new service goroutine for each. The service goroutines
// read requests and then call handler to reply to them.
//
// The handler is typically nil, in which case [http.DefaultServeMux] is used.
//
// Additionally, files containing a certificate and matching private key
// for the server must be provided. If the certificate is signed by a
// certificate authority, the certFile should be the concatenation
// of the server's certificate, any intermediates, and the CA's certificate.
//
// ServeTLS always returns a non-nil error.
func ServeTLS(l net.Listener, handler http.Handler, certFile, keyFile string) error {
	return New(&http.Server{
		Handler: handler,
	}).ServeTLS(l, certFile, keyFile)
}

// ListenAndServeTLS listens on the TCP network address srv.Addr and
// then calls [ServeTLS] to handle requests on incoming TLS connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// Filenames containing a certificate and matching private key for the
// server must be provided if neither the [Server]'s TLSConfig.Certificates
// nor TLSConfig.GetCertificate are populated. If the certificate is
// signed by a certificate authority, the certFile should be the
// concatenation of the server's certificate, any intermediates, and
// the CA's certificate.
//
// If srv.Addr is blank, ":https" is used.
//
// ListenAndServeTLS always returns a non-nil error. After [Server.Shutdown] or
// [Server.Close], the returned error is [http.ErrServerClosed].
func (s *Server) ListenAndServeTLS(certFile string, keyFile string) error {
	if s.IsShuttingDown() {
		return http.ErrServerClosed
	}
	addr := s.Addr
	if addr == "" {
		addr = ":https"
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	return s.ServeTLS(l, certFile, keyFile)
}

// ListenAndServeTLS acts identically to [http.ListenAndServe], except that it
// expects HTTPS connections. Additionally, files containing a certificate and
// matching private key for the server must be provided. If the certificate
// is signed by a certificate authority, the certFile should be the concatenation
// of the server's certificate, any intermediates, and the CA's certificate.
func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	return New(&http.Server{
		Addr:    addr,
		Handler: handler,
	}).ListenAndServeTLS(certFile, keyFile)
}

func IsHttpServerShuttingDown(srv *http.Server) bool {
	// Get private value
	inShutdown := (*atomic.Bool)(unsafe.Pointer(reflect.ValueOf(srv).Elem().FieldByName("inShutdown").UnsafeAddr()))
	return inShutdown.Load()
}

func (s *Server) IsShuttingDown() bool {
	return IsHttpServerShuttingDown(s.Server)
}

func (s *Server) log(v ...any) {
	if s.ErrorLog != nil {
		s.ErrorLog.Print(v...)
	} else {
		log.Print(v...)
	}
}

func (s *Server) logf(format string, v ...any) {
	if s.ErrorLog != nil {
		s.ErrorLog.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}
