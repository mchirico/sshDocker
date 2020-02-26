package proxy

import (
	"testing"
)

func TestSetFiles(t *testing.T) {
	serverFile := "../.credentials/SERVER"
	userFile := "../.credentials/USER"
	idRsafile := "../.credentials/id_rsa"
	creds := NewCreds(serverFile, userFile, idRsafile)
	err := creds.ReadCredentials()
	if err != nil {
		t.FailNow()
	}

	_, err = makeSshConfig(creds.user, creds.id_rsa)
	if err != nil {
		t.Fatalf("makeSshConfig: %s", creds.user)
	}

}
