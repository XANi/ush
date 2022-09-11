package file

import (
	"go.uber.org/zap"
	"io"
	"os"
)

type Config struct {
	RootDir string
	Logger  *zap.SugaredLogger
}

type File struct {
	root string
}

func New(c Config) (*File, error) {
	err := os.Mkdir(c.RootDir+"/data", 0o700)
	if err != nil {
		return nil, err
	}
	return &File{root: c.RootDir + "/data"}, nil
}

func (f *File) Store(path string) (io.WriteCloser, error) {
	sanitizedName := sanitize(path)
	f_path := f.root + "/" + sanitizedName
	file, err := os.OpenFile(f_path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (f *File) Read(path string) (io.ReadSeekCloser, error) {
	sanitizedName := sanitize(path)
	f_path := f.root + "/" + sanitizedName
	file, err := os.OpenFile(f_path, os.O_RDONLY, 0o500)
	if err != nil {
		return nil, err
	}
	return file, err
}

func sanitize(s string) string {
	return s
}
