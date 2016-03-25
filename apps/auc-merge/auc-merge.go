package main

import (
	"bufio"
	"log"
	"os"

	"github.com/gourytch/gowowuction/util"
	"golang.org/x/crypto/ssh" //see https://gist.github.com/jedy/3357393
)

/*
type Source struct {
	hostname string `json:"host"`
	username string `json:"user"`
	keyfile  string `json:"priv"`
	pathname string `json:"path"`
}
*/

var clientConfig *ssh.ClientConfig
var hostlist []string
var username string = "leech"
var pathname string = "/home/leech/leech/data/json"

func initialize() {
	privkey_fname := util.AppBaseFileName() + ".privkey"
	privkey_data, err := util.Load(privkey_fname)
	if err != nil {
		log.Panicf("privkey load error: %s", err)
	}
	signer, err := ssh.ParsePrivateKey(privkey_data)
	if err != nil {
		log.Panicf("privkey parse error: %s", err)
	}
	clientConfig = &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	hostlist_fname := util.AppBaseFileName() + ".hostlist"
	f, err := os.Open(hostlist_fname)
	if err != nil {
		log.Panicf("hostlist open error: %s", err)
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		hostlist = append(hostlist, scanner.Text())
	}
}

func sync(host string) {
	client, err := ssh.Dial("tcp", host, clientConfig)
	if err != nil {
		log.Panicf("ssh.Dial(%v) failed: %s", host, err)
	}
	session, err := client.NewSession()
	if err != nil {
		log.Panicf("NewSession() failed: %s", err)
	}
	defer session.Close()
}

func main() {
	log.Print("auc-merge")
	initialize()

}
