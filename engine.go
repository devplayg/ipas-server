package ipasserver

import (
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/devplayg/golibs/secureconfig"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

const (
	DateDefault = "2006-01-02 15:04:05"
)

var (
	CmdFlags *flag.FlagSet
	encKey   []byte
)

func init() {
	CmdFlags = flag.NewFlagSet("", flag.ExitOnError)
	key := sha256.Sum256([]byte("D?83F4 E?E"))
	encKey = key[:]
}

type Engine struct {
	ConfigPath  string
	Interval    int64
	Config      map[string]string
	appName     string
	debug       bool
	cpuCount    int
	processName string
	logOutput   int // 0: STDOUT, 1: File
}

func NewEngine(appName string, debug bool, cpuCount int, verbose bool) *Engine {
	e := Engine{
		appName:     appName,
		processName: GetProcessName(),
		cpuCount:    cpuCount,
		debug:       debug,
	}
	e.ConfigPath = filepath.Join(filepath.Dir(os.Args[0]), e.processName+".enc")
	e.initLogger(verbose)
	return &e
}

func (e *Engine) Start() error {
	var err error

	e.Config, err = secureconfig.GetConfig(e.ConfigPath, GetEncryptionKey())
	if err != nil {
		return err
	}
	if _, ok := e.Config["db.hostname"]; !ok {
		return errors.New("invalid configurations")
	}

	err = e.initDatabase()
	if err != nil {
		return err
	}

	log.Debug("Engine started")
	runtime.GOMAXPROCS(e.cpuCount)
	log.Debugf("GOMAXPROCS set to %d", runtime.GOMAXPROCS(0))
	return nil
}

func (e *Engine) initLogger(verbose bool) error {
	// Set log format
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		DisableColors: true,
	})

	// Set log level
	if e.debug {
		log.SetLevel(log.DebugLevel)
		orm.Debug = false
	}

	if verbose {
		e.logOutput = 0
		log.SetOutput(os.Stdout)
		orm.DebugLog = orm.NewLog(os.Stdout)
	} else {
		var logFile string
		if e.debug {
			logFile = filepath.Join(filepath.Dir(os.Args[0]), e.processName+"-debug.log")
			os.Remove(logFile)

		} else {
			logFile = filepath.Join(filepath.Dir(os.Args[0]), e.processName+".log")
		}

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err == nil {
			log.SetOutput(file)
			e.logOutput = 1
			orm.DebugLog = orm.NewLog(file)
		} else {
			e.logOutput = 0
			log.SetOutput(os.Stdout)
			orm.DebugLog = orm.NewLog(os.Stdout)
		}
	}

	if log.GetLevel() != log.InfoLevel {
		log.Infof("LoggingLevel=%s", log.GetLevel())
	}

	return nil
}

func (e *Engine) initDatabase() error {
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?allowAllFiles=true&charset=utf8&parseTime=true&loc=%s",
		e.Config["db.username"],
		e.Config["db.password"],
		e.Config["db.hostname"],
		e.Config["db.port"],
		e.Config["db.database"],
		"Asia%2FSeoul")
	log.Debugf("[db] hostname=%s, username=%s, port=%s, database=%s", e.Config["db.hostname"], e.Config["db.username"], e.Config["db.port"], e.Config["db.database"])
	err := orm.RegisterDataBase("default", "mysql", connStr, 3, 3)
	return err
}

func WaitForSignals() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalCh:
		log.Println("Signal received, shutting down...")
	}
}

func PrintHelp() {
	fmt.Println(strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])))
	CmdFlags.PrintDefaults()
}

func DisplayVersion(prodName, version string) {
	fmt.Printf("%s, v%s\n", prodName, version)
}

func GetProcessName() string {
	return strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
}

func GetEncryptionKey() []byte {
	return encKey
}