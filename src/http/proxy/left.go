package proxy

import (
	"net"
	"bytes"
	"bufio"
	"errors"
	"fmt"
	"http/auth"
	"net/url"
	"strings"
)

var (
	CR                    = byte('\r')
	LF                    = byte('\n')
	COLON                 = byte(':')
	SPACE                 = byte(' ')
	MAX_HEADERS_COUNT     = 100
	KEY_HEADER_AUTH       = []byte("Proxy-Authorization")
	KEY_HEADER_KEEP_ALIVE = []byte("Proxy-Connection")
	METHOD_CONNECT        = []byte("CONNECT")

	ERR_EMPTY_FIRST_LINE        = errors.New("empty first line")
	ERR_MAX_HEADER_COUNT        = errors.New(fmt.Sprintf("max headers count: %d", MAX_HEADERS_COUNT))
	ERR_INVALID_HEADER          = errors.New("invalid header")
	ERR_MISSING_AUTH_HEADER     = errors.New("missing auth header")
	ERR_MISSING_PASSWORD        = errors.New("missing auth password")
	ERR_INVALID_PASSWORD_FORMAT = errors.New("invalid password format")
	ERR_INVALID_FIRST_LINE      = errors.New("invalid first line")
	ERR_INVALID_RIGHT_ADDRESS   = errors.New("invalid right address")
)

type Left struct {
	Raw              *net.TCPConn
	BufReader        *bufio.Reader
	ReadBytes        *bytes.Buffer
	Method           []byte
	Headers          map[string]string
	IsConnect        bool
	RightAddr        *net.TCPAddr
	IsCrLf           bool
}

func NewLeft(c *net.TCPConn) *Left {
	l := new(Left)
	l.Raw = c;
	l.BufReader = bufio.NewReader(c);
	l.ReadBytes = bytes.NewBuffer([]byte{})

	return l
}

func (l *Left) auth() (readBuf *bytes.Buffer, err error) {
	var (
		line      []byte
		n         int
		k         []byte
		v         []byte
		j         int
		h         int
		idx       int
		ak        []byte // auth key
		av        []byte // auth value
		buf       *bytes.Buffer
	)

	buf = bytes.NewBuffer([]byte{})

	// parse headers to find Proxy-Authorization
	n = 0
	for {
		if n > MAX_HEADERS_COUNT {
			return nil, ERR_MAX_HEADER_COUNT
		}

		if line, err = l.BufReader.ReadBytes(LF); err != nil {
			return nil, err
		}

		j = len(line)
		h = len(ak);

		if l.IsCrLf {
			if j == 2 && h == 0 {
				return nil, ERR_MISSING_AUTH_HEADER
			}

			if j == 2 && h != 0 {
				buf.Write(line)
				break
			}
		}else {
			if j == 1 && h == 0 {
				return nil, ERR_MISSING_AUTH_HEADER
			}

			if j == 1 && h != 0 {
				buf.Write(line)
				break
			}
		}

		idx = bytes.Index(line, []byte{COLON});
		if idx == -1 {
			return nil, ERR_INVALID_HEADER
		}

		k = line[0:idx]
		v = line[idx+1:]
		j = len(v)

		// 4 = 1space + 1value + 2(CR+LF)
		// 3 = 1space + 1value + 1(LF)
		if l.IsCrLf && j < 4 || !l.IsCrLf && j < 3 {
			return nil, ERR_INVALID_HEADER
		}

		if l.IsCrLf {
			v = v[1:j-2]  // skip first space
		}else {
			v = v[1:j-1]
		}

		if bytes.Equal(k, KEY_HEADER_AUTH) {
			ak = k
			av = v
			// don't break here since https need to discard the first empty line
		}else if !bytes.Equal(k, KEY_HEADER_KEEP_ALIVE) {
			buf.Write(line)
		}

		n++
	}

	if err = auth.Basic("/", string(av)); err != nil {
		return nil, err
	}

	return buf, nil
}

func (l *Left) Prepare() (err error) {
	var (
		line  []byte
		ps    [][]byte
		n     int
		ra    string
		u     *url.URL
		buf   *bytes.Buffer
	)

	if line, err = l.BufReader.ReadBytes(LF); err != nil {
		return err
	}

	n = len(line)
	if n == 0 {
		return ERR_EMPTY_FIRST_LINE
	}

	if n > 1 && line[n-2] == '\r' {
		l.IsCrLf = true
	}

	// parse first line
	if ps = bytes.Split(line, []byte{SPACE}); len(ps) != 3 {
		return ERR_INVALID_FIRST_LINE
	}

	l.Method = ps[0]
	if bytes.Equal(l.Method, METHOD_CONNECT) {
		l.IsConnect = true
	}

	if l.IsConnect {
		// https just discard the read buffer returned by auth()
		if _, err = l.auth(); err != nil {
			return err
		}

		ra = string(ps[1])
	}else {
		l.ReadBytes.Write(line)

		if buf, err = l.auth(); err != nil {
			return err
		}

		// http need to hold the read buffer returned by auth()
		buf.WriteTo(l.ReadBytes)

		if u, err = url.Parse(string(ps[1])); err != nil {
			return ERR_INVALID_RIGHT_ADDRESS
		}

		if strings.Index(u.Host, ":") == -1 {
			ra = u.Host+":80"
		}else {
			ra = u.Host
		}
	}

	if l.RightAddr, err = net.ResolveTCPAddr("tcp", ra); err != nil {
		return err
	}

	return nil
}

func (l *Left) Close() {
	if l.Raw != nil {
		l.Raw.Close()
		l.Raw = nil
	}
}

func (l *Left) Close502() {
	if l.Raw != nil {
		l.Raw.Write([]byte("HTTP/1.1 502 Bad Gateway\nServer: " + SERVER_NAME + "\n"))

		l.Raw.Close()
		l.Raw = nil
	}
}

func (l *Left) Close407() {
	if l.Raw != nil {
		realm := auth.GetResource("/").Realm
		l.Raw.Write([]byte(auth.Get407Response(realm, SERVER_NAME)));

		l.Raw.Close()
		l.Raw = nil
	}
}
