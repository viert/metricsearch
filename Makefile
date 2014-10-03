all: metricsearch

metricsearch: dependencies src/metricsearch.go src/mstree/mstree.go src/mstree/node.go src/config/config.go src/web/web.go
	GOPATH=$(CURDIR) /usr/local/go/bin/go build src/metricsearch.go

dependencies: src/github.com/op/go-logging src/github.com/viert/properties
	GOPATH=$(CURDIR) /usr/local/go/bin/go get github.com/op/go-logging
	GOPATH=$(CURDIR) /usr/local/go/bin/go get github.com/viert/properties

clean:
	rm -f metricsearch
