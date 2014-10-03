package main

import (
	"bufio"
	"config"
	"flag"
	"fmt"
	logging "github.com/op/go-logging"
	"mstree"
	"os"
	"runtime"
	"runtime/debug"
	"web"
)

func main() {
	var format string
	log := logging.MustGetLogger("metricsearch")

	var confFile, reindexFile string
	var stdinImport bool
	flag.StringVar(&confFile, "c", "/etc/metricsearch.conf", "metricsearch config filename")
	flag.StringVar(&reindexFile, "reindex", "", "reindex from plain text metrics file")
	flag.BoolVar(&stdinImport, "stdin", false, "reindex from stdin")
	flag.Parse()

	conf := config.Load(confFile)
	logging.SetLevel(conf.LogLevel, "metricsearch")

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
		f, err := os.OpenFile(conf.Log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(0644))
		if err != nil {
			log.Error(err.Error())
			return
		}
		backend := logging.NewLogBackend(f, "", 0)
		logging.SetBackend(backend)
		defer f.Close()
		format = "[%{time:2006-01-02 15:04:05}] %{level} %{message}"
	}
	logging.SetFormatter(logging.MustStringFormatter(format))

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
