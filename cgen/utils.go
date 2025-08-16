package cgen

import (
	"io"
	"strings"
)

type indentedWriter struct {
	w       io.Writer
	level   int
	indent  string
	newline bool
}

func newIndentedWriter(w io.Writer, indent string) *indentedWriter {
	return &indentedWriter{w: w, indent: indent, newline: true}
}

func (iw *indentedWriter) WriteLine(s string) error {
	if iw.newline {
		if _, err := io.WriteString(iw.w, strings.Repeat(iw.indent, iw.level)); err != nil {
			return err
		}
	}
	iw.newline = strings.HasSuffix(s, "\n")
	_, err := io.WriteString(iw.w, s)
	return err
}

func (iw *indentedWriter) Indent(f func() error) error {
	iw.level++
	err := f()
	iw.level--
	return err
}
