package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sagar-jadhav/go-http-server/connector"
	"github.com/sagar-jadhav/go-http-server/handler"
)

func main() {
	urlListStr, present := os.LookupEnv("URL_LIST")
	if !present {
		log.Fatal("URL_LIST environment variable not set. Please set to start the server")
	}

	retryCountStr, present := os.LookupEnv("RETRY_COUNT")
	if !present {
		log.Fatal("RETRY_COUNT environment variable not set. Please set to start the server")
	}

	retryCount, err := strconv.Atoi(retryCountStr)
	if err != nil {
		log.Fatal(err)
	}

	urlList := strings.Split(urlListStr, ",")

	log.Printf("URL_LIST: %s", urlList)
	log.Printf("RETRY_COUNT: %d", retryCount)

	handler := &handler.WebsiteHandler{
		URLList:    urlList,
		RetryCount: retryCount,
		Connector: &connector.HTTPConnector{
			Client: &http.Client{},
		},
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/websites", handler.GetAllWebsites)

	log.Println("Server is listening at port 9090.")
	log.Fatal(http.ListenAndServe(":9090", r))
}
