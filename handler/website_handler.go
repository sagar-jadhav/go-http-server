package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/sagar-jadhav/go-http-server/connector"
	"github.com/sagar-jadhav/go-http-server/model"
)

const (
	RELEVANCE_SCORE string = "relevanceScore"
	VIEWS           string = "views"
	SORT_KEY        string = "sortKey"
	LIMIT           string = "limit"
)

type WebsiteHandler struct {
	URLList    []string
	RetryCount int
	Connector  *connector.HTTPConnector
}

type WebsiteChannel struct {
	Websites []model.Website
	Error    error
}

type ErrorResponse struct {
	Message   string
	ErrorCode int
}

// Gets the website
func getWebsite(url string, retryCount int, c chan WebsiteChannel, wg *sync.WaitGroup, connector *connector.HTTPConnector) {
	defer wg.Done()

	data, err := connector.Get(url, retryCount)
	if err == nil {
		websites, err := model.Object(data)
		if err == nil {
			c <- WebsiteChannel{
				Websites: websites.Data,
				Error:    nil,
			}
			return
		}
	}

	c <- WebsiteChannel{
		Websites: nil,
		Error:    err,
	}
}

// Validates request params
func ValidateQueryParams(sortKey string, limitStr string) *ErrorResponse {
	if len(sortKey) == 0 || len(limitStr) == 0 {
		return &ErrorResponse{
			Message:   fmt.Sprintf("sortKey and limit can't be empty. sortkey: %s, limit: %s", sortKey, limitStr),
			ErrorCode: http.StatusBadRequest,
		}
	}

	if !(sortKey == RELEVANCE_SCORE || sortKey == VIEWS) {
		return &ErrorResponse{
			Message:   fmt.Sprintf("invalid value of sortKey %s it could be either relevanceScore or views", sortKey),
			ErrorCode: http.StatusBadRequest,
		}
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return &ErrorResponse{
			Message:   err.Error(),
			ErrorCode: http.StatusInternalServerError,
		}
	}

	if limit <= 1 || limit >= 200 {
		return &ErrorResponse{
			Message:   fmt.Sprintf("invalid value of limit %d it could be either greater than 1 or less than 200", limit),
			ErrorCode: http.StatusBadRequest,
		}
	}

	return nil
}

// Sorts the websites in ascending order
func SortWebsites(sortKey string, websites []model.Website) {
	switch sortKey {
	case RELEVANCE_SCORE:
		sort.Sort(model.SortByRelevanceScoreWebsites(websites))
	case VIEWS:
		sort.Sort(model.SortByViewsWebsites(websites))
	}
}

// Limits the websites by given limit
func LimitWebsites(limit int, websites []model.Website) ([]model.Website, *ErrorResponse) {
	if limit > len(websites) {
		return nil, &ErrorResponse{
			Message:   fmt.Sprintf("limit %d can't be greater than the total websites %d", limit, len(websites)),
			ErrorCode: http.StatusBadRequest,
		}
	}

	return websites[:limit], nil
}

// Get all websites concurrently
func (handler WebsiteHandler) GetAllWebsites(w http.ResponseWriter, r *http.Request) {
	// Fetching the query params
	query := r.URL.Query()
	sortKey := query.Get(SORT_KEY)
	limitStr := query.Get(LIMIT)

	// Validating request params
	errResp := ValidateQueryParams(sortKey, limitStr)
	if errResp != nil {
		log.Printf("invalid request params. error: %v\n", errResp)
		w.WriteHeader(errResp.ErrorCode)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// Getting the websites concurrently
	var wg sync.WaitGroup

	c := make(chan WebsiteChannel)

	wg.Add(len(handler.URLList))

	for _, url := range handler.URLList {
		go getWebsite(url, handler.RetryCount, c, &wg, handler.Connector)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	websites := []model.Website{}

	for wc := range c {
		if wc.Error != nil {
			errResp := ErrorResponse{
				Message:   wc.Error.Error(),
				ErrorCode: http.StatusInternalServerError,
			}
			log.Printf("unable to handle request. error: %v\n", errResp)
			w.WriteHeader(errResp.ErrorCode)
			json.NewEncoder(w).Encode(errResp)
			return
		}

		websites = append(websites, wc.Websites...)
	}

	// Preparing the response
	response := model.Websites{
		Data: websites,
	}

	// Sorting the websites in ascending order
	SortWebsites(sortKey, response.Data)

	// Limiting the websites by given limit
	limit, _ := strconv.Atoi(limitStr)
	response.Data, errResp = LimitWebsites(limit, response.Data)
	if errResp != nil {
		log.Printf("unable to handle request. error: %v\n", errResp)
		w.WriteHeader(errResp.ErrorCode)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// Sending response back to the client
	response.Count = len(response.Data)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
