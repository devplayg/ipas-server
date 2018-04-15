package main

import (
	"github.com/devplayg/golibs/secureconfig"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/receiver"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"runtime"
	"time"
)

const (
	AppName    = "IPAS Receiver"
	AppVersion = "1.0.1804.11001"
)

func main() {

	// CPU 설정
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 옵션 설정
	var (
		version         = ipasserver.CmdFlags.Bool("version", false, "Version")
		debug           = ipasserver.CmdFlags.Bool("debug", false, "Debug")
		verbose         = ipasserver.CmdFlags.Bool("v", false, "Verbose")
		setConfig       = ipasserver.CmdFlags.Bool("config", false, "Edit configurations")
		batchSize       = ipasserver.CmdFlags.Int("batchsize", 1000, "Batch size")
		batchTimeout    = ipasserver.CmdFlags.Int("batchtime", 1000, "Batch timeout, in milliseconds")
		batchMaxPending = ipasserver.CmdFlags.Int("maxpending", 10000, "Maximum pending events")
		httpport        = ipasserver.CmdFlags.String("port", ":8080", "HTTP port")
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
	defer engine.Stop()
	log.Debug(engine.Config)

	// Receiver(수집기) 시작
	timeout := time.Duration(*batchTimeout) * time.Millisecond
	stacker := receiver.NewStacker(*batchSize, timeout, *batchMaxPending, engine)
	errChan := make(chan error)
	if err := stacker.Start(errChan); err != nil {
		log.Fatalf("failed to start Stacker: %s", err.Error())
	}
	log.Debugf("batching configured with size %d, timeout %s, max pending %d",
		*batchSize, timeout, *batchMaxPending)

	// 에러 출력
	go drainLog("error batch", errChan)

	// HTTP 라우터 시작
	if err := startHttpServer(stacker, *httpport); err != nil {
		log.Fatal(err)
	}

	// 종료 시그널 대기
	ipasserver.WaitForSignals()
}

func drainLog(msg string, errChan <-chan error) {
	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Errorf("%s: %s", msg, err.Error())
			}
		}
	}
}

func startHttpServer(stacker *receiver.Stacker, httpport string) error {
	router := httprouter.New()

	r1 := receiver.NewEventReceiver(router) // 로그 수신기
	r1.Start(stacker.C())
	r2 := receiver.NewStatusReceiver(router) // 상태정보 수신기
	r2.Start(stacker.C())

	go func() {
		if err := http.ListenAndServe(httpport, router); err != nil {
			log.Fatal(err)
		}
	}()

	return nil
}
