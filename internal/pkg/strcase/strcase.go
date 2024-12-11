package strcase

import (
	"strings"
)

var acronyms = map[string]string{
	"Id":   "ID",
	"Ip":   "IP",
	"Url":  "URL",
	"Uuid": "UUID",
}
var maxLen, minLen = 4, 2

// Converts a string to CamelCase
func ToCamel(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	n := strings.Builder{}
	n.Grow(len(s))
	capNext := true
	prevIsCap := false
	for i, v := range []byte(s) {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'
		if capNext {
			if vIsLow {
				v += 'A'
				v -= 'a'
			}
		} else if i == 0 {
			if vIsCap {
				v += 'a'
				v -= 'A'
			}
		} else if prevIsCap && vIsCap {
			v += 'a'
			v -= 'A'
		}
		prevIsCap = vIsCap

		if vIsCap || vIsLow {
			n.WriteByte(v)
			capNext = false
		} else if vIsNum := v >= '0' && v <= '9'; vIsNum {
			n.WriteByte(v)
			capNext = true
		} else {
			capNext = v == '_' || v == ' ' || v == '-' || v == '.'
		}
	}

	s = n.String()

	for i := minLen; i <= maxLen; i++ {
		if len(s) < i {
			continue
		}
		strEnd := s[len(s)-i:]
		if replaceVal, ok := acronyms[strEnd]; ok {
			strStart := s[:len(s)-i]
			s = strStart + replaceVal
			break
		}
	}

	return s
}
