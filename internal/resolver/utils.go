package resolver

import (
	"strconv"
	"strings"

	"github.com/diplexhq/diplex/internal/domain"
	"github.com/diplexhq/diplex/internal/utils"
)

// normalizeGenericParameter replaces all generic type arguments inside [...] with "T",
// preserving the argument count. Each top-level comma-separated argument becomes one "T".
//
// Algorithm: token-by-token scan with inBracket flag.
//  1. When NOT in brackets: use utils.splitToken for normal scanning
//  2. When IN brackets (inBracket=true): use utils.NextConcreteToken to extract full top-level args
//  3. If token is T\d* → emit "T"
//  4. If token is '[' and lastToken != "map" → enter bracket mode (inBracket = true)
//  5. If token is ']' → always exit bracket mode
//  6. Otherwise → emit token as-is
//
// Examples:
//
//	pkg.Cache[T0]                                        → pkg.Cache[T]
//	pkg.Repo[T0, T1]                                     → pkg.Repo[T, T]
//	pkg.Cache[order.Order, payment.Payment]              → pkg.Cache[T, T]
//	pkg.Cache[repo.Repo[T0], T1]                         → pkg.Cache[T, T]
//	*pkg.Cache[T0, T1]                                   → *pkg.Cache[T, T]
//	pkg.Type[]                                           → pkg.Type[]
//	pkg.Order                                            → pkg.Order (no brackets, no change)
//	map[int]string                                       → map[int]string (map skipped)
func (res *Resolver) normalizeGenericParameter(s string) string {
	var (
		result           strings.Builder
		lastToken, token string // tracks EVERY token (not just idents)
		inBracket        bool   // true when inside [...] of a generic type
	)
	result.Grow(len(s))

	for len(s) > 0 {
		lastToken = token

		if inBracket {
			token, s = utils.NextConcreteToken(s)
		} else {
			token, s = utils.NextIdentToken(s)
		}

		switch {
		// If token is T\d* → output T
		case utils.IsGenericParam(token):
			result.WriteByte('T')
		// If token is '[' and lastToken is an ident but not "map"
		case token == "[" && len(lastToken) > 0 && utils.IsIdentBaseChar(lastToken[0]) && lastToken != "map":
			// Check if next char is ']' → empty brackets
			result.WriteByte('[')

			inBracket = true
		case token == "]":
			result.WriteByte(']')

			inBracket = false
		case inBracket:
			if len(token) > 1 {
				result.WriteByte('T')
			} else {
				result.WriteByte(token[0])
			}
		default:
			result.WriteString(token)
		}
	}

	r := result.String()

	return r
}

// extractGenericPrototype extracts the generic base type and constraints from a result like "Repo[pkg.User]".
// Returns ("Repo[T]", constraints) or (result, nil) if not generic.
func (res *Resolver) extractGenericPrototype(provider *domain.Provider) (string, map[string][]string) {
	result := strings.TrimPrefix(string(provider.Result), "*")
	if before, after, ok := strings.Cut(result, "["); ok && len(before) > 0 && before != "map" {
		suffix, constraints := res.extractConstraintsFromInstantiation(after)
		return before + suffix, constraints
	}

	return result, nil
}

// extractConstraintsFromInstantiation parses the inner part of a generic instantiation like "pkg.User"
// and returns the bracket suffix "[T]" or "[T0, T1, …]" plus the constraint map for narrowing.
func (res *Resolver) extractConstraintsFromInstantiation(inner string) (string, map[string][]string) {
	closeBracket := strings.LastIndex(inner, "]")
	if closeBracket != -1 {
		inner = inner[:closeBracket]
	}

	// Inlined splitGenericArgs: split generic args by top-level commas respecting bracket nesting
	var parts []string

	for len(inner) > 0 {
		token, rest := utils.NextConcreteToken(inner)
		if len(token) > 1 || utils.IsIdentChar(token[0]) {
			parts = append(parts, token)
		}

		inner = rest
	}

	switch len(parts) {
	case 0:
		return "", nil
	case 1:
		return "[T]", map[string][]string{"T": {strings.TrimSpace(parts[0])}}
	}

	constraints := make(map[string][]string)

	for i, part := range parts {
		name := "T" + strconv.FormatInt(int64(i), 10)
		constraints[name] = []string{strings.TrimSpace(part)}
	}

	return utils.TList(len(parts)), constraints
}

// methodKey builds an index key for a method using idiomatic Go syntax:
//
//	Generic method → "Name"
//	Non-generic, no results → "Name()" or "Name(args)"
//	Non-generic, one result → "Name() result" or "Name(args) result"
//	Non-generic, two+ results → "Name() (r1, r2)" or "Name(args) (r1, r2)"
func (res *Resolver) methodKey(methodName domain.FunctionName, contract domain.MethodContract) string {
	// Fast pass: check for generic tokens in parameters first
	for _, p := range contract.Arguments {
		if utils.HasGenericToken(string(p)) {
			return string(methodName)
		}
	}

	for _, r := range contract.Results {
		if utils.HasGenericToken(string(r)) {
			return string(methodName)
		}
	}

	var b strings.Builder
	b.Grow(128)
	b.WriteString(string(methodName))

	// Params
	b.WriteByte('(')

	for i, p := range contract.Arguments {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(string(p))
	}

	b.WriteByte(')')
	// Results
	switch len(contract.Results) {
	case 0:
		// nothing
	case 1:
		b.WriteByte(' ')
		b.WriteString(string(contract.Results[0]))
	default:
		b.WriteString(" (")

		for i, r := range contract.Results {
			if i > 0 {
				b.WriteString(", ")
			}

			b.WriteString(string(r))
		}

		b.WriteByte(')')
	}

	return b.String()
}
