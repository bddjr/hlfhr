// HTTPS Listener For HTTP Redirect
//
// BSD-3-clause license
package hlfhr

import (
	"context"
	"net"
	"net/http"
)

type Server struct {
	*http.Server

	// HttpOnHttpsPortErrorHandler handles HTTP requests sent to an HTTPS port.
	// See https://github.com/bddjr/hlfhr#httponhttpsporterrorhandler
	HttpOnHttpsPortErrorHandler http.Handler

	shuttingDown bool
}

// New hlfhr Server
func New(s *http.Server) *Server {
	return &Server{Server: s}
}

// New hlfhr Server
func NewServer(s *http.Server) *Server {
	return &Server{Server: s}
}

func (s *Server) NewListener(l net.Listener) net.Listener {
	return NewListener(
		l,
		s.Server,
		s.HttpOnHttpsPortErrorHandler,
	)
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
	l = s.NewListener(l)
	return s.Server.ServeTLS(l, certFile, keyFile)
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
	srv := New(&http.Server{Handler: handler})
	return srv.ServeTLS(l, certFile, keyFile)
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
	if s.shuttingDown {
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
	srv := New(&http.Server{
		Addr:    addr,
		Handler: handler,
	})
	return srv.ListenAndServeTLS(certFile, keyFile)
}

// Close immediately closes all active net.Listeners and any
// connections in state [http.StateNew], [http.StateActive], or [http.StateIdle]. For a
// graceful shutdown, use [Server.Shutdown].
//
// Close does not attempt to close (and does not even know about)
// any hijacked connections, such as WebSockets.
//
// Close returns any error returned from closing the [Server]'s
// underlying Listener(s).
func (s *Server) Close() error {
	s.shuttingDown = true
	return s.Server.Close()
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
// If the provided context expires before the shutdown is complete,
// Shutdown returns the context's error, otherwise it returns any
// error returned from closing the [Server]'s underlying Listener(s).
//
// When Shutdown is called, [http.Serve], [http.ListenAndServe], and
// [ListenAndServeTLS] immediately return [http.ErrServerClosed]. Make sure the
// program doesn't exit and waits instead for Shutdown to return.
//
// Shutdown does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of Shutdown should
// separately notify such long-lived connections of shutdown and wait
// for them to close, if desired. See [Server.RegisterOnShutdown] for a way to
// register shutdown notification functions.
//
// Once Shutdown has been called on a server, it may not be reused;
// future calls to methods such as Serve will return [http.ErrServerClosed].
func (s *Server) Shutdown(ctx context.Context) error {
	s.shuttingDown = true
	return s.Server.Shutdown(ctx)
}

func (s *Server) IsShuttingDown() bool {
	return s.shuttingDown
}
