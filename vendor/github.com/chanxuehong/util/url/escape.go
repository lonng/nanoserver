package url

// QueryEscape 和 url.QueryEscape 基本一样, 除了空格 ' ' 转义为 "%20" 而不是 '+'
func QueryEscape(s string) string {
	hexCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			hexCount++
		}
	}

	if hexCount == 0 {
		return s
	}

	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case shouldEscape(c):
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

// QueryEscapeBytes 和 QueryEscape 功能一样
func QueryEscapeBytes(s []byte) []byte {
	hexCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			hexCount++
		}
	}

	if hexCount == 0 {
		return s
	}

	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case shouldEscape(c):
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return t
}

// 对于字符 A-Z、a-z、0-9 以及字符 '-', '_', '.', '~' 不编码
func shouldEscape(c byte) bool {
	if 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || 'A' <= c && c <= 'Z' {
		return false
	}
	switch c {
	case '-', '_', '.', '~':
		return false
	}
	return true
}
