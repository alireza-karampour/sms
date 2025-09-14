package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/onsi/gomega"
)

// HTTPClient provides methods for making HTTP requests in tests
type HTTPClient struct {
	BaseURL string
	Client  *http.Client
}

// NewHTTPClient creates a new HTTP client for testing
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RequestOptions contains options for HTTP requests
type RequestOptions struct {
	Headers map[string]string
	Body    interface{}
}

// Get makes a GET request
func (c *HTTPClient) Get(path string, opts ...RequestOptions) (*http.Response, error) {
	return c.makeRequest("GET", path, opts...)
}

// Post makes a POST request
func (c *HTTPClient) Post(path string, opts ...RequestOptions) (*http.Response, error) {
	return c.makeRequest("POST", path, opts...)
}

// Put makes a PUT request
func (c *HTTPClient) Put(path string, opts ...RequestOptions) (*http.Response, error) {
	return c.makeRequest("PUT", path, opts...)
}

// Delete makes a DELETE request
func (c *HTTPClient) Delete(path string, opts ...RequestOptions) (*http.Response, error) {
	return c.makeRequest("DELETE", path, opts...)
}

// makeRequest makes an HTTP request with the given method and path
func (c *HTTPClient) makeRequest(method, path string, opts ...RequestOptions) (*http.Response, error) {
	url := c.BaseURL + path
	
	var body io.Reader
	var headers map[string]string
	
	if len(opts) > 0 {
		opt := opts[0]
		headers = opt.Headers
		
		if opt.Body != nil {
			jsonData, err := json.Marshal(opt.Body)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(jsonData)
		}
	}
	
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	
	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	// Set custom headers
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}
	
	return c.Client.Do(req)
}

// AssertResponseStatus asserts that the response has the expected status code
func AssertResponseStatus(resp *http.Response, expectedStatus int) {
	gomega.Expect(resp.StatusCode).To(gomega.Equal(expectedStatus), 
		fmt.Sprintf("Expected status %d, got %d", expectedStatus, resp.StatusCode))
}

// ParseJSONResponse parses JSON response body into the given interface
func ParseJSONResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(body, target)
}

// AssertJSONResponse asserts that the response contains the expected JSON
func AssertJSONResponse(resp *http.Response, expected interface{}) {
	defer resp.Body.Close()
	
	var actual interface{}
	err := ParseJSONResponse(resp, &actual)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	
	gomega.Expect(actual).To(gomega.Equal(expected))
}

// TestData contains common test data structures
type TestData struct {
	User        UserData
	PhoneNumber PhoneNumberData
	SMS         SMSData
}

type UserData struct {
	Username string  `json:"username"`
	Balance  float64 `json:"balance"`
}

type PhoneNumberData struct {
	PhoneNumber string `json:"phone_number"`
}

type SMSData struct {
	ToPhoneNumber string `json:"to_phone_number"`
	Message       string `json:"message"`
}

// GetTestData returns common test data
func GetTestData() TestData {
	return TestData{
		User: UserData{
			Username: "testuser",
			Balance:  100.0,
		},
		PhoneNumber: PhoneNumberData{
			PhoneNumber: "+1234567890",
		},
		SMS: SMSData{
			ToPhoneNumber: "+0987654321",
			Message:       "Test SMS message",
		},
	}
}

// JSONBody creates a JSON body from the given data
func JSONBody(data interface{}) io.Reader {
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return bytes.NewBuffer(jsonData)
}