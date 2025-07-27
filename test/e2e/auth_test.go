package e2e_test

import (
	"testing"

	chclient "github.com/jpillora/chisel/client"
	chserver "github.com/jpillora/chisel/server"
)

//TODO tests for:
// - failed auth
// - dynamic auth (server add/remove user)
// - watch auth file

func TestAuth(t *testing.T) {
	tmpPort1 := availablePort()
	tmpPort2 := availablePort()
	//setup server, client, fileserver
	teardown := simpleSetup(t,
		&chserver.Config{
			KeySeed:  "foobar",
			AuthJSON: `{"users":[{"name":"foo","password":"bar"}]}`,
		},
		&chclient.Config{
			Remotes: []string{
				"R:" + tmpPort1 + ":$FILEPORT",
				"R:" + tmpPort2 + ":$FILEPORT",
			},
			Auth: "foo:bar",
		})
	defer teardown()
	//test first remote
	result, err := post("http://localhost:"+tmpPort1, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if result != "foo!" {
		t.Fatalf("expected exclamation mark added")
	}
	//test second remote
	result, err = post("http://localhost:"+tmpPort2, "bar")
	if err != nil {
		t.Fatal(err)
	}
	if result != "bar!" {
		t.Fatalf("expected exclamation mark added again")
	}
}
