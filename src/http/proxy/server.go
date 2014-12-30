package proxy

import (
	"net"
	"log"
	"http/auth"
)

type Server struct {
	ConnCount int64
}

const (
	SERVER_NAME = "molecricket/0.0.1"
)

func Serve() {
	users := []auth.User{
		auth.User{
			"u",
			"p",
			[]string{"molecricket"},
		},
	}

	resources := []auth.Resource{
		auth.Resource{
			"/",
			"molecricket",
		},
	}

	auth.SetUsers(users)
	auth.SetResources(resources)

	la, _ := net.ResolveTCPAddr("tcp", ":9090")
	ln, err := net.ListenTCP("tcp", la)

	if err != nil {
		log.Println(err)
	}
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}

		go func() {
			t := NewTunnel(conn)
			t.Handle()
		}()
	}
}
