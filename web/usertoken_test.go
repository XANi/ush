package web

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewUserTokenGenerator(t *testing.T) {
	utg := NewUserTokenGenerator()
	tstHdr1 := &http.Request{
		Header: map[string][]string{
			"User-Agent": {"cake"},
		},
	}
	tstHdr2 := &http.Request{
		Header: map[string][]string{
			"User-Agent": {"cake2"},
		},
	}
	token1aa := utg.GetToken("/base/file1", tstHdr1)
	token1ab := utg.GetToken("/base/file1", tstHdr1)
	token1b := utg.GetToken("/base2/file2", tstHdr1)
	token2aa := utg.GetToken("/base/file1", tstHdr2)
	token2ab := utg.GetToken("/base/file1", tstHdr2)
	token2b := utg.GetToken("/base2/file2", tstHdr2)
	assert.Equal(t, token1aa, token1ab)
	assert.NotEqual(t, token1aa, token1b)
	assert.Equal(t, token2aa, token2ab)
	assert.NotEqual(t, token2aa, token2b)
	assert.NotEqual(t, token1aa, token2aa)
}
