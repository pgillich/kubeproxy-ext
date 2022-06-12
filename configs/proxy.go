package configs

import (
	"errors"
	"net/http"
)

const (
	ObjectKeyKubectl = "kubectl"
)

var (
	ErrGenerateTableNoRow    = errors.New("generatetable no row")
	ErrGenerateTableMoreRows = errors.New("generatetable more rows")
)

type Proxy struct {
	TargetURL  string
	ListenAddr string

	HTTPServer     HTTPServer        `mapstructure:"-"`
	ProxyTransport http.RoundTripper `mapstructure:"-"`
}
