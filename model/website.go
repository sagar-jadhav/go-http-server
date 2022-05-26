package model

import (
	"encoding/json"
	"fmt"
)

type Website struct {
	URL            string  `json:"url"`
	Views          int     `json:"views"`
	RelevanceScore float32 `json:"relevanceScore"`
}

type Websites struct {
	Data  []Website `json:"data"`
	Count int       `json:"count"`
}

// Converts byte array into Websites object
func Object(data []byte) (*Websites, error) {
	w := &Websites{}
	err := json.Unmarshal(data, w)
	if err != nil {
		return nil, fmt.Errorf("error in converting byte array into websites object. error: %v", err)
	}
	return w, nil
}

// Converts Websites object into byte array
func Bytes(websites *Websites) ([]byte, error) {
	data, err := json.Marshal(websites)
	if err != nil {
		return nil, fmt.Errorf("error in converting websites object into byte array. error: %v", err)
	}
	return data, nil
}
