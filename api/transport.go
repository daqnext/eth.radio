package api

import (
	"net/http"
	"net/url"

	"github.com/labstack/gommon/log"
)

func BuildTransport(req *http.Request,
	target *url.URL, prefix string) error {

	req.Host = target.Host
	req.URL.Host = target.Host
	req.URL.Scheme = target.Scheme

	path := req.URL.Path
	req.URL.Path = prefix + path

	log.Info("req.Host ", target)
	log.Info("req.URL.Path ", req.URL.Path)

	return nil

}
