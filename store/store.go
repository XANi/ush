package store

import "io"

type Store interface {
	Store(path string, meta FileMeta) (io.WriteCloser, error)
	Read(path string) (io.ReadSeekCloser, FileMeta, error)
}
type FileMeta struct {
	Filename string `json:"filename"`
}
