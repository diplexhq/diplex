package utils

import (
	"strconv"
	"strings"
)

// ToSnakeCase converts a CamelCase string to snake_case.
func ToSnakeCase(str string) string {
	var buf strings.Builder
	buf.Grow(len(str) + 4)

	for i := range len(str) - 1 {
		if i > 0 && str[i] >= 'A' && str[i] <= 'Z' && str[i+1] >= 'a' && str[i+1] <= 'z' {
			buf.WriteByte('_')
		}

		buf.WriteByte(toLower(str[i]))
	}

	buf.WriteByte(toLower(str[len(str)-1]))

	return buf.String()
}

// ReplaceTokens replaces all occurrences of keys from the replacements map
// in the input string. Tokens are qualified Go identifiers — sequences of
// letters, digits, underscores, dots, and slashes (e.g. "pkg.Type",
// "github.com/user/pkg.X"). Non-token characters are copied verbatim.
//
// This is a hand-rolled replacement that avoids regex allocation overhead.
// Regex.ReplaceAllStringFunc creates a closure call per match; the hand-rolled
// version uses a single strings.Builder and no closures.
func ReplaceTokens(t string, replacements map[string]string) string {
	return resolveReplacement(t, replacements, false, 1)
}

// IsIdentChar reports whether a byte is valid inside a qualified Go identifier.
// This matches [\w./-]+ — letters, digits, underscore, dot (package qualifier),
// slash (module path separator), and hyphen (package path separator).
func IsIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || b == '_' || b == '.' || b == '/' || b == '-'
}

// IsIdentBaseChar reports whether b is a valid Go identifier character (letter, digit, underscore).
// Unlike IsIdentChar, this does NOT include '.' or '/' — it matches the standard Go identifier
// character set only.
func IsIdentBaseChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || b == '_'
}

// SanitizeIdent converts a string into a valid camelCase Go identifier.
// It splits on non-identifier characters (brackets, dots, slashes), capitalizes
// each segment, and joins them — producing idiomatic camelCase names.
//
// Examples:
//
//	"entityNewRepo[entity.Order]"       → "entityNewRepoEntityOrder"
//	"handlerOrderCreateNew"             → "handlerOrderCreateNew"
//	"configNewConfig"                   → "configNewConfig"
//	"NewRepo[T]"                        → "newRepoT"
func SanitizeIdent(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))

	i := 0
	for i < len(s) {
		if !IsIdentBaseChar(s[i]) {
			i++
			continue
		}

		// Collect one contiguous identifier segment.
		start := i
		for i < len(s) && IsIdentBaseChar(s[i]) {
			i++
		}

		seg := s[start:i]

		if buf.Len() == 0 {
			buf.WriteByte(toLower(seg[0]))
		} else {
			buf.WriteByte(toUpper(seg[0]))
		}

		buf.WriteString(seg[1:])
	}

	return buf.String()
}

// HasGenericToken reports whether s contains a standalone generic parameter token
// (T, T0, T1, …). A token is "standalone" if it is not preceded or followed by an
// identifier character (letter, digit, underscore, dot, slash).
//
// Equivalent to the regex: (^|[^a-zA-Z0-9_./])T\d*($|[^a-zA-Z0-9_./])
// but ~50× faster — avoids RE2 engine overhead for a simple byte scan.
//
// This is used to distinguish generic method signatures from concrete ones when
// building the interface → provider index.
func HasGenericToken(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == 'T' {
			if i > 0 && IsIdentChar(s[i-1]) {
				continue
			}

			j := i + 1
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}

			if j == len(s) || !IsIdentChar(s[j]) {
				return true
			}
		}
	}

	return false
}

