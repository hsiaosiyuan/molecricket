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
	var (
		err    error
		cfg    *Config
		addr   *net.TCPAddr
		tl     *net.TCPListener
		conn   *net.TCPConn
	)

	if cfg, err = NewConfig(); err != nil {
		log.Fatal(err)
	}

	auth.SetUsers(cfg.Users)
	auth.SetResources(cfg.Resources)

	addr, _ = net.ResolveTCPAddr("tcp", ":"+cfg.Port)
	tl, err = net.ListenTCP("tcp", addr)

	if err != nil {
		log.Println(err)
	}

	for {
		conn, err = tl.AcceptTCP()
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
