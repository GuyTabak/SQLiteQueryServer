package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	json "github.com/json-iterator/go"
)

var dbPath = "./test_db/ip_dns.db"

func TestResultCount(t *testing.T) {
	log.SetOutput(&bytes.Buffer{})

	reqString := "github.com\none.one.one.one\ngoogle-public-dns-a.google.com"

	req := httptest.NewRequest("POST",
		"http://example.org/query",
		strings.NewReader(reqString))
	w := httptest.NewRecorder()
	queryHandler, err := initQueryHandler(dbPath, "SELECT * FROM ip_dns WHERE dns = ?", 0)
	if err != nil {
		t.Fatal(err)
	}
	queryHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`resp.StatusCode (%d) != http.StatusOK (%d)`, resp.StatusCode, http.StatusOK)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf(`resp.Header.Get("Content-Type") (%s) != "application/json"`, resp.Header.Get("Content-Type"))
	}

	var answer []interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&answer)
	if err != nil {
		t.Fatal(err)
	}

	if len(answer) != 3 {
		t.Fatal(`len(answer) != 3`)
	}
}

func TestAnswersOrder(t *testing.T) {
	log.SetOutput(&bytes.Buffer{})

	reqString := "github.com\none.one.one.one\ngoogle-public-dns-a.google.com"

	req := httptest.NewRequest("POST",
		"http://example.org/query",
		strings.NewReader(reqString))
	w := httptest.NewRecorder()
	queryHandler, err := initQueryHandler(dbPath, "SELECT * FROM ip_dns WHERE dns = ?", 0)
	if err != nil {
		t.Fatal(err)
	}
	queryHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`resp.StatusCode (%d) != http.StatusOK (%d)`, resp.StatusCode, http.StatusOK)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf(`resp.Header.Get("Content-Type") (%s) != "application/json"`, resp.Header.Get("Content-Type"))
	}

	var answer []httpAnswer
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&answer)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		if len(answer[i].In) != 1 {
			t.Fatalf(`len(answer[%d].In) != 1`, i)
		}
	}

	for i, v := range []string{"github.com", "one.one.one.one", "google-public-dns-a.google.com"} {
		if answer[i].In[0] != v {
			t.Fatalf(`answer[%d].In[0] != "%s"`, i, v)
		}
	}
}

func TestAnswersHeaders(t *testing.T) {
	log.SetOutput(&bytes.Buffer{})

	reqString := "github.com\none.one.one.one\ngoogle-public-dns-a.google.com"

	req := httptest.NewRequest("POST",
		"http://example.org/query",
		strings.NewReader(reqString))
	w := httptest.NewRecorder()
	queryHandler, err := initQueryHandler(dbPath, "SELECT * FROM ip_dns WHERE dns = ?", 0)
	if err != nil {
		t.Fatal(err)
	}
	queryHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`resp.StatusCode (%d) != http.StatusOK (%d)`, resp.StatusCode, http.StatusOK)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf(`resp.Header.Get("Content-Type") (%s) != "application/json"`, resp.Header.Get("Content-Type"))
	}

	var answer []httpAnswer
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&answer)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		if len(answer[i].Headers) != 2 {
			t.Fatalf(`len(answer[%d].Headers) != 2`, i)
		}
	}

	for i := 0; i < 3; i++ {
		if answer[i].Headers[0] != "ip" {
			t.Fatalf(`answer[%d].In[0] != "ip"`, i)
		}
		if answer[i].Headers[1] != "dns" {
			t.Fatalf(`answer[%d].In[1] != "dns"`, i)
		}
	}
}

