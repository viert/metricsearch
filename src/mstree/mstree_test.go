package mstree

import (
	"os"
	"testing"
)

const (
	Data1 = "abook.qa-test1e_yandex_net.some.metric.total"
	Data2 = "abook.qa-test2e_yandex_net.some.metric.total"
	Data3 = "abook.qa-test1d_yandex_net.some.metric.total"
	Data4 = "abook.qa-test2d_yandex_net.some.metric.total"

	TestExact      = "abook.qa-test1e_yandex_net.some.metric.total"
	TestStarEnd    = "abook.qa-test1e*"
	TestStarBegin  = "abook.*net.some.metric.total"
	TestStarMiddle = "abook.qa-test1*_yandex_net"
	TestStarLonely = "abook.*.some.metric.total"

	TestQuestionEnd     = "abook.qa-test1e_yandex_ne?.some.metric.total"
	TestQuestionBegin   = "abook.?a-test1e_yandex_net.some.metric.total"
	TestQuestionMiddle  = "abook.qa-test1?_yandex_net.some.metric.total"
	TestQuestionFailure = "abook.qa-test?_yandex_net.some.metric.total"

	TestHell = "abook.q*test?e*.some.*.total"
)

var (
	tree *MSTree
	err  error
)

func prepareTestTree(t *testing.T) {
	if tree != nil {
		return
	}
	tree, err = NewTree("/tmp/test_index", 1000)
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
	os.Remove("/tmp/test_index")
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
