package store

import (
	"testing"

	"github.com/spelens-gud/Verktyg/implements/store/memcached"
)

func TestMemcache(t *testing.T) {
	store, _ := memcached.New("localhost:11211")
	s, err := store.Get("xxxx")
	if err != nil {
		t.Fatal(err)
	}
	t.Error(s)

}
