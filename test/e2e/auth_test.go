package e2e_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	chclient "github.com/jpillora/chisel/client"
	chserver "github.com/jpillora/chisel/server"
)

func TestAuthUserPass(t *testing.T) {
	tmpPort := availablePort()
	teardown := simpleSetup(t,
		&chserver.Config{
			AuthJSON: `{"users":[{"name":"foo","password":"bar"}]}`,
		},
		&chclient.Config{
			Remotes: []string{"R:" + tmpPort + ":$FILEPORT"},
			Auth:    "foo:bar",
		})
	defer teardown()

	result, err := post("http://localhost:"+tmpPort, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if result != "foo!" {
		t.Fatalf("expected foo!, got %s", result)
	}
}

func TestAuthUserPassFailed(t *testing.T) {
	layout := testLayout{
		server: &chserver.Config{
			AuthJSON: `{"users":[{"name":"foo","password":"bar"}]}`,
		},
		client: &chclient.Config{
			Remotes:          []string{"R:0:$FILEPORT"},
			Auth:             "foo:wrong-password",
			MaxRetryCount:    0,
			MaxRetryInterval: time.Millisecond,
		},
		fileServer: true,
	}

	_, client, teardown := layout.setup(t)
	defer teardown()

	err := client.Wait()
	if err == nil {
		t.Fatal("expected client to fail with auth error, but it exited cleanly")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Fatalf("expected authentication failed error, got: %v", err)
	}
}

func TestAuthToken(t *testing.T) {
	//chisel server
	serverConfig := &chserver.Config{
		AuthJSON: `{"users":[{"name":"foo","password":""}]}`, //no password needed for token auth
	}
	server, err := chserver.NewServer(serverConfig)
	if err != nil {
		t.Fatal(err)
	}
	server.Debug = debug
	chiselPort := availablePort()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.StartContext(ctx, "127.0.0.1", chiselPort); err != nil {
		t.Fatal(err)
	}
	go server.Wait()

	//http proxy which injects user
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Chisel-User", "foo")
		//proxy to chisel server
		d := net.Dialer{}
		backendConn, err := d.DialContext(r.Context(), "tcp", "127.0.0.1:"+chiselPort)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer backendConn.Close()
		//hijack
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "hijacking not supported", http.StatusInternalServerError)
			return
		}
		clientConn, _, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer clientConn.Close()
		//forward request
		if err := r.Write(backendConn); err != nil {
			return // client disconnected
		}
		//pipe
		errc := make(chan error, 2)
		go func() {
			_, err := io.Copy(clientConn, backendConn)
			errc <- err
		}()
		go func() {
			_, err := io.Copy(backendConn, clientConn)
			errc <- err
		}()
		<-errc
	}))
	defer proxy.Close()

	//fileserver
	filePort := availablePort()
	fileAddr := "127.0.0.1:" + filePort
	f := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			w.Write(append(b, '!'))
		}),
	}
	fl, err := net.Listen("tcp", fileAddr)
	if err != nil {
		t.Fatal(err)
	}
	go f.Serve(fl)
	defer f.Close()

	//chisel client
	clientPort := availablePort()
	clientConfig := &chclient.Config{
		Server:      proxy.URL,
		AuthToken:   "some-token",
		Remotes:     []string{"R:" + clientPort + ":" + filePort},
		Fingerprint: server.GetFingerprint(),
	}
	client, err := chclient.NewClient(clientConfig)
	if err != nil {
		t.Fatal(err)
	}
	client.Debug = debug
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	go client.Wait()

	//wait for client to be ready
	select {
	case <-client.Ready:
		//ready
	case <-time.After(5 * time.Second):
		t.Fatal("client not ready in time")
	}

	//test remote
	result, err := post("http://localhost:"+clientPort, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if result != "foo!" {
		t.Fatalf("expected foo!, got %s", result)
	}
}
