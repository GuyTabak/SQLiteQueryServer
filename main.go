package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var queryStmt *sql.Stmt

var dbPath string
var queryString string
var serverPort uint

var helpMessege = `SQLiteQueryServer help messege:

`

func init() {
	flag.StringVar(&dbPath, "db", "", "Path to DB")
	flag.StringVar(&queryString, "query", "", "The SQL query")
	flag.UintVar(&serverPort, "port", 80, "Port of the http server")

	flag.Parse()

	var err error
	db, err = sql.Open("sqlite3", fmt.Sprintf("file:%s&mode=rw&cache=shared", dbPath))
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(1)

	queryStmt, err = db.Prepare(queryString)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Printf("DB = %s", dbPath)
	log.Printf("Query = %s", queryString)
	log.Printf("Port = %d", serverPort)

	http.HandleFunc("/query", query)
	err := http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)

	if err != nil {
		log.Fatal(err)
	}
}

func query(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/query" {
		http.Error(w, helpMessege, http.StatusNotFound)
		return
	}
	if r.Method != "POST" {
		http.Error(w, helpMessege, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	reqReader := csv.NewReader(bufio.NewReader(r.Body))
	for {
		csvLine, err := reqReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			http.Error(w,
				fmt.Sprintf("Error reading request body: %v\n\n%s", err, helpMessege), http.StatusInternalServerError)
			return
		}

		queryParams := make([]interface{}, len(csvLine))
		for i := range csvLine {
			queryParams[i] = csvLine[i]
		}

		rows, err := queryStmt.Query(queryParams...)
		if err != nil {
			http.Error(w,
				fmt.Sprintf("Error executing query for params %v: %v\n\n%s", csvLine, err, helpMessege), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row []interface{}
			err = rows.Scan(&row)
			if err != nil {
				http.Error(w,
					fmt.Sprintf("Error reading query results for params %v: %v\n\n%s", csvLine, err, helpMessege), http.StatusInternalServerError)
				return
			}

			fmt.Println(row)
			// TODO print row to w (as part of a json [{"in":csvLine, "out":[[],[],...,[]]}])
		}
		err = rows.Err()
		if err != nil {
			http.Error(w,
				fmt.Sprintf("Error executing query: %v\n\n%s", err, helpMessege), http.StatusInternalServerError)
			return
		}
	}
}
