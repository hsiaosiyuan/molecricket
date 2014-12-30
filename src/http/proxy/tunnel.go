package proxy

// protocol details http://www.web-cache.com/Writings/Internet-Drafts/draft-luotonen-web-proxy-tunneling-01.txt

import (
	"net"
	"http/auth"
	"bytes"
	"io"
	"log"
)

type Tunnel struct {
	Left           *Left
	Right          *net.TCPConn
}

func (t *Tunnel) prepare() (err error) {
	if err = t.Left.Prepare(); err != nil {
		return err
	}

	if t.Right, err = net.DialTCP("tcp", nil, t.Left.RightAddr); err != nil {
		return err
	}

	return nil;
}

func (t *Tunnel) CloseRight() {
	if t.Right != nil {
		t.Right.Close()
		t.Right = nil
	}
}

func (t *Tunnel) Close407() {
	t.Left.Close407()
	t.CloseRight()
}

func (t *Tunnel) Close502() {
	t.Left.Close502()
	t.CloseRight()
}

func (t *Tunnel) Close() {
	t.Left.Close()
	t.CloseRight()
}

func (t *Tunnel) writeLeftReadBuf() (err error) {
	if _, err = t.Left.ReadBytes.WriteTo(t.Right); err != nil {
		if err != io.EOF {
			return err
		}
	}

	return nil
}

func (t *Tunnel) doConnectHandshake() {
	t.Left.Raw.Write([]byte("HTTP/1.0 200 Connection established\nProxy-agent: " + SERVER_NAME + "\n\n"))
}

func (t *Tunnel) Handle() {
	defer func() {
		if r := recover(); r != nil {
			t.Close()
		}
	}()

	var (
		err error
	)

	if err = t.prepare(); err != nil {
		if err == auth.ERR_INVALID_USERNAME_OR_PASSWORD || err == ERR_MISSING_AUTH_HEADER {
			t.Close407()
		}else {
			t.Close502()
		}

		log.Println(err)
		return
	}

	if t.Left.IsConnect {
		t.doConnectHandshake()
	}else {
		if err = t.writeLeftReadBuf(); err != nil {
			t.Close502()
			return
		}
	}

	// left => right
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Close()
			}
		}()

		var (
			err error
			buf []byte
			i   int
		)

		for {
			buf = make([]byte, bytes.MinRead)

			if i, err = t.Left.BufReader.Read(buf); err != nil {
				if err == io.EOF {
					t.Close()
				}else {
					t.Close502()
				}

				break
			}else {
				if _, err = t.Right.Write(buf[:i]); err != nil {
					t.Close502()
					break
				}
			}
		}
	}()

	// right => left
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Close()
			}
		}()

		var (
			err error
		)

		for {
			if _, err = io.Copy(t.Left.Raw, t.Right); err != nil {
				t.Close502()
			}

			t.Close()
		}
	}()
}

func NewTunnel(conn *net.TCPConn) *Tunnel {
	t := new(Tunnel)
	t.Left = NewLeft(conn)

	return t
}
