package pkg

import (
	"fmt"
	"github.com/mchirico/sshproxy-action/sshDocker/ssh"
	"os"
	"path/filepath"
	"time"
)

func Walk(root string) []string {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file)
	}

	return files

}

func Speak(args []string) string {

	WriteEnv()
	current := time.Now()
	return fmt.Sprintf("::set-output name=time::%v %s\n", args, current)
}

func WriteEnv() {
	Walk("/credentials")
	ssh.RunME()

}
