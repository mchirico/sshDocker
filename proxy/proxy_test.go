package proxy

import "testing"

func TestSpeak(t *testing.T) {
	serverFile := "../.credentials/SERVER"
	userFile := "../.credentials/USER"
	idRsafile := "../.credentials/id_rsa"
	ReadCredentials(serverFile, userFile, idRsafile)
}
