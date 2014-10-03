package main

import (
	"config"
	"flag"
	"fmt"
	logging "github.com/op/go-logging"
	"mstree"
	"runtime"
	"runtime/debug"
	"web"
)

func main() {
	format := "%{color}%{level} %{color:reset}%{message}"
	logging.SetFormatter(logging.MustStringFormatter(format))
	log := logging.MustGetLogger("metricsearch")

	var confFile, reindexFile string
	flag.StringVar(&confFile, "c", "/etc/metricsearch.conf", "metricsearch config filename")
	flag.StringVar(&reindexFile, "reindex", "", "reindex from plain text metrics file")
	flag.Parse()

	conf := config.Load(confFile)
	tree, err := mstree.NewTree(conf.IndexDirectory, conf.SyncBufferSize)
	if err != nil {
		log.Critical("No way to continue, exiting.")
		return
	}

	log.Debug("Configuring runtime: GCPercent(%d), MaxCores(%d), MaxThreads(%d)", conf.GCPercent, conf.MaxCores, conf.MaxThreads)
	runtime.GOMAXPROCS(conf.MaxCores)
	debug.SetGCPercent(conf.GCPercent)
	debug.SetMaxThreads(conf.MaxThreads)
	if reindexFile != "" {
		err := tree.LoadTxt(reindexFile, -1)
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
