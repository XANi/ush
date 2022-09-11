package file

import (
	crand "crypto/rand"
	"encoding/binary"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func init() {
	b := make([]byte, 8)
	crand.Read(b)
	i := int(binary.BigEndian.Uint64(b))
	r.Seed(int64(time.Now().Nanosecond() + i))
}

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
	f_path_tmp := f_path + ".tmp" + strconv.Itoa(rand.Int()) // hide the tmpfile during upload
	file, err := os.OpenFile(f_path_tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, err
	}
	return &tmpWriteCloser{
		tmpFilename: f_path_tmp,
		dstFilename: f_path,
		i:           file,
	}, nil
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

type tmpWriteCloser struct {
	tmpFilename string
	dstFilename string
	i           io.WriteCloser
}

func (t tmpWriteCloser) Write(p []byte) (n int, err error) {
	return t.i.Write(p)
}

func (t tmpWriteCloser) Close() error {
	err := t.i.Close()
	if err != nil {
		return err
	}
	return os.Rename(t.tmpFilename, t.dstFilename)

}
