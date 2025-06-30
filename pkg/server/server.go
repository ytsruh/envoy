package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

func StartServer() {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "test" && string(pass) == "password" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateBytes, err := generatePrivateKey()
	if err != nil {
		log.Fatal("Failed to generate private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", "localhost:2222")
	if err != nil {
		log.Fatal("failed to listen for connection", err)
	}
	log.Printf("Listening on %s", listener.Addr())

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept incoming connection: %v", err)
			continue
		}

		_, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			log.Printf("failed to handshake: %v", err)
			continue
		}
		log.Printf("New SSH connection from %s", nConn.RemoteAddr())

		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		return
	}
	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("could not accept channel: %v", err)
		return
	}

	go func(in <-chan *ssh.Request) {
		for req := range in {
			switch req.Type {
			case "shell":
				// We only accept shell requests with no payload
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
				// We accept pty requests
				req.Reply(true, nil)
			case "exec":
				// We don't support exec, so we reject it
				req.Reply(false, nil)
			}
		}
	}(requests)

	go func() {
		defer channel.Close()
		// Simple echo server
		buf := make([]byte, 4096)
		for {
			n, err := channel.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("error reading from channel: %v", err)
				}
				return
			}
			_, err = channel.Write(buf[:n])
			if err != nil {
				log.Printf("error writing to channel: %v", err)
				return
			}
		}
	}()
}

func generatePrivateKey() ([]byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)
	privateKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)

	return privateKeyPem, nil
}
