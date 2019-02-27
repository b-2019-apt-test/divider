package mocks

import (
	"errors"
	"time"
)

// Failer makes bad things happen.
type Failer struct {
	failOn int
	failFn FailFn
	every  int
	c      int
}

// FailFn makes bad things.
type FailFn func() error

// FakeWriter does not write anything.
type FakeWriter struct{}

// BrokenWriter returns ErrBrokenWriter after n-th Write call.
type BrokenWriter struct {
	n, c int
}

var (
	// ErrBrokenWriter raised by BrokenWriter after n bytes written.
	ErrBrokenWriter = errors.New("Writer is broken")

	// ErrPanicWriter is an error PanicWritter panics with.
	ErrPanicWriter = errors.New("Panic should happen")

	// PanicFailFn panics with ErrPanicWriter.
	PanicFailFn = func() error { panic(ErrPanicWriter) }

	// StuckFailFn introduces a delay of 1s.
	StuckFailFn = func() error { <-time.After(time.Second); return nil }
)

// NewFailer creates a new Failer.
func NewFailer() *Failer {
	return new(Failer)
}

// NewErrFailFn creates a new FailFn from error.
func NewErrFailFn(err error) FailFn {
	return func() error {
		return err
	}
}

// FailOn n-th Fail call.
func (f *Failer) FailOn(n int) *Failer {
	f.failOn = n
	return f
}

// FailFn to be called on n-th call.
func (f *Failer) FailFn(fn func() error) *Failer {
	f.failFn = fn
	return f
}

// FailEvery n-th call.
func (f *Failer) FailEvery(n int) *Failer {
	f.every = n
	return f
}

// Fail happen on n-th call.
func (f *Failer) Fail() error {
	if f.every > 0 && f.c%f.every == 0 {
		return f.failFn()
	}
	if f.c == f.failOn && f.failFn != nil {
		return f.failFn()
	}
	f.c++
	return nil
}

// NewFakeWriter creates a new FakeWriter.
func NewFakeWriter() *FakeWriter {
	return new(FakeWriter)
}

func (w *FakeWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// NewBrokenWriter creates a new BrokenWriter.
func NewBrokenWriter(n int) *BrokenWriter {
	return &BrokenWriter{n: n}
}

func (w *BrokenWriter) Write(p []byte) (int, error) {
	if w.c >= w.n {
		return 0, ErrBrokenWriter
	}
	w.c++
	return len(p), nil
}
