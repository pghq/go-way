package maxmind

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDB_Get(t *testing.T) {
	t.Run("bad open", func(t *testing.T) {
		_, err := Open(context.TODO(), "does-not-exist")
		assert.NotNil(t, err)
	})

	t.Run("open timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 0)
		defer cancel()

		_, err := Open(ctx, "does-not-exist")
		assert.NotNil(t, err)
	})

	t.Run("bad gzip", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("bad body"))
		}))

		_, err := Open(context.TODO(), s.URL)
		assert.NotNil(t, err)
	})

	t.Run("bad tar", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var b bytes.Buffer
			gz := gzip.NewWriter(&b)
			gz.Write([]byte("bad body"))
			gz.Close()
			w.Write(b.Bytes())
		}))

		_, err := Open(context.TODO(), s.URL)
		assert.NotNil(t, err)
	})

	t.Run("bad mmdb", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var b bytes.Buffer
			gz := gzip.NewWriter(&b)
			tw := tar.NewWriter(gz)
			tw.WriteHeader(&tar.Header{
				Name: "bad.mmdb",
				Mode: 0600,
				Size: int64(len("bad body")),
			})
			tw.Write([]byte("bad body"))
			tw.Close()
			gz.Close()
			w.Write(b.Bytes())
		}))

		_, err := Open(context.TODO(), s.URL)
		assert.NotNil(t, err)
	})

	s := serve("../testdata/maxmind-test.tgz")
	db, _ := Open(context.TODO(), s.URL)

	t.Run("closed db", func(t *testing.T) {
		db, _ := Open(context.TODO(), s.URL)
		db.Close()

		_, err := db.Get(net.ParseIP("81.2.69.142"))
		assert.NotNil(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := db.Get(net.ParseIP("192.168.1.1"))
		assert.NotNil(t, err)

		db.cache.Wait()
		_, err = db.Get(net.ParseIP("192.168.1.1"))
		assert.NotNil(t, err)
	})

	t.Run("found", func(t *testing.T) {
		city, err := db.Get(net.ParseIP("81.2.69.142"))
		assert.Nil(t, err)
		assert.NotNil(t, city)

		db.cache.Wait()
		city, err = db.Get(net.ParseIP("81.2.69.142"))
		assert.Nil(t, err)
		assert.NotNil(t, city)
	})
}

func serve(path string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}))
}