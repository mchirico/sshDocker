package proxy

// Inital code by https://github.com/sosedoff
// Modified by https://github.com/mchirico

import (
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

// Get default location of a private key
func privateKeyPath() string {
	//return os.Getenv("HOME") + "/.ssh/id_rsa"
	//return os.Getenv("HOME") + "/.ssh/google_compute_engine"
	return "/credentials/id_rsa"
}

// Get private key for ssh authentication
func parsePrivateKey(keyPath string) (ssh.Signer, error) {
	buff, _ := ioutil.ReadFile(keyPath)
	return ssh.ParsePrivateKey(buff)
}

// Get ssh client config for our connection
// SSH config will use 2 authentication strategies: by key and by password
func makeSshConfig(user string) (*ssh.ClientConfig, error) {
	key, err := parsePrivateKey(privateKeyPath())
	if err != nil {
		return nil, err
	}

	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}
	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	return &config, nil
}

// Handle local client connections and tunnel data to the remote serverq
// Will use io.Copy - http://golang.org/pkg/io/#Copy
func handleClient(client net.Conn, conn *ssh.Client, remoteAddr string) {

	// Establish connection with remote server
	remote, err := conn.Dial("tcp", remoteAddr)
	if err != nil {
		log.Fatalln(err)
	}

	defer client.Close()
	defer remote.Close()

	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Println("error while copy remote->local:", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Transfer done\n")
		chDone <- true
	}()

	<-chDone
}

func Server(conn *ssh.Client, remoteAddr string, localAddr string) {

	// Start local server to forward traffic to remote connection
	local, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Printf("listen: %v\n", err)
	}

	for {
		client, err := local.Accept()
		if err != nil {
			log.Printf("...Server stopped")
			log.Fatalln(err)
		}
		go handleClient(client, conn, remoteAddr)
	}
}

func Proxy() {
	// Connection settings



	server, err := ioutil.ReadFile("/credentials/SERVER")
	if err != nil {
		log.Fatalf("can't read SERVER")
	}
	user, err := ioutil.ReadFile("/credentials/USER")
	if err != nil {
		log.Fatalf("can't read USER")
	}

	sshAddr := strings.TrimSuffix(string(server), "\n")
	suser := strings.TrimSuffix(string(user), "\n")

	localAddr := "0.0.0.0:3000"
	remoteAddr := "127.0.0.1:3000"

	// Build SSH client configuration
	cfg, err := makeSshConfig(suser)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("...start Dial...")
	// Establish connection with SSH server
	conn, err := ssh.Dial("tcp", sshAddr, cfg)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	// Handle incoming connection
	Server(conn, remoteAddr, localAddr)
	log.Printf("connection...")
	// Multiple
	//go Server(conn, remoteAddr, localAddr)
	//Server(conn, "127.0.0.1:9090", "127.0.0.1:9090")

}
