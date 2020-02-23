package ssh

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getHostKey(host string) (ssh.PublicKey, error) {
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error parsing %q: %v", fields[2], err))
			}
			break
		}
	}

	if hostKey == nil {
		return nil, errors.New(fmt.Sprintf("no hostkey for %s", host))
	}
	return hostKey, nil
}

func exec(user string, server string, command string, results chan string) {

	// If you want to valid keys ...
	// hostKey, err := getHostKey("smtp.aipiggybot.io")
	//if err != nil {
	//	log.Fatal(err)
	//}

	var hostKey ssh.PublicKey

	key, err := ioutil.ReadFile("/credentials/id_rsa")

	if err != nil {
		log.Fatalf("Unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
			// ssh.Password("yourpassword"),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	// Ignore key validation
	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	log.Printf("server: %v\n",server)
	client, err := ssh.Dial("tcp", server, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}

	defer client.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	results <- b.String()
}

// Simple Example
func RunME() {

	results := make(chan string, 0)

	user, err := ioutil.ReadFile("/credentials/USER")
	if err != nil {
		log.Fatalf("can't read USER")
	}
	server, err :=  ioutil.ReadFile("/credentials/SERVER")
	if err != nil {
		log.Fatalf("can't read SERVER")
	}

	sserver := strings.TrimSuffix(string(server), "\n")
	suser := strings.TrimSuffix(string(user), "\n")
	log.Printf("server: ->%s<-\n",string(server))
	command := "uptime >out.out;sleep 10"

	go exec(suser, sserver, command, results)
	go exec(suser, sserver, "date", results)

	fmt.Println(<-results)
	fmt.Println(<-results)

	close(results)
}
