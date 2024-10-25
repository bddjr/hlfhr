package hlfhr

// Reference:
//  - https://developer.mozilla.org/docs/Web/HTTP/Methods
//  - https://tls12.xargs.org/#client-hello
//  - https://tls13.xargs.org/#client-hello
func ConnFirstByteLooksLikeHttp(firstByte byte) bool {
	switch firstByte {
	case 'G', // GET
		'H', // HEAD
		'P', // POST PUT PATCH
		'O', // OPTIONS
		'D', // DELETE
		'C', // CONNECT
		'T': // TRACE

		return true
	}
	return false
}
