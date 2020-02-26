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
	"sync"
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
func makeSshConfig(user string, key ssh.Signer) (*ssh.ClientConfig, error) {

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

type Creds struct {
	sync.Mutex
	user       string
	server     string
	id_rsa     ssh.Signer
	userFile   string
	serverFile string
	id_rsaFile string
}

func NewCreds(serverFile string, userFile string, id_rsaFile string) *Creds {

	return &Creds{serverFile: serverFile, userFile: userFile, id_rsaFile: id_rsaFile}
}

func (c *Creds) ReadCredentials() error {
	c.Lock()
	defer c.Unlock()

	rawServer, err := ioutil.ReadFile(c.serverFile)
	if err != nil {
		log.Printf("Can't read serverFile: %s\n", c.serverFile)
		return err
	}
	rawUser, err := ioutil.ReadFile(c.userFile)
	if err != nil {
		log.Printf("Can't read userFile: %s\n", c.userFile)
		return err
	}

	buff, _ := ioutil.ReadFile(c.id_rsaFile)
	id_rsa, err := ssh.ParsePrivateKey(buff)
	if err != nil {
		log.Printf("Can't read id_rsaFile: %s\n", c.id_rsaFile)
		return err
	}

	c.server = strings.TrimSuffix(string(rawServer), "\n")
	c.user = strings.TrimSuffix(string(rawUser), "\n")
	c.id_rsa = id_rsa
	return nil
}

// localAddr := "0.0.0.0:3000"
// remoteAddr := "127.0.0.1:3000"
func (c *Creds) Connect(localAddr string, remoteAddr string) {
	c.Lock()
	defer c.Unlock()

	// Build SSH client configuration
	cfg, err := makeSshConfig(c.user, c.id_rsa)
	if err != nil {
		log.Printf("makeSshConfig: %s", c.user)
		log.Fatalln(err)
	}

	log.Printf("...start Dial...")
	// Establish connection with SSH rawServer
	conn, err := ssh.Dial("tcp", c.server, cfg)
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

func Proxy() error {
	serverFile := "/credentials/SERVER"
	userFile := "/credentials/USER"
	idRsafile := "/credentials/id_rsa"
	creds := NewCreds(serverFile, userFile, idRsafile)
	err := creds.ReadCredentials()
	if err != nil {
		return err
	}

	localAddr := "0.0.0.0:3000"
	remoteAddr := "127.0.0.1:3000"
	creds.Connect(localAddr, remoteAddr)
	return nil
}
