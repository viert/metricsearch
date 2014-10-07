package mstree

import (
	"bufio"
	"fmt"
	logging "github.com/op/go-logging"
	"io"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type MSTree struct {
	indexDir           string
	Root               *node
	syncBufferSize     int
	indexWriteChannels map[string]chan string
	indexWriterLock    *sync.Mutex
	fullReindex        bool
}
type eventChan chan error
type TreeCreateError struct {
	msg string
}

func (tce *TreeCreateError) Error() string {
	return tce.msg
}

var (
	log *logging.Logger = logging.MustGetLogger("metricsearch")
)

func NewTree(indexDir string, syncBufferSize int) (*MSTree, error) {
	stat, err := os.Stat(indexDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(indexDir, os.FileMode(0755))
			if err != nil {
				log.Error(err.Error())
				return nil, err
			}
		} else {
			log.Error(err.Error())
			return nil, err
		}
	} else {
		if !stat.IsDir() {
			log.Error("'%s' exists and is not a directory", indexDir)
			return nil, &TreeCreateError{fmt.Sprintf("'%s' exists and is not a directory", indexDir)}
		}
	}
	indexWriteChannels := make(map[string]chan string)
	root := newNode()
	tree := &MSTree{indexDir, root, syncBufferSize, indexWriteChannels, new(sync.Mutex), false}
	log.Debug("Tree created. indexDir: %s syncBufferSize: %d", indexDir, syncBufferSize)
	log.Debug("Background index sync started")
	return tree, nil
}

func separateSyncWorker(indexDir string, indexToken string, dataChannel chan string) {
	var err error
	idxFilename := fmt.Sprintf("%s/%s.idx", indexDir, indexToken)

	f, err := os.OpenFile(idxFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(0644))
	if err != nil {
		log.Critical("Error opening indexFile %s for writing: %s", idxFilename, err.Error())
		panic(err)
	}
	defer f.Close()
	for line := range dataChannel {
		if line == "" {
			continue
		} else {
			_, err := io.WriteString(f, line+"\n")
			if err != nil {
				log.Error("Index update error: %s", err.Error())
				continue
			} else {
				log.Debug("Metric '%s.%s' synced to disk", indexToken, line)
			}
		}
	}
}

func dumpWorker(idxFile string, idxNode *node, ev eventChan) {
	log.Debug("<%s> dumper started", idxFile)
	f, err := os.Create(idxFile)
	if err != nil {
		log.Debug("<%s> dumper finished with error: %s", idxFile, err.Error())
		ev <- err
		return
	}
	defer f.Close()
	idxNode.traverseDump("", f)
	log.Debug("<%s> dumper finished", idxFile)
	ev <- nil
}

func loadWorker(idxFile string, idxNode *node, ev eventChan) {
	log.Debug("<%s> loader started", idxFile)
	f, err := os.Open(idxFile)
	if err != nil {
		log.Error("<%s> loader finished with error: %s", idxFile, err.Error())
		ev <- err
		return
	}
	defer f.Close()
	inserted := true
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n")
		tokens := strings.Split(line, ".")
		idxNode.insert(tokens, &inserted)
	}
	log.Debug("<%s> loader finished", idxFile)
	ev <- nil
}

func (t *MSTree) AddNoSync(metric string) bool {
	if metric == "" {
		return false
	}
	tokens := strings.Split(metric, ".")
	inserted := false
	t.Root.insert(tokens, &inserted)
	return inserted
}

func (t *MSTree) Add(metric string) {
	inserted := t.AddNoSync(metric)
	if inserted {
		delimPos := strings.Index(metric, ".")
		if delimPos <= 0 || delimPos == len(metric)-1 {
			return
		}
		indexToken := metric[:delimPos]
		metricTail := metric[delimPos+1:]
		ch, ok := t.indexWriteChannels[indexToken]
		if !ok {
			tm := time.Now()
			t.indexWriterLock.Lock()
			ch = make(chan string, t.syncBufferSize)
			t.indexWriteChannels[indexToken] = ch
			t.indexWriterLock.Unlock()
			go separateSyncWorker(t.indexDir, indexToken, ch)
			log.Notice("Writer created for %s.idx in %s", indexToken, time.Now().Sub(tm).String())
		}
		ch <- metricTail
	}
}

