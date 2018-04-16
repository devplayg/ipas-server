package main

import (
	"github.com/devplayg/golibs/secureconfig"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/calculator"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
)

const (
	AppName    = "IPAS Statistics Calculator"
	AppVersion = "1.0.1804.11501"
)

func main() {
	// CPU 설정
	runtime.GOMAXPROCS(2)

	// 옵션 설정
	var (
		version   = ipasserver.CmdFlags.Bool("version", false, "Version")
		debug     = ipasserver.CmdFlags.Bool("debug", false, "Debug")
		verbose   = ipasserver.CmdFlags.Bool("v", false, "Verbose")
		setConfig = ipasserver.CmdFlags.Bool("config", false, "Edit configurations")
		top       = ipasserver.CmdFlags.Int("top", 100, "Top N")
		interval  = ipasserver.CmdFlags.Int64("i", 2000, "Interval(ms)")
		date      = ipasserver.CmdFlags.String("date", "", "Target date")
		report    = ipasserver.CmdFlags.String("report", "", "Period report(StartDate,EndDate,MarkDate)")
	)
	ipasserver.CmdFlags.Usage = ipasserver.PrintHelp
	ipasserver.CmdFlags.Parse(os.Args[1:])

	// 버전 출력
	if *version {
		ipasserver.DisplayVersion(AppName, AppVersion)
		return
	}

	// 엔진 설정
	engine := ipasserver.NewEngine(AppName, *debug, *verbose)
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
		log.Fatal(err)
	}
	log.Debug(engine.Config)

	// 데이터베이스 연결
	if err := engine.InitDatabase(); err != nil {
		log.Fatal(err)
	}

	// 데이터 분류 시작
	cal := calculator.NewCalculator(engine, *top, *interval, *date, *report)
	if err := cal.Start(); err != nil {
		log.Fatal(err)
	}

	// 종료 시그널 대기
	if len(*date) < 1 && len(*report) < 1 {
		ipasserver.WaitForSignals()
	}
}