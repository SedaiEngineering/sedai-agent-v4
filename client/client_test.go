package chclient

import (
	"crypto/elliptic"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/jpillora/chisel/share/ccrypto"
	"golang.org/x/crypto/ssh"
)

func TestCustomHeaders(t *testing.T) {
	//fake server
	wg := sync.WaitGroup{}
	wg.Add(1)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Foo") != "Bar" {
			t.Fatal("expected header Foo to be 'Bar'")
		}
		wg.Done()
	}))
	defer server.Close()
	//client
	headers := http.Header{}
	headers.Set("Foo", "Bar")
	config := Config{
		KeepAlive:        time.Second,
		MaxRetryInterval: time.Second,
		Server:           server.URL,
		Remotes:          []string{"R:9000"},
		Headers:          headers,
	}
	c, err := NewClient(&config)
	if err != nil {
		log.Fatal(err)
	}
	go c.Run()
	//wait for test to complete
	wg.Wait()
	c.Close()
}


func TestVerifyFingerprint(t *testing.T) {
	config := Config{
		Fingerprint: "qmrRoo8MIqePv3jC8+wv49gU6uaFgD3FASQx9V8KdmY=",
	}
	c, err := NewClient(&config)
	if err != nil {
		t.Fatal(err)
	}
	//this is the key which has the fingerprint above
	pemBytes := []byte(`
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHi0F0A0E2TzSg3bC5xI0i1aAiqtB3Kj5s70fE0m97YgoAoGCCqGSM49
AwEHoUQDQgAEJmUpZk19a32c2uT+hJ17v0379Kbx46F4rX+5n/4X+HqZ+y1Yc21S
iG94qA6y+Nn+1Z9a9d7Sj0jC6iFq/Q/v7A==
-----END EC PRIVATE KEY-----
`)
	priv, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	err = c.verifyServer("", nil, priv.PublicKey())
	if err != nil {
		t.Fatal(err)
	}
}
