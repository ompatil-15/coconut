package iostreams

import (
	"io"
	"os"
)

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

func System() *IOStreams {
	return &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}
