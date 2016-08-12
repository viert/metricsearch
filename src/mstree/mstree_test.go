package mstree

import (
	"bufio"
	"fmt"
	logging "github.com/op/go-logging"
	"os"
	"testing"
	"time"
)

const (
	Data1 = "abook.qa-test1e_yandex_net.some.metric.total"
	Data2 = "abook.qa-test2e_yandex_net.some.metric.total"
	Data3 = "abook.qa-test1d_yandex_net.some.metric.total"
	Data4 = "abook.qa-test2d_yandex_net.some.metric.total"

	DataEmptyToken = "mail.mail_xivahub_var..xivahub.total.1xx"

	LongData = "r9eueKkpYgeezGQsVE5ENyYBpL5XOSgpaskNpNoeVCYfUghzKi89jck8RIcZy3jjOTIQJAAxfpmvFG03Ye2rrTM9c5uup41PCikq8idBObXxBxW07qBtfBP5mxy5MkuhKxCsdnyH06xMn0IF2sILrD9cGu0hFCWs28VwCl4vMifwGd25HIOOgSS3nh8PSsCD34FCgRYcqDnvKVc4s6V0STbTwTIAOHTel3NF56rTETpqAW1Y2XZhP1sbV9VLKMjKq4dfb2Cm7ZZy4JmTGNHtMBEW0M89lQXUnqn4KuLiirENoUaLo33c7L2lh3qWbDXoZQDGIfk7k0cJIL77pP5IKbTCGUEpogSwiuRbwfzhR09F7gZ3x3tDGUliUqV3qWJjtUjw0Qi2w2ixUDSI3OSsacJ90AULlzU8zz8Mbca21odiVuIL2I0uiPxKUOhD3HNdsgdvKugODDCp5acQRNRmoUkp8HkruEVzBCixcyQYdaM7LgbHbJL3i7Jyp3jQ0j8ovhNFrbtoSl074HrPPASrPQPStRWvKbd48dPJwQIfXTydjUmcIwWeEFRXoulA65xGliI1ybqOLXesGOPsaMq5R3Fdn2lFnvmBN1RBZGg4UtABWpRzu.some.valid.tokens"

	InvalidMetric = "'()&%<acx><ScRiPt >prompt(915633)<.if(some){.ops"

	TestExact      = "abook.qa-test1e_yandex_net.some.metric.total"
	TestStarEnd    = "abook.qa-test1e*"
	TestStarBegin  = "abook.*net.some.metric.total"
	TestStarMiddle = "abook.qa-test1*_yandex_net"
	TestStarLonely = "abook.*.some.metric.total"

	TestQuestionEnd     = "abook.qa-test1e_yandex_ne?.some.metric.total"
	TestQuestionBegin   = "abook.?a-test1e_yandex_net.some.metric.total"
	TestQuestionMiddle  = "abook.qa-test1?_yandex_net.some.metric.total"
	TestQuestionFailure = "abook.qa-test?_yandex_net.some.metric.total"

	TestHell    = "abook.q*test?e*.some.*.total"
	TestBraces  = "abook.qa-test[12][ed]*.some.metric.total"
	TestBraces2 = "abook.qa-test[12]e*.some.metric.total"
)

var (
	tree *MSTree
	err  error
)

func prepareTestTree(t testing.TB) {
	if tree != nil {
		return
	}
	dropTestTree()
	tree, err = NewTree("/tmp/test_index", 1000, true)
	if err != nil {
		t.Error(err)
	}
	err = tree.LoadIndex()
	if err != nil {
		t.Error(err)
	}
	tree.Add(Data1)
	tree.Add(Data2)
	tree.Add(Data3)
	tree.Add(Data4)
}

func dropTestTree() {
	os.RemoveAll("/tmp/test_index")
	tree = nil
}

func TestExactMatch(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestExact)
	if len(results) != 1 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
	if results[0] != Data1 {
		t.Errorf("Incorrect result:\n  Got %s\n  Expected %s", results[0], Data1)
	}
}

func TestLongDataDrop(t *testing.T) {
	prepareTestTree(t)
	if tree.TotalMetrics != 4 {
		t.Errorf("Incorrect metrics count on test start:\n  Got %d\n  Expected 4", tree.TotalMetrics)
	}
	tree.Add(LongData)
	if tree.TotalMetrics != 4 {
		t.Errorf("Incorrect metrics count on test end:\n  Got %d\n  Expected 4", tree.TotalMetrics)
	}
}

