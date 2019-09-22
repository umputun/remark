package migrator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlMapper_URL(t *testing.T) {
	// want remap urls from https://radio-t.com to https://www.radio-t.com
	// also map individual urls
	rules := strings.NewReader(`
https://radio-t.com* https://www.radio-t.com*
https://radio-t.com/p/2018/09/22////podcast-616/ https://www.radio-t.com/p/2018/09/22/podcast-616/
https://radio-t.com/p/2018/09/22/podcast-616/?with_query=1 https://www.radio-t.com/p/2018/09/22/podcast-616/
`)

	mapper, err := NewUrlMapper(rules)
	assert.NoError(t, err)

	// if url not matched mapper should return given url
	assert.Equal(t, "https://any.com/post/1/", mapper.URL("https://any.com/post/1/"))
	assert.Equal(t, "https://radio-t.co", mapper.URL("https://radio-t.co"))

	// check strict matching
	assert.Equal(t, "https://www.radio-t.com/p/2018/09/22/podcast-616/", mapper.URL("https://radio-t.com/p/2018/09/22////podcast-616/"))
	assert.Equal(t, "https://www.radio-t.com/p/2018/09/22/podcast-616/", mapper.URL("https://radio-t.com/p/2018/09/22/podcast-616/?with_query=1"))

	// check pattern matching (by prefix)
	assert.Equal(t, "https://www.radio-t.com/p/post/123/", mapper.URL("https://radio-t.com/p/post/123/"))
}

func TestUrlMapper_New(t *testing.T) {
	casesError := []struct {
		rules       string
		expectError bool
	}{
		// bad input, expect error
		{
			rules:       "https://radio-t.com ",
			expectError: true,
		},
		{
			rules:       "https://radio-t.com https://radio-t.com https://radio-t.com",
			expectError: true,
		},
		{
			rules:       "https://radio-t.com https://radio-t.com\n https://radio-t.com",
			expectError: true,
		},
		{
			rules:       "https://radio-t.com   \n https://radio-t.com https://radio-t.com",
			expectError: true,
		},

		// valid input, no error
		{
			rules: "https://radio-t.com* https://www.radio-t.com*",
		},
		{
			rules: "https://radio-t.com/p/2018/09/22/podcast-616/?with_query=1 https://www.radio-t.com/p/2018/09/22/podcast-616/",
		},
		{
			rules: "https://any.com/p/111 https://any.com/p/222   \n https://any.com/p/333 https://any.com/p/222   \n",
		},
	}
	for _, c := range casesError {
		_, err := NewUrlMapper(strings.NewReader(c.rules))
		if c.expectError {
			assert.Error(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}
