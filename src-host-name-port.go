package hlfhr

import "strings"

// "[::1]:5678" => "[::1]", "5678"
func SplitHostnamePort(Host string) (hostname string, port string) {
	if !strings.HasSuffix(Host, "]") {
		if i := strings.LastIndexByte(Host, ':'); i != -1 {
			return Host[:i], Host[i+1:]
		}
	}
	return Host, ""
}

// "[::1]:5678" => "[::1]"
func Hostname(Host string) (hostname string) {
	hostname, _ = SplitHostnamePort(Host)
	return
}

// "[::1]:5678" => "5678"
func Port(Host string) (port string) {
	_, port = SplitHostnamePort(Host)
	return
}

// "[::1]",	"5678"	=> "[::1]:5678"
//
// "[::1]",	":5678"	=> "[::1]:5678"
//
// "[::1]",	""		=> "[::1]"
//
// "[::1]",	":"		=> "[::1]"
//
// "::1"  ,	"5678"	=> "[::1]:5678"
//
// "::1"  ,	":5678"	=> "[::1]:5678"
//
// "::1"  ,	""		=> "[::1]"
//
// "::1"  ,	":"		=> "[::1]"
func HostnameAppendPort(hostname string, port string) string {
	if strings.Contains(hostname, ":") {
		if !strings.HasPrefix(hostname, "[") {
			hostname = "[" + hostname
		}
		if !strings.HasSuffix(hostname, "]") {
			hostname += "]"
		}
	}
	switch port {
	case "", ":":
		return hostname
	}
	if strings.HasPrefix(port, ":") {
		return hostname + port
	}
	return hostname + ":" + port
}

// "[::1]:5678", "localhost" => "localhost:5678"
func ReplaceHostname(Host string, name string) string {
	return HostnameAppendPort(name, Port(Host))
}

// "[::1]:5678", "7890" => "localhost:7890"
func ReplacePort(Host string, port string) string {
	return HostnameAppendPort(Hostname(Host), port)
}

// "[::1]" => "::1"
func Ipv6CutPrefixSuffix(v6 string) string {
	v6, _ = strings.CutPrefix(v6, "[")
	v6, _ = strings.CutSuffix(v6, "]")
	return v6
}
