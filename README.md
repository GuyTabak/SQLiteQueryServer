# SQLiteQueryServer

Bulk query SQLite database over the network.

# Installation

```bash
go get -u github.com/assafmo/SQLiteQueryServer
```

# Usage

```
Usage of SQLiteQueryServer:
  -db string
        Filesystem path of the SQLite database
  -port uint
        HTTP port to listen on (default 80)
  -query string
        SQL query to prepare for
```

# Examples

## Server

```bash
SQLiteQueryServer --db ./db_example/ip_dns.db --query "SELECT * FROM ip_dns WHERE dns = ?" --port 8080
```

```bash
SQLiteQueryServer --db "$DB_PATH" --query "$PARAMETERIZED_SQL_QUERY" --port "$PORT"
```

This will expose the `./db_example/ip_dns.db` database with the query `SELECT * FROM ip_dns WHERE dns = ?` on port `8080`.  
Requests will need to provide the query parameters.

## Request

```bash
echo -e "github.com\none.one.one.one\ngoogle-public-dns-a.google.com" | curl "http://localhost:8080/query" --data-binary @-
```

```bash
echo -e "$QUERY1_PARAM1,$QUERY1_PARAM2\n$QUERY2_PARAM1,$QUERY2_PARAM2" | curl "http://$ADDRESS:$PORT/query" --data-binary @-
```

```bash
curl "http://$ADDRESS:$PORT/query" -d "$PARAM_1,$PARAM_2,...,$PARAM_N"
```

- Request must be a HTTP POST to "http://$ADDRESS:$PORT/query".
- Request body must be a valid CSV.
- Request body must not have a CSV header.
- Each request body line is a different query.
- Each param in a line corresponds to a query param (a question mark in the query string).

## Response

```bash
echo -e "github.com\none.one.one.one\ngoogle-public-dns-a.google.com" | curl "http://localhost:8080/query" --data-binary @-
```

```json
[
  {
    "in": ["github.com"],
    "headers": ["ip", "dns"],
    "out": [["192.30.253.112", "github.com"], ["192.30.253.113", "github.com"]]
  },
  {
    "in": ["one.one.one.one"],
    "headers": ["ip", "dns"],
    "out": [["1.1.1.1", "one.one.one.one"]]
  },
  {
    "in": ["google-public-dns-a.google.com"],
    "headers": ["ip", "dns"],
    "out": [["8.8.8.8", "google-public-dns-a.google.com"]]
  }
]
```

- Response is a JSON array (Content-Type: application/json).
- Each element in the array:
  - Is a result of a query
  - Has an "in" fields which is an array of the input params (a request body line).
  - Has an "headers" fields which is an array of headers of the SQL query result.
  - Has an "out" field which is an array of arrays of results. Each inner array is a result row.
- Element #1 is the result of query #1, Element #2 is the result of query #2, and so forth.
