package auth

// https://www.ietf.org/rfc/rfc2617.txt

type User struct {
	Name      string
	Password  string
	Realms    []string
}

func (u *User) HasRealm(realm string) bool {
	for _, r := range u.Realms {
		if r == realm {
			return true
		}
	}

	return false;
}

type Resource struct {
	Path  string
	Realm string
}

const (
	RESPONSE_TPL_407 = `HTTP/1.0 407 Proxy Authentication Required
Date: %s
Server: %s
Proxy-Authenticate: Basic realm="%s"
status: 407 Proxy Authentication
Content-Length: %d
Connection: close

`
)

var users []User
var resources []Resource

func SetUsers(u []User) {
	users = u
}

func SetResources(r []Resource) {
	resources = r
}

func GetUser(username string) (user User) {
	for i := 0; i < len(users); i++ {
		if users[i].Name == username {
			user = users[i]
			break
		}
	}

	return user
}

func GetResource(path string) (resource Resource) {
	for i := 0; i < len(resources); i++ {
		if resources[i].Path == path {
			resource = resources[i]
			break
		}
	}

	return resource
}
