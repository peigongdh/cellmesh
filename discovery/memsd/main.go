package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davyxu/cellmesh/discovery"
	"github.com/davyxu/cellmesh/discovery/memsd/api"
	"github.com/davyxu/cellmesh/discovery/memsd/model"
)

var (
	flagCmd      = flag.String("cmd", "", "sub command, empty to launch memsd service")
	flagAddr     = flag.String("addr", "", "service discovery address")
	flagDataFile = flag.String("datafile", "", "persist values to file")
	flagDebug    = flag.Bool("debug", false, "show debug info")
	flagVersion  = flag.Bool("version", false, "show version")
)

type DiscoveryExtend interface {
	discovery.Discovery

	QueryAll() (ret []*discovery.ServiceDesc)

	ClearKey()

	ClearService()

	GetRawValueList(prefix string) (ret []discovery.ValueMeta)
}

func initSD() DiscoveryExtend {
	config := memsd.DefaultConfig()
	if *flagAddr != "" {
		config.Address = *flagAddr
	}

	return memsd.NewDiscovery(config).(DiscoveryExtend)
}

func main() {

	flag.Parse()

	if *flagVersion {
		fmt.Println("version", model.Version)
		return
	}

	model.Debug = *flagDebug

	go startCheckRedundantValue()

	if *flagDataFile != "" {
		loadPersistFile(*flagDataFile)
		go startPersistCheck(*flagDataFile)
	}

	switch *flagCmd {
	case "": // addr
		startSvc()
	case "viewsvc": // addr
		ViewSvc()
	case "viewkey": // addr
		ViewKey()
	case "clearsvc": // addr
		ClearSvc()
	case "clearvalue": // addr
		ClearValue()
	case "deletevalue":
		if flag.NArg() < 1 {
			fmt.Println("deletevalue <key>")
			os.Exit(1)
		}
		DeleteValue(flag.Arg(0))
	case "getvalue":
		if flag.NArg() < 1 {
			fmt.Println("getvalue <key>")
			os.Exit(1)
		}
		GetValue(flag.Arg(0))
	case "setvalue":
		if flag.NArg() < 2 {
			fmt.Println("setvalue <key> <value>")
			os.Exit(1)
		}

		SetValue(flag.Arg(0), flag.Arg(1))
	default:
		fmt.Printf("Unknown command '%s'\n", *flagCmd)
		os.Exit(1)
	}
}
