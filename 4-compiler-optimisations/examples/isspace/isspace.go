package isspace

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
func isSpace1(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}

func isSpace2(ch byte) bool {
	return !(ch > ' '+1) && (ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r')
}

var wsm = map[byte]bool{
	' ':  true,
	'\t': true,
	'\n': true,
	'\r': true,
}

func isSpace3(ch byte) bool { return wsm[ch] }

var wsp = [256]bool{
	' ':  true,
	'\t': true,
	'\n': true,
	'\r': true,
}

func isSpace4(ch byte) bool { return wsp[ch] }
