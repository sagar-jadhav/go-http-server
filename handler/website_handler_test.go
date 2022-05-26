package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sagar-jadhav/go-http-server/connector"
	"github.com/sagar-jadhav/go-http-server/model"
	"github.com/stretchr/testify/assert"
)

var urlToFile = map[string]string{
	"http://get-data-1.com": "test_data_url_1.json",
	"http://get-data-2.com": "test_data_url_2.json",
	"http://get-data-3.com": "test_data_url_3.json",
}

func TestValidateQueryParamsEmptySortKeyEmptyLimit(t *testing.T) {
	expected := &ErrorResponse{
		Message:   "sortKey and limit can't be empty. sortkey: , limit: ",
		ErrorCode: 400,
	}

	actual := ValidateQueryParams("", "")

	assert.Equal(t, expected, actual)
}

func TestValidateQueryParamsInvalidSortKey(t *testing.T) {
	expected := &ErrorResponse{
		Message:   "invalid value of sortKey test it could be either relevanceScore or views",
		ErrorCode: 400,
	}

	actual := ValidateQueryParams("test", "5")

	assert.Equal(t, expected, actual)
}

func TestValidateQueryParamsValidSortKeyInvalidLimit(t *testing.T) {
	expected := &ErrorResponse{
		Message:   "strconv.Atoi: parsing \"s\": invalid syntax",
		ErrorCode: 500,
	}

	actual := ValidateQueryParams("views", "s")

	assert.Equal(t, expected, actual)
}

func TestValidateQueryParamsValidSortKeyInvalidLimitValue(t *testing.T) {
	expected := &ErrorResponse{
		Message:   "invalid value of limit 0 it could be either greater than 1 or less than 200",
		ErrorCode: 400,
	}

	actual := ValidateQueryParams("views", "0")

	assert.Equal(t, expected, actual)
}

func TestSortWebsitesByViews(t *testing.T) {
	websites := []model.Website{
		{
			URL:            "http://localhost:8080/getData",
			Views:          2000,
			RelevanceScore: 0.1,
		},
		{
			URL:            "http://localhost:8080/getData",
			Views:          1000,
			RelevanceScore: 0.2,
		},
	}

	expectedWebsites := []model.Website{
		{
			URL:            "http://localhost:8080/getData",
			Views:          1000,
			RelevanceScore: 0.2,
		},
		{
			URL:            "http://localhost:8080/getData",
			Views:          2000,
			RelevanceScore: 0.1,
		},
	}

	SortWebsites(VIEWS, websites)

	assert.Equal(t, expectedWebsites, websites)
}

func TestSortWebsitesByRelevanceScore(t *testing.T) {
	websites := []model.Website{
		{
			URL:            "http://localhost:8080/getData",
			Views:          1000,
			RelevanceScore: 0.2,
		},
		{
			URL:            "http://localhost:8080/getData",
			Views:          2000,
			RelevanceScore: 0.1,
		},
	}

	expectedWebsites := []model.Website{
		{
			URL:            "http://localhost:8080/getData",
			Views:          2000,
			RelevanceScore: 0.1,
		},
		{
			URL:            "http://localhost:8080/getData",
			Views:          1000,
			RelevanceScore: 0.2,
		},
	}

	SortWebsites(RELEVANCE_SCORE, websites)

	assert.Equal(t, expectedWebsites, websites)
}

func TestLimitWebsitesInvalidLimit(t *testing.T) {
	websites := []model.Website{
		{
			URL:            "http://localhost:8080/getData",
			Views:          1000,
			RelevanceScore: 0.2,
		},
		{
			URL:            "http://localhost:8080/getData",
			Views:          2000,
			RelevanceScore: 0.1,
		},
	}

	_, errResp := LimitWebsites(10, websites)

	expectedErrResp := &ErrorResponse{
		Message:   "limit 10 can't be greater than the total websites 2",
		ErrorCode: 400,
	}

	assert.Equal(t, expectedErrResp, errResp)
}

func TestLimitWebsites(t *testing.T) {
	websites := []model.Website{
		{
			URL:            "http://localhost:8080/getData",
			Views:          1000,
			RelevanceScore: 0.2,
		},
		{
			URL:            "http://localhost:8080/getData",
			Views:          2000,
			RelevanceScore: 0.1,
		},
	}

	actual, _ := LimitWebsites(1, websites)

	expected := []model.Website{
		{
			URL:            "http://localhost:8080/getData",
			Views:          1000,
			RelevanceScore: 0.2,
		}}

	assert.Equal(t, expected, actual)
}

type MockHttpClient struct {
}

func (c *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	fileName := urlToFile[req.URL.String()]

	b, err := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", fileName))
	if err != nil {
		return nil, err
	}
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(b)),
		StatusCode: 200,
	}, nil
}

/**
 * SortKey = views, Limit = 2
 */
func TestGetAllWebsitesViewsAsSortKeyTwoLimit(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/websites?sortKey=views&limit=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	var urlList []string

	for key := range urlToFile {
		urlList = append(urlList, key)
	}

	websiteHandler := WebsiteHandler{
		URLList:    urlList,
		RetryCount: 5,
		Connector: &connector.HTTPConnector{
			Client: &MockHttpClient{},
		},
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(websiteHandler.GetAllWebsites)

	handler.ServeHTTP(rr, req)

	expectedStr := `{"data":[{"url":"www.example.com/abc1","views":1000,"relevanceScore":0.1},{"url":"www.example.com/abc2","views":2000,"relevanceScore":0.2}],"count":2}`
	expected, err := model.Object([]byte(expectedStr))
	if err != nil {
		t.Fatal(err)
	}

	actual, err := model.Object(rr.Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected, actual)
}

/**
 * SortKey = test, Limit = 2
 */
func TestGetAllWebsitesInvalidSortKeyValidLimit(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/websites?sortKey=test&limit=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	var urlList []string

	for key := range urlToFile {
		urlList = append(urlList, key)
	}

	websiteHandler := WebsiteHandler{
		URLList:    urlList,
		RetryCount: 5,
		Connector: &connector.HTTPConnector{
			Client: &MockHttpClient{},
		},
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(websiteHandler.GetAllWebsites)

	handler.ServeHTTP(rr, req)

	expected := ErrorResponse{
		Message:   "invalid value of sortKey test it could be either relevanceScore or views",
		ErrorCode: 400,
	}

	actual := ErrorResponse{}
	err = json.Unmarshal(rr.Body.Bytes(), &actual)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected, actual)
}
