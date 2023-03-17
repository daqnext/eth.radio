package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"

	"github.com/daqnext/eth.radio/api"
	"github.com/daqnext/eth.radio/cmd"
	"github.com/daqnext/eth.radio/config"
	"github.com/daqnext/eth.radio/ensc"
	"github.com/ethereum/go-ethereum/ethclient"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"github.com/wealdtech/go-ens/v3"
)

const DOMAIN_SUFFIX = `(.*)[.][a-zA-Z]`

func setupApp() *echo.Echo {
	app := echo.New()

	return app
}

var appEnv config.AppEnv

func loagConfig() {

	viper.SetConfigName("config") // actually 'fileName + yml'
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error reading config file : ")
	}

	err := viper.Unmarshal(&appEnv)
	if err != nil {
		log.Fatal("unable to decode into struct : ")
	}
}

func main() {

	cmd.CmdOption()
	flag.Parse()

	if cmd.Opt.H {
		flag.Usage()
		return
	}
	if cmd.Opt.V {
		cmd.Version()
		return
	}

	loagConfig()

	app := setupApp()

	client, err := ethclient.Dial(appEnv.Connection)
	if err != nil {
		log.Fatal(err)
	}

	// Obtain the registry contract
	registry, err := ens.NewRegistry(client)
	if err != nil {
		log.Fatal(err)
	}

	ens_ctx := ensc.ENS{
		Client:   client,
		Registry: registry,
	}

	target_url, _ := url.Parse(appEnv.Ipfsgateway)
	proxy := httputil.NewSingleHostReverseProxy(target_url)

	app.Use(func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			host_domain := c.Request().Host
			req_url := c.Request().URL
			resource := req_url.Path
			// log.Info("path ", resource)
			if resource == "/" {
				resource = ""
			}

			Regexp, _ := regexp.Compile(DOMAIN_SUFFIX)

			grp := Regexp.FindStringSubmatch(host_domain)
			if len(grp) <= 1 {
				return c.String(http.StatusOK, host_domain)
			}

			ethDomain := grp[1] + "."
			log.Info("ethDomain ", ethDomain)
			if ethDomain == "eth." || ethDomain == "www.eth." {
				return handlerFunc(c)
			}

			hashcontent, err := ens_ctx.Query(ethDomain, ethDomain)
			if err != nil {
				log.Error("Error ", err.Error())
				return c.String(http.StatusOK, host_domain)
			}

			req := c.Request()
			res := c.Response().Writer

			api.BuildTransport(req, target_url, hashcontent)

			// Note that ServeHttp is non blocking and uses a go routine under the hood
			proxy.ServeHTTP(res, req)
			return nil
		}
	})

	api.DeclareRoute(app)

	host := appEnv.Host
	port := appEnv.Port

	if host == "*" {
		host = ""
	}

	if port == 0 {
		port = 8009
	}

	if cmd.Opt.Host != "" {
		host = cmd.Opt.Host
	}
	if cmd.Opt.Port != 0 {
		port = cmd.Opt.Port
	}

	app.Start(fmt.Sprintf("%s:%d", host, port))
}
