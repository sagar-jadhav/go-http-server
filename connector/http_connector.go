package connector

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HTTPConnector struct {
	Client HTTPClient
}

func get(url string, client HTTPClient) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error in fetching the data from the server. url: %s, status code: %d", url, resp.StatusCode)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Gets the data from the server
func (connector *HTTPConnector) Get(url string, retryCount int) ([]byte, error) {
	if retryCount < 0 {
		return nil, fmt.Errorf("retry count %d cannot be less than 0", retryCount)
	}

	data, err := get(url, connector.Client)
	if err != nil {
		for {
			if retryCount == 0 {
				break
			}

			log.Printf("retrying fething the data from the server. error: %v, retryCount: %d", err, retryCount)

			data, err = get(url, connector.Client)

			if err != nil {
				retryCount--
			}
		}
	}

	return data, err
}
