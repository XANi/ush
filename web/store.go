package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (b *WebBackend) Store(c *gin.Context) {
	path := c.Param("name")
	writer, err := b.store.Store(path)
	defer c.Request.Body.Close()
	if err != nil {
		c.String(http.StatusInternalServerError, "server error")
		b.l.Warnf("could not store %s:%s", path, err)
		return
	}
	n, err := io.Copy(writer, c.Request.Body)
	writer.Close()
	reqLen, err := strconv.Atoi(c.Request.Header.Get("content-length"))
	if err == nil {
		if reqLen != int(n) {
			b.l.Warnf("[%s]content-length: %d, saved: %d", path, reqLen, n)
			c.String(http.StatusInternalServerError, "file size mismatch")
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
	read, err := b.store.Read(path)
	if err != nil {
		c.String(http.StatusNotFound, "server error")
		b.l.Warnf("could not read %s:%s", path, err)
		return
	}
	http.ServeContent(c.Writer, c.Request, path, time.Now(), read)
}
