package slow

import (
	"github.com/XANi/toolbox/project-templates/go-gin-embedded/store"
	"io"
	"time"
)

type Slow struct {
	store store.Store
}

func New(backendStore store.Store) *Slow {
	return &Slow{store: backendStore}
}

func (f *Slow) Store(path string) (io.WriteCloser, error) {
	rc, err := f.store.Store(path)
	return NewSlowWriteCloser(rc), err
}
func (f *Slow) Read(path string) (io.ReadSeekCloser, error) {
	rc, err := f.store.Read(path)
	return NewSlowReadSeekCloser(rc), err
}

type SlowReadSeekCloser struct {
	i     io.ReadSeekCloser
	delay time.Duration
}

func NewSlowReadSeekCloser(i io.ReadSeekCloser) io.ReadSeekCloser {
	srsc := SlowReadSeekCloser{
		i:     i,
		delay: time.Millisecond * 100,
	}
	return &srsc

}

func (s SlowReadSeekCloser) Read(p []byte) (n int, err error) {
	time.Sleep(s.delay)
	time.Sleep(s.delay * time.Duration(len(p)/10240))
	return s.i.Read(p)
}

func (s SlowReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	return s.i.Seek(offset, whence)
}

func (s SlowReadSeekCloser) Close() error {
	time.Sleep(s.delay)
	return s.i.Close()
}
func NewSlowWriteCloser(i io.WriteCloser) io.WriteCloser {
	srsc := SlowWriteCloser{
		i:     i,
		delay: time.Millisecond * 100,
	}
	return &srsc

}

type SlowWriteCloser struct {
	i     io.WriteCloser
	delay time.Duration
}

func (s SlowWriteCloser) Write(p []byte) (n int, err error) {
	time.Sleep(s.delay)
	time.Sleep(s.delay * time.Duration(len(p)/10240))
	return s.i.Write(p)
}

func (s SlowWriteCloser) Close() error {
	time.Sleep(s.delay)
	return s.i.Close()
}
