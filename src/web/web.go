package web

import (
	"fmt"
	logging "github.com/op/go-logging"
	"io"
	"mstree"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

type Server struct {
	tree        *mstree.MSTree
	selfMonitor bool
}

type handlerCounters struct {
	add    uint64
	search uint64
	dump   uint64
}

type rpsCounters struct {
	add    float64
	search float64
	dump   float64
}

const (
	monitorHost = "127.0.0.1:42000"
)

var (
	log           *logging.Logger = logging.MustGetLogger("metricsearch")
	totalRequests handlerCounters
	lastRequests  handlerCounters
	rps           rpsCounters
	selfHostname  string
)

func (s *Server) sendMetrics() {
	conn, err := net.Dial("tcp", monitorHost)
	if err != nil {
		return
	}
	defer conn.Close()
	ts := time.Now().Unix()
	sqs, _ := s.tree.SyncQueueSize()
	fmt.Fprintf(conn, "%s.metricsearch.rps.add %.4f %d\n", selfHostname, rps.add, ts)
	fmt.Fprintf(conn, "%s.metricsearch.rps.search %.4f %d\n", selfHostname, rps.search, ts)
	fmt.Fprintf(conn, "%s.metricsearch.rps.dump %.4f %d\n", selfHostname, rps.dump, ts)
	fmt.Fprintf(conn, "%s.metricsearch.reqs.add %.2f %d\n", selfHostname, float32(totalRequests.add), ts)
	fmt.Fprintf(conn, "%s.metricsearch.reqs.search %.2f %d\n", selfHostname, float32(totalRequests.search), ts)
	fmt.Fprintf(conn, "%s.metricsearch.reqs.dump %.2f %d\n", selfHostname, float32(totalRequests.dump), ts)
	fmt.Fprintf(conn, "%s.metricsearch.metrics %.2f %d\n", selfHostname, float64(s.tree.TotalMetrics), ts)
	fmt.Fprintf(conn, "%s.metricsearch.sync_queue %.2f %d\n", selfHostname, float64(sqs), ts)
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint64(&totalRequests.search, 1)
	w.Header().Set("Content-Type", "text/plain")
	r.ParseForm()
	query := r.Form.Get("query")
	tm := time.Now()
	data := s.tree.Search(query)
	dur := time.Now().Sub(tm)
	if dur > time.Millisecond {
		// slower than 1ms
		log.Debug("Searching %s took %s\n", query, dur.String())
	}
	for _, item := range data {
		io.WriteString(w, item+"\n")
	}
}

func (s *Server) addHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint64(&totalRequests.add, 1)
	w.Header().Set("Content-Type", "text/plain")
	r.ParseForm()
	name := r.Form.Get("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Specify 'name' parameter")
		return
	}
	tm := time.Now()
	s.tree.Add(name)
	dur := time.Now().Sub(tm)
	if dur > time.Millisecond*100 {
		log.Debug("Indexing %s took %s\n", name, dur.String())
	}
	io.WriteString(w, "Ok")
}

func (s *Server) stackHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	buf := make([]byte, 65536)
	n := runtime.Stack(buf, true)
	w.Write(buf[:n])
}

func (s *Server) dumpHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint64(&totalRequests.dump, 1)
	w.Header().Set("Content-Type", "text/plain")
	s.tree.Root.TraverseDump("", w)
}

func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Total requests (online):\n=============================\n")
	io.WriteString(w, fmt.Sprintf("  add:    %d\n", totalRequests.add))
	io.WriteString(w, fmt.Sprintf("  search: %d\n", totalRequests.search))
	io.WriteString(w, fmt.Sprintf("  dump:   %d\n", totalRequests.dump))
	io.WriteString(w, "\n")
	io.WriteString(w, "RPS (refreshes every minute):\n=============================\n")
	io.WriteString(w, fmt.Sprintf("  add:    %.3f\n", rps.add))
	io.WriteString(w, fmt.Sprintf("  search: %.3f\n", rps.search))
	io.WriteString(w, fmt.Sprintf("  dump:   %.3f\n", rps.dump))
	io.WriteString(w, "\n")
	sqs, _ := s.tree.SyncQueueSize()
	io.WriteString(w, fmt.Sprintf("Total Metrics: %d\n", s.tree.TotalMetrics))
	io.WriteString(w, fmt.Sprintf("Sync Queue Size: %d\n", sqs))
}

func (s *Server) recalcRPS() {
	ticker := time.Tick(time.Minute)
	for _ = range ticker {
		rps.add = float64(totalRequests.add-lastRequests.add) / 60
		rps.dump = float64(totalRequests.dump-lastRequests.dump) / 60
		rps.search = float64(totalRequests.search-lastRequests.search) / 60
		lastRequests = totalRequests
		if s.selfMonitor {
			s.sendMetrics()
		}
	}
}

func NewServer(tree *mstree.MSTree, selfMonitor bool) *Server {
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	selfHostname = strings.Replace(host, ".", "_", -1)
	server := &Server{tree, selfMonitor}
	http.HandleFunc("/search", server.searchHandler)
	http.HandleFunc("/add", server.addHandler)
	http.HandleFunc("/debug/stack", server.stackHandler)
	http.HandleFunc("/dump", server.dumpHandler)
	http.HandleFunc("/stats", server.statsHandler)
	return server
}

func (s *Server) Start(listenAddr string) {
	log.Notice("Starting background stats job")
	go s.recalcRPS()
	log.Notice("Starting HTTP")
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}
}
