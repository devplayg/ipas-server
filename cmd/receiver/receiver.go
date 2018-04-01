package main

import (
	"github.com/devplayg/golibs/secureconfig"
	"github.com/devplayg/ipas-server"
	log "github.com/sirupsen/logrus"
	"os"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/devplayg/ipas-server/receiver"
)

const (
	AppName    = "IPAS Receiver"
	AppVersion = "1.0.1803.10801"
)

func main() {
	var (
		version   = ipasserver.CmdFlags.Bool("version", false, "Version")
		debug     = ipasserver.CmdFlags.Bool("debug", false, "Debug")
		cpu       = ipasserver.CmdFlags.Int("cpu", 3, "CPU Count")
		verbose   = ipasserver.CmdFlags.Bool("v", false, "Verbose")
		setConfig = ipasserver.CmdFlags.Bool("config", false, "Edit configurations")
		//interval  = ipasserver.CmdFlags.Int64("i", 5000, "Interval(ms)")
	)
	ipasserver.CmdFlags.Usage = ipasserver.PrintHelp
	ipasserver.CmdFlags.Parse(os.Args[1:])

	// 버전 출력
	if *version {
		ipasserver.DisplayVersion(AppName, AppVersion)
		return
	}

	// 엔진 설정
	engine := ipasserver.NewEngine(AppName, *debug, *cpu, *verbose)
	if *setConfig {
		secureconfig.SetConfig(
			engine.ConfigPath,
			"db.hostname, db.port, db.username, db.password, db.database",
			ipasserver.GetEncryptionKey(),
		)

		return
	}

	// 엔진 시작
	if err := engine.Start(); err != nil {
		log.Error(err)
		return
	}
	log.Debug(engine.Config)

	// 라우터 시작
	router := httprouter.New()
	receiver.NewEventReceiver(engine, router)
	receiver.NewStatusReceiver(engine, router)
	log.Fatal(http.ListenAndServe(":8080", router))

	// Wait for signal
	ipasserver.WaitForSignals()

}