// NextConcreteToken extracts the first concrete type from s.
// Balances square brackets [], parentheses (), and braces {} independently.
// Stops at ',' or ']' when all nesting depths are 0.
//
// If the string starts with ',', ']' or space — returns that character as a token.
func NextConcreteToken(s string) (token, rest string) {
	if len(s) == 0 {
		return "", s
	}

	// Leading separator or closing bracket — return as single-char token
	if s[0] == ',' || s[0] == ']' || s[0] == ' ' {
		return s[:1], s[1:]
	}

	depthSq := 0
	depthPn := 0
	depthBr := 0

	end := 0
	for end < len(s) {
		ch := s[end]
		switch ch {
		case '[':
			depthSq++
		case ']':
			if depthSq > 0 {
				depthSq--
			} else {
				return s[:end], s[end:]
			}
		case '(':
			depthPn++
		case ')':
			if depthPn > 0 {
				depthPn--
			} else {
				return s[:end], s[end:]
			}
		case '{':
			depthBr++
		case '}':
			if depthBr > 0 {
				depthBr--
			} else {
				return s[:end], s[end:]
			}
		case ',':
			if depthSq == 0 && depthPn == 0 && depthBr == 0 {
				return s[:end], s[end:]
			}
		}

		end++
	}

	return s[:end], s[end:]
}

// NextIdentToken extracts the first token from s and returns (token, remainder).
//
// Token rules:
//
//	Ident sequences: a-z, A-Z, 0-9, _ (e.g., "cache", "Cache", "T0", "10")
//	Single non-ident chars: ., [, ], *, ~, (, ), {, }, ,, etc.
func NextIdentToken(s string) (string, string) {
	if len(s) == 0 {
		return "", ""
	}

	// Check if first char starts an ident
	if IsIdentBaseChar(s[0]) {
		i := 1
		for i < len(s) && IsIdentChar(s[i]) {
			i++
		}

		return s[:i], s[i:]
	}

	// Single char non-ident token
	return s[:1], s[1:]
}

// IsGenericParam reports whether s is a generic type parameter: T, T0, T1, T99, etc.
func IsGenericParam(s string) bool {
	if len(s) == 0 || s[0] != 'T' {
		return false
	}

	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

// IsExported reports whether name is an exported Go identifier.
// In Go, an identifier is exported if it starts with an uppercase letter.
func IsExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}

// ResolveReplacements recursively resolves token replacements by processing
// replacement values through the same map before substituting. This handles
// chained dependencies where a replacement value itself contains tokens that
// need resolution.
//
// For example, with replacements {"A": "B", "B": "C"}, ResolveReplacements("A", ...)
// will resolve to "C" (A→B→C), not "B".
//
// Parameters:
//   - t: the input string containing tokens to replace
//   - replacements: map of token→replacement mappings
//   - depth: current recursion depth (starts at 1 from caller)
//   - maxDepth: maximum allowed recursion depth before panic
//   - visited: set of keys currently being resolved (for cycle detection)
//
// Panics with "circular replacement detected" if a cycle is found.
// Panics with "recursion depth exceeded" if maxDepth is reached.
func ResolveReplacements(replacements map[string]string) {
	for k, v := range replacements {
		replacements[k] = resolveReplacement(v, replacements, true, len(replacements))
	}
}

// resolveReplacement recursively resolves a replacement value.
// It checks each token in the string — if a token is a key in replacements,
// it marks it as visited, recurses to resolve the replacement, then unmarks it.
func resolveReplacement(s string, replacements map[string]string, inDeep bool, maxDepth int) string {
	if maxDepth == 0 {
		panic("recursion detected")
	}

	var buf strings.Builder
	buf.Grow(len(s))

	i := 0

	for i < len(s) {
		start := i
		for i < len(s) && IsIdentChar(s[i]) {
			i++
		}

		if i > start {
			token := s[start:i]
			if repl, ok := replacements[token]; ok {
				if inDeep {
					if resolved := resolveReplacement(repl, replacements, inDeep, maxDepth-1); repl != resolved {
						repl = resolved
						replacements[token] = resolved
					}
				}

				buf.WriteString(repl)
			} else {
				buf.WriteString(token)
			}
		} else {
			buf.WriteByte(s[i])
			i++
		}
	}

	return buf.String()
}

// toLower lowercases a single byte to its lowercase equivalent.
func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 'a' - 'A'
	}

	return c
}

// toUpper uppercases a single byte to its uppercase equivalent.
func toUpper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c + 'A' - 'a'
	}

	return c
}

func TList(count int) string {
	switch count {
	case 0:
		return ""
	case 1:
		return "[T]"
	}

	sb := strings.Builder{}
	sb.Grow(count * 4)
	sb.WriteString("[T0")

	for i := 1; i < count; i++ {
		sb.WriteString(", T")
		sb.WriteString(strconv.FormatInt(int64(i), 10))
	}

	sb.WriteRune(']')

	return sb.String()
}
