package auth

import (
	"encoding/base64"
	"bytes"
	"errors"
	"fmt"
	"time"
	"strings"
)

var (
	ERR_MISSING_USERNAME_OR_PASSWORD = errors.New("missing username or password")
	ERR_MISSING_SPACE_OR_AUTH        = errors.New("mssing space or auth")
	ERR_INVALID_USERNAME_OR_PASSWORD = errors.New("invalid username or password")
	ERR_NO_PERMISSION                = errors.New("no permission")
)

func decodeAuth(auth string) (username string, password string, err error) {
	var (
		up    []byte
		idx   int
	)

	up, err = base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return "", "", err
	}

	idx = bytes.Index(up, []byte(":"))
	if idx == -1 {
		return "", "", ERR_MISSING_USERNAME_OR_PASSWORD
	}

	return string(up[0:idx]), string(up[idx + 1:]), nil
}

func Get407Response(realm string, server string) string {
	date := time.Now().Format(time.RFC1123)
	return fmt.Sprintf(RESPONSE_TPL_407, date, server, realm, 0)
}

func Basic(path string, auth string) error {
	var (
		username  string
		password  string
		err       error
		u         User
		r         Resource
		p         []string
	)

	r = GetResource(path)

	if auth == "" || r.Realm == "" {
		return ERR_MISSING_SPACE_OR_AUTH
	}

	if p = strings.Split(auth, " "); len(p) != 2 {
		return ERR_INVALID_USERNAME_OR_PASSWORD
	}

	auth = p[1]
	if username, password, err = decodeAuth(auth); err != nil {
		return err
	}

	u = GetUser(username)

	if u.Name == "" || u.Password != password {
		return ERR_INVALID_USERNAME_OR_PASSWORD
	}

	if !u.HasRealm(r.Realm) {
		return ERR_NO_PERMISSION
	}

	return nil
}
