package connector

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHttpClientSuccess struct {
}

func (c *MockHttpClientSuccess) Do(req *http.Request) (*http.Response, error) {
	b, err := ioutil.ReadFile("./testdata/test_data.json")
	if err != nil {
		return nil, err
	}
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(b)),
		StatusCode: 200,
	}, nil
}

type MockHttpClientFailure struct {
	mock.Mock
}

func (c *MockHttpClientFailure) Do(req *http.Request) (*http.Response, error) {
	c.Called()

	return &http.Response{
		StatusCode: 500,
	}, nil
}

func TestGetSuccess(t *testing.T) {
	connector := &HTTPConnector{
		Client: &MockHttpClientSuccess{},
	}

	actual, err := connector.Get("http://localhost:8080/getData", 5)
	if err != nil {
		t.Fail()
	}

	expected, err := ioutil.ReadFile("./testdata/test_data.json")
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, expected, actual)
}

func TestGetFail(t *testing.T) {
	var mockConn MockHttpClientFailure = MockHttpClientFailure{}
	connector := &HTTPConnector{
		Client: &mockConn,
	}

	// 6 = retrycount + 1
	mockConn.On("Do").Times(6)

	_, err := connector.Get("random_url", 5)
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf("error in fetching the data from the server. url: random_url, status code: 500"), err)
	}
}
