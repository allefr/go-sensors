package env

type Driver interface {
	String() string
	StringJSON() (string, error)
	IsConnected() error
}
