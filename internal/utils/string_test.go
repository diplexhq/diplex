package utils

import (
	"testing"
)

func TestHasGenericToken(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		s    string
		want bool
	}{
		// --- plain T ---
		{name: "bare T", s: "T", want: true},
		{name: "T at end", s: "pkg.Repo[T]", want: true},
		{name: "T at start", s: "T, error", want: true},
		{name: "T in middle", s: "(T, int)", want: true},
		{name: "T not standalone — Repo", s: "Repo", want: false},
		{name: "T not standalone — Target", s: "Target", want: false},
		{name: "T not standalone — Type", s: "Type", want: false},
		{name: "T not standalone — Tmp", s: "Tmp", want: false},
		{name: "T not standalone — myT", s: "myT", want: false},

		// --- T0, T1, T2, … ---
		{name: "bare T0", s: "T0", want: true},
		{name: "T1", s: "T1", want: true},
		{name: "T99", s: "T99", want: true},
		{name: "T0 in brackets", s: "*pkg.Cache[T0, T1]", want: true},
		{name: "T0 not standalone — Token", s: "Token", want: false},
		{name: "T100", s: "T100", want: true},

		// --- mixed ---
		{name: "mixed generic and concrete", s: "*pkg.Repo[T]", want: true},
		{name: "no generics", s: "pkg.TypeA", want: false},
		{name: "empty string", s: "", want: false},

		// --- edge: T as part of larger token ---
		{name: "TT", s: "TT", want: false},
		{name: "T_T", s: "T_T", want: false},
		{name: "T.pkg", s: "T.pkg", want: false},
		{name: "T/pkg", s: "T/pkg", want: false},

		// --- digit suffix boundaries ---
		{name: "T1a — digit then letter", s: "T1a", want: false},
		{name: "T1. — digit then dot", s: "T1.", want: false},
		{name: "T1/pkg — digit then slash", s: "T1/pkg", want: false},
		{name: "T1_ — digit then underscore", s: "T1_", want: false},
		{name: "T19 — digit then digit end", s: "T19", want: true},
		{name: "T1 space — digit then space", s: "T1 ", want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := HasGenericToken(tc.s)
			if got != tc.want {
				t.Errorf("HasGenericToken(%q) = %v, want %v", tc.s, got, tc.want)
			}
		})
	}
}

func TestIsIdentBaseChar(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		ch   byte
		want bool
	}{
		{"lowercase a", 'a', true},
		{"lowercase z", 'z', true},
		{"uppercase A", 'A', true},
		{"uppercase Z", 'Z', true},
		{"digit 0", '0', true},
		{"digit 9", '9', true},
		{"underscore", '_', true},
		{"dot", '.', false},
		{"slash", '/', false},
		{"space", ' ', false},
		{"open paren", '(', false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := IsIdentBaseChar(tc.ch)
			if got != tc.want {
				t.Errorf("IsIdentBaseChar(%q) = %v, want %v", tc.ch, got, tc.want)
			}
		})
	}
}

