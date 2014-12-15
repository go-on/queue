package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/go-on/queue.v2"
	"gopkg.in/go-on/wrap.v2"
	"gopkg.in/go-on/wrap-contrib.v2/helper"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type ErrorWriter interface {
	// WriteError gets the ResponseWriter, the request and the error
	// and handles the error writing
	WriteError(http.ResponseWriter, *http.Request, error)
}

// ErrorResponse is an error handler, that is also a responsewriter
type ErrorResponse struct {
	*helper.ResponseBuffer
	// last error is the first error that stopped the queue
	LastError error
}

// HandleError fulfills the queue.ErrHandler interface and always returns an error to stop
// the queue
func (e *ErrorResponse) HandleError(in error) (out error) {
	e.LastError = in
	return in
}

type HTTPStatusError struct {
	Code    int
	Header  http.Header
	Message string
}

func (h HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP Status Code Error: Code %d, Message: %s", h.Code, h.Message)
}

type errorResponseWrapper struct {
	ErrorWriter
}

func NewErrorResponseWrapper(errHandler ErrorWriter) wrap.Wrapper {
	return &errorResponseWrapper{errHandler}
}

func (e *errorResponseWrapper) Wrap(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := helper.NewResponseBuffer(w)
		errResp := &ErrorResponse{buf, nil}
		inner.ServeHTTP(errResp, r)
		if buf.Code > 399 && errResp.LastError == nil {
			errResp.LastError = HTTPStatusError{buf.Code, buf.Header(), buf.Buffer.String()}
		}

		if errResp.LastError != nil {
			// should modify the responsebuffer
			e.WriteError(w, r, errResp.LastError)
		}

		buf.WriteHeadersTo(w)
		buf.WriteCodeTo(w)
		buf.WriteTo(w)
	})
}

type ValidationError interface {
	Error() string
	ValidationErrors() map[string][]string
}

// use it
type errorHandler struct {
	logger *log.Logger
}

func (eh errorHandler) log(r *http.Request, err error) {
	eh.logger.Printf("Error %T in %s %s: %s",
		err,
		r.Method,
		r.URL.String(),
		err.Error(),
	)
}

func (eh errorHandler) WriteError(w http.ResponseWriter, r *http.Request, err error) {
	eh.log(r, err)

	switch v := err.(type) {
	case HTTPStatusError:
		// write all headers but the X- ones
		for k, _ := range v.Header {
			if !strings.HasPrefix(strings.ToLower(k), "x-") {
				w.Header().Set(k, v.Header.Get(k))
			}
		}

		switch v.Code {
		case 301, 302:
			w.WriteHeader(v.Code)
		case 404:
			w.WriteHeader(404)
			// TODO: switch content type
			w.Write([]byte("beautiful 404 message"))
		default:
			w.WriteHeader(v.Code)
			if v.Code < 500 {
				w.Write([]byte(err.Error()))
			}
		}
	case *strconv.NumError:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Not a number"))
	case ValidationError:
		m, _ := json.Marshal(v.ValidationErrors())
		w.WriteHeader(http.StatusBadRequest)
		w.Write(m)
	default:
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
	}
}

func handleWithErrors(w http.ResponseWriter, r *http.Request) {
	queue.New().
		OnError(
		w.(queue.ErrHandler),
	).Add(
		strconv.Atoi, r.URL.Query().Get("number"),
	).Add(
		fmt.Fprintf,
		w,
		"parsed: %d",
		queue.PIPE,
	).CheckAndRun()
}

func main() {
	stack := wrap.New(
		NewErrorResponseWrapper(errorHandler{
			log.New(os.Stdout, "website errors ", log.Ltime|log.Ldate|log.Lshortfile),
		}),
		wrap.HandlerFunc(handleWithErrors),
	)

	http.ListenAndServe(":8085", stack)
}
