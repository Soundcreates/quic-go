package http3

import "strings"

const defaultPriorityUrgency int8 = 3

// parsePriority parses the RFC 9218 Priority header field value, returning the given defaults
// for the parameters that the field doesn't (validly) set. If the field value is not a valid
// structured field dictionary (RFC 9651), both defaults are returned.
//
// This is not a full structured field parser: it only splits the value into its top-level
// dictionary members, and only inspects the "u" and "i" keys. Strings are the only member
// values that may contain a comma, so tracking them is enough to keep extensions using any
// other type from being misinterpreted.
//
// We intentionally don't use github.com/dunglas/httpsfv to keep quic-go dependency-free.
// The proposal to export an SFV parser from the Go standard library is not implemented yet:
// https://github.com/golang/go/issues/41046
func parsePriority(value string, defaultUrgency int8, defaultIncremental bool) (urgency int8, incremental bool) {
	urgency, incremental = defaultUrgency, defaultIncremental
	var inString, escapes bool
	for i, start := 0, 0; i <= len(value); i++ {
		switch {
		case i == len(value) || (!inString && value[i] == ','):
			// Cutting parameters at the first semicolon is only correct if the member value
			// doesn't contain one, but "u" and "i" values never do.
			member, _, _ := strings.Cut(strings.TrimSpace(value[start:i]), ";")
			start = i + 1
			key, item, hasValue := strings.Cut(member, "=")
			switch key {
			case "u":
				if hasValue && len(item) == 1 && item[0] >= '0' && item[0] <= '7' {
					urgency = int8(item[0] - '0')
				}
			case "i":
				if !hasValue || item == "?1" {
					incremental = true
				} else if item == "?0" {
					incremental = false
				}
			}
		case escapes && value[i] == '\\':
			i++
		case value[i] == '"':
			inString = !inString
			// display strings (%"...") use percent-encoding instead of backslash escapes
			escapes = inString && (i == 0 || value[i-1] != '%')
		}
	}
	if inString { // unterminated string: not a valid dictionary
		return defaultUrgency, defaultIncremental
	}
	return urgency, incremental
}