func TestIsIdentChar(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		ch   byte
		want bool
	}{
		{"lowercase a", 'a', true},
		{"lowercase z", 'z', true},
		{"uppercase A", 'A', true},
		{"uppercase Z", 'Z', true},
		{"digit 0", '0', true},
		{"digit 9", '9', true},
		{"underscore", '_', true},
		{"dot", '.', true},
		{"slash", '/', true},
		{"space", ' ', false},
		{"open paren", '(', false},
		{"close paren", ')', false},
		{"open bracket", '[', false},
		{"close bracket", ']', false},
		{"comma", ',', false},
		{"star", '*', false},
		{"tilde", '~', false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := IsIdentChar(tc.ch)
			if got != tc.want {
				t.Errorf("IsIdentChar(%q) = %v, want %v", tc.ch, got, tc.want)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"single lowercase", "a", "a"},
		{"single uppercase", "A", "a"},
		{"already snake", "already_snake", "already_snake"},
		{"lowercase", "lowercase", "lowercase"},
		{"CamelCase", "CamelCase", "camel_case"},
		{"HelloWorld", "HelloWorld", "hello_world"},
		{"FooBarBaz", "FooBarBaz", "foo_bar_baz"},
		{"aB", "aB", "ab"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ToSnakeCase(tc.in)
			if got != tc.want {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestReplaceTokens(t *testing.T) {
	t.Parallel()

	repl := map[string]string{
		"entity.Order": "diplexdi.EntityOrder",
		"Config":       "NewConfig",
		"T":            "int",
	}

	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"no replacements", "Hello World", "Hello World"},
		{"empty input", "", ""},
		{"single token", "entity.Order", "diplexdi.EntityOrder"},
		{"token with space before", " func(entity.Order)", " func(diplexdi.EntityOrder)"},
		{"multiple tokens", "entity.Order[T]", "diplexdi.EntityOrder[int]"},
		{"single long token no match", "github.com/user/pkg.Config", "github.com/user/pkg.Config"},
		{"no map", "no map", "no map"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ReplaceTokens(tc.in, repl)
			if got != tc.want {
				t.Errorf("ReplaceTokens(%q, _) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}

	t.Run("empty map returns original", func(t *testing.T) {
		t.Parallel()

		got := ReplaceTokens("hello", map[string]string{})
		if got != "hello" {
			t.Errorf("ReplaceTokens with empty map = %q, want %q", got, "hello")
		}
	})
}

func TestSanitizeIdent(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"simple identifier", "hello", "hello"},
		{"already camelCase", "helloWorld", "helloWorld"},
		{"with brackets", "entityNewRepo[entity.Order]", "entityNewRepoEntityOrder"},
		{"with parens", "NewRepo(T)", "newRepoT"},
		{"mixed separators", "cache[entity.Order]", "cacheEntityOrder"},
		{"single char", "a", "a"},
		{"uppercase start", "Hello", "hello"},
		{"multiple spaces", "  hello  world  ", "helloWorld"},
		{"with slash", "github.com/user/pkg", "githubComUserPkg"},
		{"with dots", "pkg.Type", "pkgType"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := SanitizeIdent(tc.in)
			if got != tc.want {
				t.Errorf("SanitizeIdent(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNextConcreteToken(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		in       string
		want     string
		wantRest string
	}{
		// --- basic cases ---
		{"simple", "string, error", "string", ", error"},
		{"comma first", ",rest", ",", "rest"},
		{"space first", " rest", " ", "rest"},
		{"single token", "string", "string", ""},
		{"empty", "", "", ""},
		{"unmatched close bracket", "]rest", "]", "rest"},

		// --- map types ---
		{"map literal type", "map[string]int, error", "map[string]int", ", error"},
		{"map with complex key and value", "map[entity.Key]*entity.Value, rest", "map[entity.Key]*entity.Value", ", rest"},
		{"nested map type", "map[string]map[int]string, rest", "map[string]map[int]string", ", rest"},
		{"map with generic value", "map[string]pkg.Cache[T]", "map[string]pkg.Cache[T]", ""},
		{"map with pointer value", "map[string]*pkg.Item", "map[string]*pkg.Item", ""},
		{"map with struct value", "map[string]struct{ Name string }", "map[string]struct{ Name string }", ""},
		{"map with channel value", "map[string]chan int, rest", "map[string]chan int", ", rest"},
		{"map with function value type", "map[string]func(int) string, rest", "map[string]func(int) string", ", rest"},

		// --- channel types ---
		{"channel type", "chan int, error", "chan int", ", error"},
		{"chan string direction", "chan<- string, rest", "chan<- string", ", rest"},
		{"<-chan direction", "<-chan int, rest", "<-chan int", ", rest"},
		{"chan with package type", "chan pkg.Message, rest", "chan pkg.Message", ", rest"},
		{"chan of pointers", "chan *pkg.Item, rest", "chan *pkg.Item", ", rest"},
		{"nested channel", "chan chan int, rest", "chan chan int", ", rest"},

		// --- generic types with multiple params ---
		{"generic single param", "pkg.Cache[T], rest", "pkg.Cache[T]", ", rest"},
		{"generic multiple params", "pkg.Service[T, U], rest", "pkg.Service[T, U]", ", rest"},
		{"generic nested", "pkg.Wrapper[pkg.Inner[T]], rest", "pkg.Wrapper[pkg.Inner[T]]", ", rest"},
		{"generic with map", "map[K]pkg.Cache[V], rest", "map[K]pkg.Cache[V]", ", rest"},
		{"generic with channel", "chan pkg.Cache[T], rest", "chan pkg.Cache[T]", ", rest"},
		{"generic with pointer", "*pkg.Cache[T], rest", "*pkg.Cache[T]", ", rest"},

		// --- complex nested types ---
		{"pointer to map", "*map[string]int, rest", "*map[string]int", ", rest"},
		{"pointer to channel", "*chan int, rest", "*chan int", ", rest"},
		{"pointer to generic", "*pkg.Cache[T], rest", "*pkg.Cache[T]", ", rest"},
		{"slice of maps", "[]map[string]int, rest", "[]map[string]int", ", rest"},
		{"array of channels", "[2]chan int, rest", "[2]chan int", ", rest"},

		// --- function types in braces ---
		{"function type", "func(int) string, error", "func(int) string", ", error"},
		{"function with result in parens", "func(int) (string, error), rest", "func(int) (string, error)", ", rest"},
		{"method receiver style", "func(ctx context.Context) error", "func(ctx context.Context) error", ""},
		{"generic function", "func(T) pkg.Cache[T], rest", "func(T) pkg.Cache[T]", ", rest"},

		// --- struct types with braces ---
		{"anonymous struct", "struct{ Name string }", "struct{ Name string }", ""},
		{"struct with multiple fields", "struct{ Name string; Age int }, rest", "struct{ Name string; Age int }", ", rest"},
		{"struct with embedded generic", "struct{ Data pkg.Cache[T] }, rest", "struct{ Data pkg.Cache[T] }", ", rest"},

		// --- deeply nested combinations ---
		{"map of chan of generic", "map[string]chan pkg.Cache[T], rest", "map[string]chan pkg.Cache[T]", ", rest"},
		{"generic with map and channel", "pkg.Service[map[string]chan int], rest", "pkg.Service[map[string]chan int]", ", rest"},
		{"complex pointer chain", "***pkg.Cache[T], rest", "***pkg.Cache[T]", ", rest"},
		{"slice of pointers to map", "[]*map[string]*int, rest", "[]*map[string]*int", ", rest"},

		// --- edge cases with brackets ---
		// Note: [ ( { at start are NOT treated as single-char tokens like ] ) }
		// They increment depth and consume the entire string if unbalanced
		{"unmatched open bracket consumes all", "[rest", "[rest", ""},
		{"unmatched open paren consumes all", "(rest", "(rest", ""},
		{"unmatched open brace consumes all", "{rest", "{rest", ""},
		{"balanced open bracket closed", "[rest]", "[rest]", ""},
		{"balanced open paren closed", "(rest)", "(rest)", ""},
		{"balanced open brace closed", "{rest}", "{rest}", ""},
		{"balanced deep nesting", "pkg[A[B[C[D]]]]", "pkg[A[B[C[D]]]]", ""},
		{"balanced mixed nesting", "map[chan[pkg[T]]]struct{}", "map[chan[pkg[T]]]struct{}", ""},

		// --- type with asterisk inside generic ---
		{"generic with pointer inside", "pkg.Cache[*string], rest", "pkg.Cache[*string]", ", rest"},
		{"generic with slice inside", "pkg.Cache[]int, rest", "pkg.Cache[]int", ", rest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, gotRest := NextConcreteToken(tc.in)
			if got != tc.want || gotRest != tc.wantRest {
				t.Errorf("NextConcreteToken(%q) = (%q, %q), want (%q, %q)", tc.in, got, gotRest, tc.want, tc.wantRest)
			}
		})
	}
}

func TestNextIdentToken(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		in       string
		want     string
		wantRest string
	}{
		{"identifier", "HelloWorld, rest", "HelloWorld", ", rest"},
		{"single char", "a rest", "a", " rest"},
		{"digit start", "123abc rest", "123abc", " rest"},
		{"bracket first", "[rest", "[", "rest"},
		{"comma first", ",rest", ",", "rest"},
		{"dots in ident", "pkg.Type rest", "pkg.Type", " rest"},
		{"slashes in ident", "github.com/pkg rest", "github.com/pkg", " rest"},
		{"empty", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, gotRest := NextIdentToken(tc.in)
			if got != tc.want || gotRest != tc.wantRest {
				t.Errorf("NextIdentToken(%q) = (%q, %q), want (%q, %q)", tc.in, got, gotRest, tc.want, tc.wantRest)
			}
		})
	}
}

func TestIsGenericParam(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		want bool
	}{
		{"T", "T", true},
		{"T0", "T0", true},
		{"T1", "T1", true},
		{"T99", "T99", true},
		{"t", "t", false},
		{"t1", "t1", false},
		{"TT", "TT", false},
		{"empty", "", false},
		{"Ta", "Ta", false},
		{"T1a", "T1a", false},
		{"A", "A", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := IsGenericParam(tc.in)
			if got != tc.want {
				t.Errorf("IsGenericParam(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsExported(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"single lowercase", "a", false},
		{"single uppercase", "A", true},
		{"lowercase", "hello", false},
		{"uppercase start", "Hello", true},
		{"camelCase", "helloWorld", false},
		{"PascalCase", "HelloWorld", true},
		{"digit", "1hello", false},
		{"underscore", "_hello", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := IsExported(tc.in)
			if got != tc.want {
				t.Errorf("IsExported(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestTList(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		count int
		want  string
	}{
		{name: "zero", count: 0, want: ""},
		{name: "one", count: 1, want: "[T]"},
		{name: "two", count: 2, want: "[T0, T1]"},
		{name: "three", count: 3, want: "[T0, T1, T2]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := TList(tc.count)
			if got != tc.want {
				t.Errorf("TList(%d) = %q, want %q", tc.count, got, tc.want)
			}
		})
	}
}

func TestResolveReplacements(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name             string
		replacements     map[string]string
		wantReplacements map[string]string
		wantPanic        string
	}{
		{
			name:             "empty map",
			replacements:     map[string]string{},
			wantReplacements: map[string]string{},
		},
		{
			name:             "no chain",
			replacements:     map[string]string{"A": "Concrete", "T": "int"},
			wantReplacements: map[string]string{"A": "Concrete", "T": "int"},
		},
		{
			name:             "simple chain A->B->C",
			replacements:     map[string]string{"A": "B", "B": "C"},
			wantReplacements: map[string]string{"A": "C", "B": "C"},
		},
		{
			name:             "chained replacement",
			replacements:     map[string]string{"A": "B", "B": "Concrete"},
			wantReplacements: map[string]string{"A": "Concrete", "B": "Concrete"},
		},
		{
			name:             "deep chain A->B->C->D->Final",
			replacements:     map[string]string{"A": "B", "B": "C", "C": "D", "D": "Final"},
			wantReplacements: map[string]string{"A": "Final", "B": "Final", "C": "Final", "D": "Final"},
		},
		{
			name:             "qualified type chain",
			replacements:     map[string]string{"pkg.A": "pkg.B", "pkg.B": "concrete.Type"},
			wantReplacements: map[string]string{"pkg.A": "concrete.Type", "pkg.B": "concrete.Type"},
		},
		{
			name:             "no chain needed but has sub-key",
			replacements:     map[string]string{"A": "Concrete", "Con": "other"},
			wantReplacements: map[string]string{"A": "Concrete", "Con": "other"},
		},
		{
			name:         "self cycle should panic",
			replacements: map[string]string{"A": "A"},
			wantPanic:    "recursion detected",
		},
		{
			name:         "cycle A->B->A should panic",
			replacements: map[string]string{"A": "B", "B": "A"},
			wantPanic:    "recursion detected",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			replCopy := make(map[string]string, len(tc.replacements))
			for k, v := range tc.replacements {
				replCopy[k] = v
			}

			if tc.wantPanic != "" {
				defer func() {
					r := recover()
					if r == nil {
						t.Fatalf("expected panic with %q, got nil", tc.wantPanic)
					}

					if msg := r.(string); msg != tc.wantPanic {
						t.Errorf("panicked with %q, want %q", msg, tc.wantPanic)
					}
				}()
			}

			ResolveReplacements(replCopy)

			if len(replCopy) != len(tc.wantReplacements) {
				t.Errorf("ResolveReplacements() = map has %d keys, want %d", len(replCopy), len(tc.wantReplacements))
				return
			}

			for k, wantV := range tc.wantReplacements {
				gotV, ok := replCopy[k]
				if !ok {
					t.Errorf("key %q missing from result", k)
					continue
				}

				if gotV != wantV {
					t.Errorf("ResolveReplacements()[%q] = %q, want %q", k, gotV, wantV)
				}
			}
		})
	}
}
