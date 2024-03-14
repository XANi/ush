package web

import (
	"fmt"
	"github.com/XANi/toolbox/project-templates/go-gin-embedded/store"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (b *WebBackend) Store(c *gin.Context) {
	path := c.Param("name")
	pathS := strings.Split(path, "/")
	if len(pathS) < 1 {
		c.String(http.StatusBadRequest, "path needs filename in it")
		return
	}
	fileInfo := store.FileMeta{
		Filename: pathS[len(pathS)-1],
	}
	// for single file, we will just anonymize file name
	if len(pathS) == 1 {
		path = b.token.GetToken(path, c.Request)
	} else {
		base := b.token.GetToken(
			strings.Join(pathS[0:len(pathS)-1], "/"),
			c.Request)
		path = base + "/" + pathS[len(pathS)-1]
	}

	writer, err := b.store.Store(path, fileInfo)
	defer c.Request.Body.Close()
	if err != nil {
		c.String(http.StatusInternalServerError, "server error")
		b.l.Warnf("could not store %s:%s", path, err)
		return
	}

	// TODO exit if progress is too small
	saved := int64(0)
	for n, err := io.Copy(writer, c.Request.Body); n > 0; {
		b.l.Infof("copied %d[%d]: %s", n, saved, err)
		saved += n
		if err == nil {
			break
		}
		_ = err
	}

	if err != nil {
		b.l.Warnf("error while writing %s:%s", path, err)
	}
	writer.Close()
	reqLen, err := strconv.Atoi(c.Request.Header.Get("content-length"))
	if err == nil {
		if reqLen != int(saved) {
			b.l.Warnf("[%s]content-length: %d, saved: %d", path, reqLen, saved)
			c.String(http.StatusInternalServerError, "file size mismatch")
			return
		}
	}

	c.String(http.StatusOK, fmt.Sprintf(
		"download URL: %s\n",
		b.GetResourceURL(c, path),
	))

}

func (b *WebBackend) GetResourceURL(c *gin.Context, path string, key ...string) string {
	if len(key) > 1 {
		b.l.Panicf("key should be singular")
	}
	proto := "http"
	// FIXME check whether header comes from proxy
	if c.Request.TLS != nil ||
		c.Request.Header.Get("x-forwarded-proto") == "https" {
		proto = "https"
	}
	return fmt.Sprintf("%s://%s/d/%s", proto, c.Request.Host, path)

}

func (b *WebBackend) Get(c *gin.Context) {
	path := c.Param("name")
	read, fm, err := b.store.Read(path)
	_ = fm
	if err != nil {
		c.String(http.StatusNotFound, "server error")
		b.l.Warnf("could not read %s:%s", path, err)
		return
	}
	http.ServeContent(c.Writer, c.Request, path, time.Now(), read)
}
