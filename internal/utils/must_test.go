package utils

import "testing"

func TestMust_ReturnsValue(t *testing.T) {
	t.Parallel()

	got := Must("hello", nil)
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestMust_PanicsOnError(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on error")
		}
	}()

	Must("value", testErr("something failed"))
}

func TestNoErr_DoesNothing(t *testing.T) {
	t.Parallel()

	NoErr(nil) // should not panic
}

func TestNoErr_PanicsOnError(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on error")
		}
	}()

	NoErr(testErr("operation failed"))
}

type testErr string

func (e testErr) Error() string { return string(e) }
