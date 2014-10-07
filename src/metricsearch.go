package main

import (
	"bufio"
	"config"
	"flag"
	"fmt"
	logging "github.com/op/go-logging"
	"mstree"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"web"
)

const (
	DEFAULT_CONFIG_FILE = "/etc/metricsearch.conf"
)

var (
	logfileName string
	logfile     *os.File
	log         *logging.Logger = logging.MustGetLogger("metricsearch")
)

func reopenLog() {
	var err error
	if logfile != nil {
		logfile.Close()
	}
	logfile, err = os.OpenFile(logfileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(0644))
	if err != nil {
		return
	}
	backend := logging.NewLogBackend(logfile, "", 0)
	logging.SetBackend(backend)
}

func hupCatcher() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP)
	for _ = range c {
		log.Debug("HUP signal catched, reopening logfile %s", logfileName)
		reopenLog()
	}
}

func main() {
	var format string
	var confFile, reindexFile string
	var stdinImport bool
	flag.StringVar(&confFile, "c", DEFAULT_CONFIG_FILE, "metricsearch config filename")
	flag.StringVar(&reindexFile, "reindex", "", "reindex from plain text metrics file")
	flag.BoolVar(&stdinImport, "stdin", false, "reindex from stdin")
	flag.Parse()

	conf := config.Load(confFile)

	switch conf.Log {
	case "syslog":
		backend, err := logging.NewSyslogBackend("metricsearch")
		if err != nil {
			log.Error(err.Error())
			return
		}
		logging.SetBackend(backend)
		format = "%{program} %{level} %{message}"
	case "":
		format = "%{color}%{level} %{color:reset}%{message}"
	default:
		logfileName = conf.Log
		reopenLog()
		go hupCatcher()
		format = "[%{time:2006-01-02 15:04:05}] %{level} %{message}"
	}
	logging.SetFormatter(logging.MustStringFormatter(format))
	logging.SetLevel(conf.LogLevel, "metricsearch")

	tree, err := mstree.NewTree(conf.IndexDirectory, conf.SyncBufferSize)
	if err != nil {
		log.Critical("No way to continue, exiting.")
		return
	}

	log.Debug("Configuring runtime: GCPercent(%d), MaxCores(%d), MaxThreads(%d)", conf.GCPercent, conf.MaxCores, conf.MaxThreads)
	runtime.GOMAXPROCS(conf.MaxCores)
	debug.SetGCPercent(conf.GCPercent)
	debug.SetMaxThreads(conf.MaxThreads)

	if stdinImport {
		err := tree.DropIndex()
		if err != nil {
			log.Critical("Error dropping index: %s", err.Error())
			return
		}
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			tree.Add(sc.Text())
		}
		log.Notice("Reindexing complete")
		return
	}

	if reindexFile != "" {
		err := tree.DropIndex()
		if err != nil {
			log.Critical("Error dropping index: %s", err.Error())
			return
		}
		err = tree.LoadTxt(reindexFile, -1)
		if err != nil {
			log.Critical("Reindexing error, exiting.")
			return
		} else {
			log.Notice("Reindexing complete")
			return
		}
	} else {
		tree.LoadIndex()
		server := web.NewServer(tree)
		addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
		server.Start(addr)
	}

}