func TestAnswersRows(t *testing.T) {
	log.SetOutput(&bytes.Buffer{})

	reqString := "github.com\none.one.one.one\ngoogle-public-dns-a.google.com"

	req := httptest.NewRequest("POST",
		"http://example.org/query",
		strings.NewReader(reqString))
	w := httptest.NewRecorder()
	queryHandler, err := initQueryHandler(dbPath, "SELECT * FROM ip_dns WHERE dns = ?", 0)
	if err != nil {
		t.Fatal(err)
	}
	queryHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`resp.StatusCode (%d) != http.StatusOK (%d)`, resp.StatusCode, http.StatusOK)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf(`resp.Header.Get("Content-Type") (%s) != "application/json"`, resp.Header.Get("Content-Type"))
	}

	var answer []httpAnswer
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&answer)
	if err != nil {
		t.Fatal(err)
	}

	expectedAnswer := []httpAnswer{
		httpAnswer{
			Out: [][]interface{}{
				[]interface{}{"192.30.253.112", "github.com"},
				[]interface{}{"192.30.253.113", "github.com"},
			}},
		httpAnswer{
			Out: [][]interface{}{
				[]interface{}{"1.1.1.1", "one.one.one.one"},
			}},
		httpAnswer{
			Out: [][]interface{}{
				[]interface{}{"8.8.8.8", "google-public-dns-a.google.com"},
			}},
	}

	for i, v := range expectedAnswer {
		if len(v.Out) != len(answer[i].Out) {
			t.Fatalf(`len(v.Out) (%v) != len(answer[%d].Out) (%v)`, len(v.Out), i, len(answer[i].Out))
		}

		for rowI, rowV := range v.Out {
			if len(rowV) != len(answer[i].Out[rowI]) {
				t.Fatalf(`len(rowV) (%v) != len(answer[%d].Out[%d]) (%v)`, len(rowV), i, rowI, len(answer[i].Out[rowI]))
			}

			for cellI, cellV := range rowV {
				if cellV != answer[i].Out[rowI][cellI] {
					t.Fatalf(`cellV (%v) != answer[%d].Out[%d][%d] (%v)`, cellV, i, rowI, cellI, answer[i].Out[rowI][cellI])
				}
			}
		}
	}
}

func TestBadParamsCount(t *testing.T) {
	log.SetOutput(&bytes.Buffer{})

	reqString := "github.com,1"

	req := httptest.NewRequest("POST",
		"http://example.org/query",
		strings.NewReader(reqString))
	w := httptest.NewRecorder()
	queryHandler, err := initQueryHandler(dbPath, "SELECT * FROM ip_dns WHERE dns = ?", 0)
	if err != nil {
		t.Fatal(err)
	}
	queryHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf(`resp.StatusCode (%d) != http.StatusInternalServerError (%d)`, resp.StatusCode, http.StatusInternalServerError)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	respString := string(respBytes)

	if !strings.Contains(respString, "sql: expected 1 arguments, got 2") {
		t.Fatal(`Error string should contain "sql: expected 1 arguments, got 2"`)
	}
}

func TestBadPathRequest(t *testing.T) {
	log.SetOutput(&bytes.Buffer{})

	reqString := `github.com`

	req := httptest.NewRequest("POST",
		"http://example.org/queri",
		strings.NewReader(reqString))
	w := httptest.NewRecorder()
	queryHandler, err := initQueryHandler(dbPath, "SELECT * FROM ip_dns WHERE dns = ?", 0)
	if err != nil {
		t.Fatal(err)
	}
	queryHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf(`resp.StatusCode (%d) != http.StatusNotFound (%d)`, resp.StatusCode, http.StatusNotFound)
	}
}

func TestBadMethodRequest(t *testing.T) {
	log.SetOutput(&bytes.Buffer{})

	req := httptest.NewRequest("GET",
		"http://example.org/query",
		nil)
	w := httptest.NewRecorder()
	queryHandler, err := initQueryHandler(dbPath, "SELECT * FROM ip_dns WHERE dns = ?", 0)
	if err != nil {
		t.Fatal(err)
	}
	queryHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf(`resp.StatusCode (%d) != http.StatusMethodNotAllowed (%d)`, resp.StatusCode, http.StatusMethodNotAllowed)
	}
}
