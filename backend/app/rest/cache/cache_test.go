package cache

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umputun/remark/backend/app/rest"
	"github.com/umputun/remark/backend/app/store"
)

func TestCache_Keys(t *testing.T) {
	tbl := []struct {
		key    string
		scopes []string
		full   string
	}{
		{"key1", []string{"s1"}, "s1@@key1@@site"},
		{"key2", []string{"s11", "s2"}, "s11$$s2@@key2@@site"},
		{"key3", []string{}, "@@key3@@site"},
	}

	for n, tt := range tbl {
		k := NewKey("site").ID(tt.key).Scopes(tt.scopes...)
		full := k.Merge()
		assert.Equal(t, tt.full, full, "making key, #%d", n)

		k, e := ParseKey(full)
		assert.Nil(t, e)
		assert.Equal(t, tt.scopes, k.scopes)
		assert.Equal(t, tt.key, k.id)
	}

	_, err := ParseKey("abc")
	assert.Error(t, err)
	_, err = ParseKey("")
	assert.Error(t, err)
}

func TestCache_URLKey(t *testing.T) {
	r, err := http.NewRequest("GET", "http://blah/123", nil)
	assert.Nil(t, err)
	key := URLKey(r)
	assert.Equal(t, "http://blah/123", key)

	r, err = http.NewRequest("GET", "http://blah/123?key=v&k2=v2", nil)
	assert.Nil(t, err)
	key = URLKey(r)
	assert.Equal(t, "http://blah/123?key=v&k2=v2", key)

	user := store.User{Admin: true}
	r = rest.SetUserInfo(r, user)
	key = URLKey(r)
	assert.Equal(t, "admin!!http://blah/123?key=v&k2=v2", key)
}
