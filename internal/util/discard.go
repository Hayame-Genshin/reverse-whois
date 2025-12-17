package util

import "io"

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func NewDiscardWriter() io.Writer { return discardWriter{} }
