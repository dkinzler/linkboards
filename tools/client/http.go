package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

func buildUrl(base url.URL, path string, queryParams map[string]string) string {
	result := base
	result.Path = path

	qp := url.Values{}
	for key, value := range queryParams {
		qp.Add(key, value)
	}
	result.RawQuery = qp.Encode()

	return result.String()
}

func encodeJSONBody(x interface{}) io.Reader {
	b, _ := json.Marshal(x)
	return bytes.NewBuffer(b)
}

func decodeJSONBody(body io.Reader, result interface{}) error {
	return json.NewDecoder(body).Decode(result)
}

func newError(inner error, message string) error {
	m := message
	if inner != nil {
		m = fmt.Sprintf("%v, error: %v", message, inner)
	}
	return errors.New(m)
}

type jsonError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type jsonErrorWrapper struct {
	E jsonError `json:"error"`
}

type response struct {
	HttpCode int

	ErrorCode    int
	ErrorMessage string
}

type request struct {
	Method      string
	BaseUrl     url.URL
	Path        string
	QueryParams map[string]string
	Body        interface{}
	Headers     map[string]string
}

func doRequest(r request, respBody interface{}) (response, error) {
	ctx, cancel := getContext()
	defer cancel()

	var result response

	u := buildUrl(r.BaseUrl, r.Path, r.QueryParams)

	var body io.Reader
	if r.Body != nil {
		body = encodeJSONBody(r.Body)
	}

	request, err := http.NewRequestWithContext(ctx, r.Method, u, body)
	if err != nil {
		return result, newError(err, "could not create http request")
	}
	for key, value := range r.Headers {
		request.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return result, newError(err, "request failed")
	}

	result.HttpCode = resp.StatusCode

	if resp.StatusCode >= 300 {
		var e jsonErrorWrapper
		err = json.NewDecoder(resp.Body).Decode(&e)
		if err == nil {
			result.ErrorMessage = e.E.Message
			result.ErrorCode = e.E.Code
		} else if err != io.EOF {
			return result, newError(err, "could not decode error response body")
		}
	} else if respBody != nil {
		err = decodeJSONBody(resp.Body, respBody)
		if err != nil {
			return result, newError(nil, "could not decode response body")
		}
	}

	return result, nil
}
