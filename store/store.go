package store

import "io"

type Store interface {
	Store(path string) (io.WriteCloser, error)
	Read(path string) (io.ReadSeekCloser, error)
}
