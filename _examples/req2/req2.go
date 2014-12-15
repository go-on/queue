package main

/*
as suggested by https://groups.google.com/d/msg/golang-nuts/9MUZ9WEAz6k/Y6St2rNJdxoJ

variant with external error handling

the get struct, the getBody function and the handleError are reusable in
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

	"gopkg.in/go-on/queue.v2"
	. "gopkg.in/go-on/queue.v2/q"
)

type User struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

var client = &http.Client{}

type get struct {
	*http.Request
}

func (g *get) Get(url string) (err error) {
	g.Request, err = http.NewRequest("GET", url, nil)
	return
}

func (g *get) AddHeader(key, val string) {
	g.Request.Header.Add(key, val)
}

func (g *get) DoRequest() (resp *http.Response, err error) {
	return client.Do(g.Request)
}

func getBody(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil, aborting")
	}
	switch resp.StatusCode {
	case 304:
		// in reality you would return a CacheNotFound error, if there
		// is nothing in the cache
		return []byte(`something from the cache`), nil
	case 200:
		return ioutil.ReadAll(resp.Body)
	default:
		return nil, fmt.Errorf("Status code is not 200, but: %d", resp.StatusCode)
	}
}

var handleError = queue.ErrHandlerFunc(func(err error) error {
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError, *json.InvalidUTF8Error:
			fmt.Printf("invalid json response: %s\n", err)
		default:
			fmt.Printf("Error: %s\n", err)
		}
	}
	// always pass the error back to stop further execution
	return err
})

func main() {
	g := &get{}
	user := &User{}

	defer func() {
		if g.Request != nil && g.Request.Body != nil {
			g.Request.Body.Close()
		}
	}()

	Err(handleError)(
		g.Get, "https://api.github.com/users/metakeule",
	)(
		g.AddHeader, "X-Somekey", "some value",
	)(
		g.DoRequest,
	)(
		getBody, V,
	)(
		json.Unmarshal, V, user,
	).Run()

	fmt.Printf("user is %v\n", user)
}
