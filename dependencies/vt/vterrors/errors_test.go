/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vterrors

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	vtrpcpb "github.com/redhajuanda/sqlparser/dependencies/vt/proto/vtrpc"
)

func TestWrapNil(t *testing.T) {
	got := Wrap(nil, "no error")
	if got != nil {
		t.Errorf("Wrap(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		err         error
		message     string
		wantMessage string
		wantCode    vtrpcpb.Code
	}{
		{io.EOF, "read error", "read error: EOF", vtrpcpb.Code_UNKNOWN},
		{New(vtrpcpb.Code_ALREADY_EXISTS, "oops"), "client error", "client error: oops", vtrpcpb.Code_ALREADY_EXISTS},
	}

	for _, tt := range tests {
		got := Wrap(tt.err, tt.message)
		if got.Error() != tt.wantMessage {
			t.Errorf("Wrap(%v, %q): got: [%v], want [%v]", tt.err, tt.message, got, tt.wantMessage)
		}
		if Code(got) != tt.wantCode {
			t.Errorf("Wrap(%v, %v): got: [%v], want [%v]", tt.err, tt, Code(got), tt.wantCode)
		}
	}
}

func TestUnwrap(t *testing.T) {
	tests := []struct {
		err       error
		isWrapped bool
	}{
		{fmt.Errorf("some error: %d", 17), false},
		{errors.New("some new error"), false},
		{Errorf(vtrpcpb.Code_INVALID_ARGUMENT, "some msg %d", 19), false},
		{Wrapf(errors.New("some wrapped error"), "some msg"), true},
		{nil, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.err), func(t *testing.T) {
			{
				wasWrapped, unwrapped := Unwrap(tt.err)
				assert.Equal(t, tt.isWrapped, wasWrapped)
				if !wasWrapped {
					assert.Equal(t, tt.err, unwrapped)
				}
			}
			{
				wrapped := Wrap(tt.err, "some message")
				wasWrapped, unwrapped := Unwrap(wrapped)
				assert.Equal(t, wasWrapped, (tt.err != nil))
				assert.Equal(t, tt.err, unwrapped)
			}
		})
	}
}

func TestUnwrapAll(t *testing.T) {
	tests := []struct {
		err error
	}{
		{fmt.Errorf("some error: %d", 17)},
		{errors.New("some new error")},
		{Errorf(vtrpcpb.Code_INVALID_ARGUMENT, "some msg %d", 19)},
		{nil},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.err), func(t *testing.T) {
			{
				// see that unwrapping a non-wrapped error just returns the same error
				unwrapped := UnwrapAll(tt.err)
				assert.Equal(t, tt.err, unwrapped)
			}
			{
				// see that unwrapping a 5-times wrapped error returns the original error
				wrapped := tt.err
				for range rand.Perm(5) {
					wrapped = Wrap(wrapped, "some message")
				}
				unwrapped := UnwrapAll(wrapped)
				assert.Equal(t, tt.err, unwrapped)
			}
		})
	}

}

type nilError struct{}

func (nilError) Error() string { return "nil error" }

func TestRootCause(t *testing.T) {
	x := New(vtrpcpb.Code_FAILED_PRECONDITION, "error")
	tests := []struct {
		err  error
		want error
	}{{
		// nil error is nil
		err:  nil,
		want: nil,
	}, {
		// explicit nil error is nil
		err:  (error)(nil),
		want: nil,
	}, {
		// typed nil is nil
		err:  (*nilError)(nil),
		want: (*nilError)(nil),
	}, {
		// uncaused error is unaffected
		err:  io.EOF,
		want: io.EOF,
	}, {
		// caused error returns cause
		err:  Wrap(io.EOF, "ignored"),
		want: io.EOF,
	}, {
		err:  x, // return from errors.New
		want: x,
	}}

	for i, tt := range tests {
		got := RootCause(tt.err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("test %d: got %#v, want %#v", i+1, got, tt.want)
		}
	}
}

func TestCause(t *testing.T) {
	x := New(vtrpcpb.Code_FAILED_PRECONDITION, "error")
	tests := []struct {
		err  error
		want error
	}{{
		// nil error is nil
		err:  nil,
		want: nil,
	}, {
		// uncaused error is nil
		err:  io.EOF,
		want: nil,
	}, {
		// caused error returns cause
		err:  Wrap(io.EOF, "ignored"),
		want: io.EOF,
	}, {
		err:  x, // return from errors.New
		want: nil,
	}}

	for i, tt := range tests {
		got := Cause(tt.err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("test %d: got %#v, want %#v", i+1, got, tt.want)
		}
	}
}

