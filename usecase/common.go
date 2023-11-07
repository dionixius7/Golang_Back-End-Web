package usecase

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func NewHttpRequest(method string, endpoint string, payload interface{}) (*http.Request, error) {
	// Check if the payload is nil
	if payload == nil {
		// If the payload is nil, create a new request with an empty body
		return http.NewRequest(method, endpoint, nil)
	}
	// Convert the payload to JSON
	payloadInByte, err := json.Marshal(payload)
	if err != nil {
		// If an error occurred while converting to JSON, return the error
		return nil, err
	}
	// Create a new request with the payload included in the request body
	return http.NewRequest(method, endpoint, bytes.NewBuffer(payloadInByte))
}