func TestInvalidDataDrop(t *testing.T) {
	prepareTestTree(t)
	if tree.TotalMetrics != 4 {
		t.Errorf("Incorrect metrics count on test start:\n  Got %d\n  Expected 4", tree.TotalMetrics)
	}
	tree.Add(InvalidMetric)
	if tree.TotalMetrics != 4 {
		t.Errorf("Incorrect metrics count on test end:\n  Got %d\n  Expected 4", tree.TotalMetrics)
	}
}

func TestStarAtTheEnd(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestStarEnd)
	mustMatch := "abook.qa-test1e_yandex_net."
	if len(results) != 1 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
	if results[0] != mustMatch {
		t.Errorf("Incorrect result:\n  Got %s\n  Expected %s", results[0], mustMatch)
	}
}

func TestStarAtTheBegin(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestStarBegin)
	if len(results) != 4 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
}

func TestStarAtTheMiddle(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestStarMiddle)
	if len(results) != 2 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
	for _, metric := range results {
		if metric != "abook.qa-test1d_yandex_net." && metric != "abook.qa-test1e_yandex_net." {
			t.Errorf("Unexpected metric %s", metric)
		}
	}
}

func TestStarLonelyMatch(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestStarLonely)
	if len(results) != 4 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
}

func TestQuestionAtTheEnd(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestQuestionEnd)
	if len(results) != 1 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
	if results[0] != Data1 {
		t.Errorf("Incorrect result:\n  Got %s\n  Expected %s", results[0], Data1)
	}
}

func TestQuestionAtTheBegin(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestQuestionBegin)
	if len(results) != 1 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
	if results[0] != Data1 {
		t.Errorf("Incorrect result:\n  Got %s\n  Expected %s", results[0], Data1)
	}
}

func TestQuestionAtTheMiddle(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestQuestionMiddle)
	if len(results) != 2 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
	for _, metric := range results {
		if metric != Data1 && metric != Data3 {
			t.Errorf("Unexpected metric %s", metric)
		}
	}
}

func TestQuestionNullMatch(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestQuestionFailure)
	if len(results) != 0 {
		for _, metric := range results {
			t.Errorf("Unexpected metric %s", metric)
		}
	}
}

func TestBracesPattern(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestBraces)
	if len(results) != 4 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
}

func TestBracesPattern2(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestBraces2)
	if len(results) != 2 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
}

func TestHellPattern(t *testing.T) {
	prepareTestTree(t)
	results := tree.Search(TestHell)
	if len(results) != 2 {
		t.Errorf("Incorrect results length: %d", len(results))
	}
	for _, metric := range results {
		if metric != Data1 && metric != Data2 {
			t.Errorf("Unexpected metric %s", metric)
		}
	}
}

func TestMetricCount(t *testing.T) {
	prepareTestTree(t)
	if tree.TotalMetrics != 4 {
		t.Errorf("Invalid metrics count after prepare: 4 expected, but %d got", tree.TotalMetrics)
	}
	retries := 10
	for retries > 0 && !tree.Synced() {
		retries--
		time.Sleep(10 * time.Millisecond)
	}
	if !tree.Synced() {
		t.Errorf("Error syncing tree, not synced after 10 retries 10ms each")
	}

	tree, err = NewTree("/tmp/test_index", 1000, true)
	if err != nil {
		t.Error(err)
	}
	tree.LoadIndex()
	if tree.TotalMetrics != 4 {
		t.Errorf("Invalid metrics count after loading index: 4 expected, but %d got", tree.TotalMetrics)
	}
}

func TestEmptyToken(t *testing.T) {
	prepareTestTree(t)
	if tree.TotalMetrics != 4 {
		t.Errorf("Invalid metrics count after prepare: 4 expected, but %d got", tree.TotalMetrics)
	}
	tree.Add(DataEmptyToken)
	if tree.TotalMetrics != 4 {
		t.Errorf("Empty token metric inserted into index!")
	}
}

func BenchmarkTreeAdd(b *testing.B) {
	dropTestTree()
	prepareTestTree(b)
	logging.SetLevel(logging.ERROR, "metricsearch")
	f, err := os.Open("payload.txt")
	if err != nil {
		b.Error("Please provide metric list in payload.txt file in the current directory")
		return
	}
	defer f.Close()

	payload := make([]string, 0, 10000000)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		payload = append(payload, sc.Text())
	}

	c := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Add(payload[c])
		c++
		if c == len(payload) {
			fmt.Println("Warning: Payload is too short")
			c = 0
		}
	}
}
