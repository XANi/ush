package web

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"net/http"
	"strings"
	"time"
)

type UserTokenGenerator struct {
	cache *lru.LRU[string, []byte]
}

func NewUserTokenGenerator() *UserTokenGenerator {
	return &UserTokenGenerator{
		cache: lru.NewLRU[string, []byte](1024, nil, time.Hour*10),
	}
}
func (utg *UserTokenGenerator) GetToken(basePath string, r *http.Request) (token string) {
	h := sha256.New()
	ua := r.Header.Get("user-agent")
	h.Write([]byte(basePath))
	if len(ua) > 0 {
		h.Write([]byte(ua))
	}
	xff := r.Header.Get("x-forwarded-for")
	if len(xff) > 0 {
		xffs := strings.Split(xff, ",")
		h.Write([]byte(xffs[0]))
	}

	cacheKey, ok := utg.cache.Get(string(h.Sum([]byte{})))
	if !ok {
		token := make([]byte, 16)
		rand.Read(token)
		utg.cache.Add(string(h.Sum([]byte{})), token)
		cacheKey = token
	}
	return base64.URLEncoding.EncodeToString(h.Sum(cacheKey))[0:14]
}
