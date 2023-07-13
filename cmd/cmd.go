package cmd

import (
	"flag"
	"fmt"
	"os"
)

type opt struct {
	H bool

	V bool

	Port int
	Host string
}

var Opt opt

func CmdOption() {
	flag.BoolVar(&Opt.H, "h", false, "this help")

	flag.BoolVar(&Opt.V, "V", false, "show version and exit")

	flag.StringVar(&Opt.Host, "b", "", "bind host")
	flag.IntVar(&Opt.Port, "p", 0, "set port")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `ens-server version: 1.10.0
Usage: ens-server  [-hvV] [-b host] [-p port]

Options:
`)
	flag.PrintDefaults()
}

func Version() {
	fmt.Fprintf(os.Stderr, `version: 1.10.0
`)
}
