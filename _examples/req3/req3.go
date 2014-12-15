package main

/*
as suggested by https://groups.google.com/d/msg/golang-nuts/9MUZ9WEAz6k/Y6St2rNJdxoJ

variant with internal error handling

the get struct and the getBody function are reusable in
different requests.

TASK:

Do you have an example of, say, how one might make an HTTP request?

  1. Create request.
  2. Set some headers.
  3. Have client execute it.
  4. On err, do error thing.
  5. On success, verify HTTP status, consider it an error if not 200.
     5a. If 304, reuse existing content
  6. Parse JSON response.
  7. Handle JSON errors.
  8. Close request if we got past 4.
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	. "gopkg.in/go-on/queue.v2/q"
)

var client = &http.Client{}

func main() {
	user := &User{}

	err := Q(
		http.NewRequest, "GET", "https://api.github.com/users/metakeule", nil,
	).Tee(
		setHeader, V, "X-Somekey", "some value", // V is the request
	).TeeAndRun(
		Q( // new queue feeded with request
			client.Do, V, // V is the request
		)(
			json.Unmarshal, Fallback(
				Q(getCache, V), // V is the request
				Q(readBody, V), // V is the request
			), user,
		),
	).Tee(
		closeReq, V, // V is the request
	).
		LogErrorsTo(os.Stdout).
		Run()

	if err != nil {
		fmt.Printf("Error: %s (%T)\n", err, err)
		return
	}
	fmt.Printf("user is %v\n", user)
}

type User struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func setHeader(rq *http.Request, key, value string) {
	rq.Header.Add(key, value)
}

func getCache(resp *http.Response) ([]byte, error) {
	if resp.StatusCode == 304 {
		// in reality you would return a CacheNotFound error, if there
		// is nothing in the cache
		return []byte(`something from the cache`), nil
	}
	return nil, fmt.Errorf("not cached")
}

func readBody(resp *http.Response) ([]byte, error) {
	if resp.StatusCode == 200 {
		return ioutil.ReadAll(resp.Body)
	}
	return nil, fmt.Errorf("Status code is not 200, but: %d", resp.StatusCode)
}

func closeReq(req *http.Request) {
	if req != nil && req.Body != nil {
		req.Body.Close()
	}
}