func TestWrapfNil(t *testing.T) {
	got := Wrapf(nil, "no error")
	if got != nil {
		t.Errorf("Wrapf(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWrapf(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{Wrapf(io.EOF, "read error without format specifiers"), "client error", "client error: read error without format specifiers: EOF"},
		{Wrapf(io.EOF, "read error with %d format specifier", 1), "client error", "client error: read error with 1 format specifier: EOF"},
	}

	for _, tt := range tests {
		got := Wrap(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("Wrapf(%v, %q): got: %v, want %v", tt.err, tt.message, got, tt.want)
		}
	}
}

func TestErrorf(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{Errorf(vtrpcpb.Code_DATA_LOSS, "read error without format specifiers"), "read error without format specifiers"},
		{Errorf(vtrpcpb.Code_DATA_LOSS, "read error with %d format specifier", 1), "read error with 1 format specifier"},
	}

	for _, tt := range tests {
		got := tt.err.Error()
		if got != tt.want {
			t.Errorf("Errorf(%v): got: %q, want %q", tt.err, got, tt.want)
		}
	}
}

func innerMost() error {
	return Wrap(io.ErrNoProgress, "oh noes")
}

func middle() error {
	return innerMost()
}

func outer() error {
	return middle()
}

func TestStackFormat(t *testing.T) {
	err := outer()
	got := fmt.Sprintf("%v", err)

	assertContains(t, got, "innerMost", false)
	assertContains(t, got, "middle", false)
	assertContains(t, got, "outer", false)

	setLogErrStacks(true)
	defer func() { setLogErrStacks(false) }()
	got = fmt.Sprintf("%v", err)
	assertContains(t, got, "innerMost", true)
	assertContains(t, got, "middle", true)
	assertContains(t, got, "outer", true)
}

// errors.New, etc values are not expected to be compared by value
// but the change in errors#27 made them incomparable. Assert that
// various kinds of errors have a functional equality operator, even
// if the result of that equality is always false.
func TestErrorEquality(t *testing.T) {
	vals := []error{
		nil,
		io.EOF,
		errors.New("EOF"),
		New(vtrpcpb.Code_ALREADY_EXISTS, "EOF"),
		Errorf(vtrpcpb.Code_INVALID_ARGUMENT, "EOF"),
		Wrap(io.EOF, "EOF"),
		Wrapf(io.EOF, "EOF%d", 2),
	}

	for i := range vals {
		for j := range vals {
			_ = vals[i] == vals[j] // mustn't panic
		}
	}
}

func TestCreation(t *testing.T) {
	testcases := []struct {
		in, want vtrpcpb.Code
	}{{
		in:   vtrpcpb.Code_CANCELED,
		want: vtrpcpb.Code_CANCELED,
	}, {
		in:   vtrpcpb.Code_UNKNOWN,
		want: vtrpcpb.Code_UNKNOWN,
	}}
	for _, tcase := range testcases {
		if got := Code(New(tcase.in, "")); got != tcase.want {
			t.Errorf("Code(New(%v)): %v, want %v", tcase.in, got, tcase.want)
		}
		if got := Code(Errorf(tcase.in, "")); got != tcase.want {
			t.Errorf("Code(Errorf(%v)): %v, want %v", tcase.in, got, tcase.want)
		}
	}
}

func TestCode(t *testing.T) {
	testcases := []struct {
		in   error
		want vtrpcpb.Code
	}{{
		in:   nil,
		want: vtrpcpb.Code_OK,
	}, {
		in:   errors.New("generic"),
		want: vtrpcpb.Code_UNKNOWN,
	}, {
		in:   New(vtrpcpb.Code_CANCELED, "generic"),
		want: vtrpcpb.Code_CANCELED,
	}, {
		in:   context.Canceled,
		want: vtrpcpb.Code_CANCELED,
	}, {
		in:   context.DeadlineExceeded,
		want: vtrpcpb.Code_DEADLINE_EXCEEDED,
	}}
	for _, tcase := range testcases {
		if got := Code(tcase.in); got != tcase.want {
			t.Errorf("Code(%v): %v, want %v", tcase.in, got, tcase.want)
		}
	}
}

func TestWrapping(t *testing.T) {
	err1 := Errorf(vtrpcpb.Code_UNAVAILABLE, "foo")
	err2 := Wrapf(err1, "bar")
	err3 := Wrapf(err2, "baz")
	errorWithoutStack := fmt.Sprintf("%v", err3)

	setLogErrStacks(true)
	errorWithStack := fmt.Sprintf("%v", err3)
	setLogErrStacks(false)

	assertEquals(t, err3.Error(), "baz: bar: foo")
	assertContains(t, errorWithoutStack, "foo", true)
	assertContains(t, errorWithoutStack, "bar", true)
	assertContains(t, errorWithoutStack, "baz", true)
	assertContains(t, errorWithoutStack, "TestWrapping", false)

	assertContains(t, errorWithStack, "foo", true)
	assertContains(t, errorWithStack, "bar", true)
	assertContains(t, errorWithStack, "baz", true)
	assertContains(t, errorWithStack, "TestWrapping", true)

}

func assertContains(t *testing.T, s, substring string, contains bool) {
	t.Helper()
	if doesContain := strings.Contains(s, substring); doesContain != contains {
		t.Errorf("string `%v` contains `%v`: %v, want %v", s, substring, doesContain, contains)
	}
}

func assertEquals(t *testing.T, a, b any) {
	if a != b {
		t.Fatalf("expected [%s] to be equal to [%s]", a, b)
	}
}
