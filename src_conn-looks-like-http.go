package hlfhr

// Reference:
//   - https://developer.mozilla.org/docs/Web/HTTP/Methods
//   - https://tls12.xargs.org/#client-hello
//   - https://tls13.xargs.org/#client-hello
func ConnFirstByteLooksLikeHttp(firstByte byte) bool {
	switch firstByte {
	case 0x16:
		// TLS handshake
		return false

	case 'G', // GET
		'H', // HEAD
		'P', // POST PUT PATCH
		'O', // OPTIONS
		'D', // DELETE
		'C', // CONNECT
		'T': // TRACE

		return true
	}

	// error
	return false
}
