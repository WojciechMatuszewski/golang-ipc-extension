// Package ipc provides an server for the lambda function to use as an endpoint for ssm parameter
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/patrickmn/go-cache"
)

const (
	cacheKey      = "parameter"
	parameterName = "extension-parameter"
)

type server struct {
	port    string
	service ssmiface.SSMAPI
}

// New creates new ipc server
func New(port string, service ssmiface.SSMAPI) *server {
	return &server{port, service}
}

func (s *server) Start(ctx context.Context) {
	c := cache.New(30*time.Second, 1*time.Minute)
	handler := newHandler(s.service, c)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	svr := http.Server{
		Addr:        fmt.Sprintf(":%s", s.port),
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     mux,
	}

	go func() {
		log.Println("server starting")

		err := svr.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal("Listen and serve failure", err)
		}
	}()

	<-ctx.Done()
	shutdowCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	log.Println("server shutting down!")
	err := svr.Shutdown(shutdowCtx)
	if err != nil {
		log.Printf("HTTP server ListenAndServe: %v", err)
	}
}

type response struct {
	Body string `json:"body"`
}

func newHandler(service ssmiface.SSMAPI, c *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		param, found := c.Get(cacheKey)
		if found {
			respondSuccess(w, string(param.(string)))
			return
		}

		resp, err := service.GetParameter(&ssm.GetParameterInput{Name: aws.String(parameterName)})
		if err != nil {
			respondError(w, err)
			return
		}

		c.Set(cacheKey, *resp.Parameter.Value, cache.DefaultExpiration)
		respondSuccess(w, *resp.Parameter.Value)
	}
}

func respondSuccess(w http.ResponseWriter, value string) {
	w.Header().Set("Content-Type", "application/json")

	resp := response{Body: value}
	buf, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(buf)
}

func respondError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	resp := response{Body: err.Error()}
	buf, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(buf)
}
