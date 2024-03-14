package file

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/XANi/toolbox/project-templates/go-gin-embedded/store"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
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
	var err error
	if _, err := os.Stat(c.RootDir + "/data"); os.IsNotExist(err) {
		err = os.Mkdir(c.RootDir+"/data", 0o700)
	}
	if err != nil {
		return nil, err
	}
	return &File{root: c.RootDir + "/data"}, nil
}

func (f *File) Store(path string, fileMeta store.FileMeta) (io.WriteCloser, error) {
	sanitizedName := sanitize(path)
	if len(sanitizedName) > 250 {
		return nil, fmt.Errorf("too long name")
	}
	f_path := f.root + "/" + sanitizedName
	ss := strings.Split(sanitizedName, "/")
	if len(ss) > 1 {
		f_dir := f.root + "/" + strings.Join(ss[0:len(ss)-1], "/")
		if _, err := os.Stat(f_dir); os.IsNotExist(err) {
			err := os.MkdirAll(f_dir, 0o700)
			if err != nil {
				return nil, fmt.Errorf("could not make dir %w", err)
			}
		}
	}
	f_path_tmp := f_path + ".tmp" + strconv.Itoa(rand.Int()) // hide the tmpfile during upload
	file, err := os.OpenFile(f_path_tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("could not open file for write: %w", err)
	}
	fmFd, err := os.OpenFile(f_path+".meta", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("could not open filemeta for write: %w", err)
	}
	enc := json.NewEncoder(fmFd)
	err = enc.Encode(&fileMeta)
	fmFd.Sync()
	fmFd.Close()
	if err != nil {
		file.Close()
		return nil, err
	}
	return &tmpWriteCloser{
		tmpFilename: f_path_tmp,
		dstFilename: f_path + ".data",
		i:           file,
	}, err
}

func (f *File) Read(path string) (fd io.ReadSeekCloser, fileMeta store.FileMeta, err error) {
	sanitizedName := sanitize(path)
	f_path := f.root + "/" + sanitizedName
	file, err := os.OpenFile(f_path+".data", os.O_RDONLY, 0o500)
	if err != nil {
		return nil, fileMeta, fmt.Errorf("could not find file %s:%s", f_path, err)
	}
	filemeta_raw, err := os.OpenFile(f_path+".meta", os.O_RDONLY, 0o500)
	if err != nil {
		return file, fileMeta, fmt.Errorf("could not open metadata: %s", err)
	}
	fm_j := json.NewDecoder(filemeta_raw)
	err = fm_j.Decode(&fileMeta)
	if err != nil {
		return file, fileMeta, fmt.Errorf("could not decode metadata: %s", err)
	}
	filemeta_raw.Close()
	return file, fileMeta, err
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
