package main

import (
	//	"bufio"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
	//	"os"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"github.com/gourytch/gowowuction/util"
	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

const (
	USER                 = "user"
	PASS                 = "password"
	LISTEN_HOST_AND_PORT = "127.0.0.1:22022"
)

func passwordCallback(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	log.Printf("login: '%s' with password '%s'", c.User(), string(pass))
	if c.User() == USER && string(pass) == PASS {
		return nil, nil
	}
	return nil, fmt.Errorf("password rejected for user %s", c.User())
}

func fingerprintKey(k ssh.PublicKey) string {
	bytes := md5.Sum(k.Marshal())
	strbytes := make([]string, len(bytes))
	for i, b := range bytes {
		strbytes[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(strbytes, ":")
}

func pubkeyCallback(c ssh.ConnMetadata, pubkey ssh.PublicKey) (*ssh.Permissions, error) {
	log.Printf("login: '%s'. pubkey fingerprint: %s", c.User(), fingerprintKey(pubkey))
	if c.User() == USER {
		return nil, nil
	}
	return nil, fmt.Errorf("pubkey rejected for user %s", c.User())
}

func parseDims(b []byte) (w, h uint32) {
	w = binary.BigEndian.Uint32(b)
	h = binary.BigEndian.Uint32(b[4:])
	return
}

type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16
	y      uint16
}

func SetWinsize(fd uintptr, w, h uint32) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}

func handleChannel(c ssh.NewChannel) {
	if t := c.ChannelType(); t != "session" {
		log.Println("rejected unknown channel type:", t)
		c.Reject(ssh.UnknownChannelType, "unknown channel type")
	}
	connection, requests, err := c.Accept()
	if err != nil {
		log.Println("channel not accepted:", err)
		return
	}
	bash := exec.Command("/bin/bash")
	close := func() {
		connection.Close()
		_, err := bash.Process.Wait()
		if err != nil {
			log.Println("bash not exited:", err)
		}
		log.Println("session closed")
	}
	bashf, err := pty.Start(bash)
	if err != nil {
		log.Println("pty not started:", err)
		close()
		return
	}
	var once sync.Once
	go func() {
		io.Copy(connection, bashf)
		once.Do(close)
	}()
	go func() {
		io.Copy(bashf, connection)
		once.Do(close)
	}()
	go func() {
		for req := range requests {
			log.Println("got request:", req.Type, "want reply:", req.WantReply)
			switch req.Type {
			case "shell":
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
				termLen := req.Payload[3]
				w, h := parseDims(req.Payload[termLen+4:])
				SetWinsize(bashf.Fd(), w, h)
				req.Reply(true, nil)
			case "window-change":
				w, h := parseDims(req.Payload)
				SetWinsize(bashf.Fd(), w, h)
			}
		}
	}()
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for c := range chans {
		go handleChannel(c)
	}
}

func serve() {
	config := &ssh.ServerConfig{
		PasswordCallback:  passwordCallback,
		PublicKeyCallback: pubkeyCallback,
	}
	priv_fname := util.AppBaseFileName() + ".privkey"
	log.Print("loading private key from " + priv_fname + " ...")
	priv_bytes, err := ioutil.ReadFile(priv_fname)
	if err != nil {
		log.Panicln("private key read error:", err)
	}
	log.Print("parsing private key...")
	private, err := ssh.ParsePrivateKey(priv_bytes)
	if err != nil {
		log.Panicln("private key parse error:", err)
	}
	log.Print("key fingerprint:", fingerprintKey(private.PublicKey()))
	log.Print("adding private key to host...")
	config.AddHostKey(private)

	log.Println("creating listener for", LISTEN_HOST_AND_PORT, " ...")
	listener, err := net.Listen("tcp", LISTEN_HOST_AND_PORT)
	log.Println("entering in the main service loop ...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept incoming connection:", err)
			continue
		}
		log.Println("new connection accepted from", conn.RemoteAddr())
		log.Println("upgrading connection to ssh...")
		sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
		if err != nil {
			log.Println("handshake failed:", err)
			continue
		}
		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}

}

func main() {
	serve()
}
