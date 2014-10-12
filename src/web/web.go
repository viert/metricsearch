package web

import (
	logging "github.com/op/go-logging"
	"io"
	"mstree"
	"net/http"
	"runtime"
	"time"
)

type Server struct {
	tree *mstree.MSTree
}

var (
	log *logging.Logger = logging.MustGetLogger("metricsearch")
)

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-Type", "text/plain")
	s.tree.Root.TraverseDump("", w)
}

func NewServer(tree *mstree.MSTree) *Server {
	server := &Server{tree}
	http.HandleFunc("/search", server.searchHandler)
	http.HandleFunc("/add", server.addHandler)
	http.HandleFunc("/debug/stack", server.stackHandler)
	http.HandleFunc("/dump", server.dumpHandler)
	return server
}

func (s *Server) Start(listenAddr string) {
	log.Notice("Starting HTTP")
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}
}
