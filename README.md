metricsearch
============

standalone graphite metric name search (for usage with custom storage backends)

build instructions:
-------------------
```
git clone git@github.com:viert/metricsearch.git
cd metricsearch
export GOPATH=`pwd`
go get github.com/viert/properties
go get github.com/op/go-logging
go build src/metricsearch.go
```

usage:
------

```
metricsearch 
  -c="/etc/metricsearch.conf": metricsearch config filename
  -reindex="": reindex from plain text metrics file
```

metrics file is a text file with metric names separated by "\n"

metricsearch listens at port 7000 by default and has the following http handlers:

`/add?name=<metricname>` adds metric **metricname** to index, automatically syncing it to disk in background.

`/search?query=<searchquery>` searches for metrics. Metric names are returned line by line, partials (for graphite /metrics/find) are flagged by the following ".". For exapmle:

```
curl "http://localhost:7000/search?query=addressbook.*"
addressbook.host1.
addressbook.host2.
addressbook.total_rps
```
This means host1 and host2 are graphite directories, whereas total_rps is a complete leaf with timeseries.