func (t *MSTree) LoadTxt(filename string, limit int) error {
	t.fullReindex = true
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Turn GC off
	prevGC := debug.SetGCPercent(-1)
	// Defer to turn GC back on
	defer debug.SetGCPercent(prevGC)

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n")
		t.AddNoSync(line)
		count++
		if count%1000000 == 0 {
			log.Info("Reindexed %d items", count)
		}
		if limit != -1 && count == limit {
			break
		}
	}
	log.Info("Reindexed %d items", count)
	err = t.DumpIndex()
	if err != nil {
		return err
	}
	t.fullReindex = false
	return nil
}

func (t *MSTree) DropIndex() error {
	files, err := ioutil.ReadDir(t.indexDir)
	if err != nil {
		log.Error("Error opening index: " + err.Error())
		return err
	}
	if len(files) > 0 {
		for _, file := range files {
			fName := fmt.Sprintf("%s/%s", t.indexDir, file.Name())
			if strings.HasSuffix(fName, ".idx") {
				err := os.Remove(fName)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *MSTree) DumpIndex() error {
	log.Info("Syncinc the entire index")
	err := os.MkdirAll(t.indexDir, os.FileMode(0755))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	procCount := 0
	ev := make(eventChan, len(t.Root.Children))
	for first, node := range t.Root.Children {
		idxFile := fmt.Sprintf("%s/%s.idx", t.indexDir, first)
		go dumpWorker(idxFile, node, ev)
		procCount++
	}
	var globalErr error = nil
	for e := range ev {
		procCount--
		if e != nil {
			globalErr = e
		}
		if procCount == 0 {
			break
		}
	}
	log.Info("Sync complete")
	return globalErr
}

func (t *MSTree) LoadIndex() error {
	var globalErr error = nil
	files, err := ioutil.ReadDir(t.indexDir)
	if err != nil {
		log.Error("Error loading index: " + err.Error())
		return err
	}
	if len(files) > 0 {

		// Turn GC off
		prevGC := debug.SetGCPercent(-1)
		// Defer to turn GC back on
		defer debug.SetGCPercent(prevGC)

		ev := make(eventChan, len(files))
		procCount := 0
		for _, idxFile := range files {
			fName := idxFile.Name()
			if !strings.HasSuffix(fName, ".idx") {
				continue
			}
			pref := fName[:len(fName)-4]
			fName = fmt.Sprintf("%s/%s", t.indexDir, fName)
			idxNode := newNode()
			t.Root.Children[pref] = idxNode
			go loadWorker(fName, idxNode, ev)
			procCount++
		}
		tm := time.Now()

		for e := range ev {
			procCount--
			if e != nil {
				globalErr = e
			}
			if procCount == 0 {
				break
			}
		}
		log.Notice("Index load complete in %s", time.Now().Sub(tm).String())
	} else {
		log.Debug("Index is empty. Hope that's ok")
	}
	return globalErr
}

func (t *MSTree) Search(pattern string) []string {
	tokens := strings.Split(pattern, ".")
	nodesToSearch := make(map[string]*node)
	nodesToSearch[""] = t.Root
	for _, token := range tokens {
		prefRes := make(map[string]*node)
		for k, node := range nodesToSearch {
			sRes := node.search(token)
			if k == "" {
				// root node, no prefix
				for j, resNode := range sRes {
					prefRes[j] = resNode
				}
			} else {
				for j, resNode := range sRes {
					prefRes[k+"."+j] = resNode
				}
			}
		}
		nodesToSearch = prefRes
	}
	results := make([]string, len(nodesToSearch))
	i := 0
	for k, node := range nodesToSearch {
		if len(node.Children) == 0 {
			results[i] = k
		} else {
			results[i] = k + "."
		}
		i++
	}
	return results
}
