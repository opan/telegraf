package http_out

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type HttpOut struct {
	Name    string
	Server  string
	Data    map[string]string
	Headers map[string]string
}

type Metric struct {
	Name   string                 `json:"name"`
	Fields map[string]interface{} `json:"fields"`
	Tags   map[string]string      `json:"tags"`
	Time   int64                  `json:"time"`
}

func (h *HttpOut) Description() string {
	return `Send telegraf metric through HTTP(s) request`
}

func (h *HttpOut) SampleConfig() string {
	return `
  [[outputs.http_out]]
    name = "http_out_test"
    server = "http://localhost:3000"

    [outputs.http_out.headers]
      Content-Type = "application/json;charset=UTF-8"

    [outputs.http_out.data]
      token = "YourDataToken"
`
}

// Connect to the Output
func (h *HttpOut) Connect() error {
	return nil
}

// Close any connections to the Output
func (h *HttpOut) Close() error {
	return nil
}

// Write takes in group of points to be written to the Output
func (h *HttpOut) Write(metrics []telegraf.Metric) error {
	// Don't make any request if metrics empty
	if len(metrics) == 0 {
		return nil
	}

	if h.Server == "" {
		return fmt.Errorf("You need to setup a server")
	}

	// Prepare URL
	requestURL, err := url.Parse(h.Server)
	if err != nil {
		return fmt.Errorf("Invalid server URL \"%s\"", h.Server)
	}

	// Collect metrics
	var Metrics []Metric
	for _, metric := range metrics {
		var timestamp time.Duration
		unitsNanoseconds := timestamp.Nanoseconds()

		// if the units passed in were less than or equal to zero,
		// then serialize the timestamp in seconds (the default)
		if unitsNanoseconds <= 0 {
			unitsNanoseconds = 1000000000
		}

		m := Metric{
			Name:   metric.Name(),
			Tags:   metric.Tags(),
			Fields: metric.Fields(),
			Time:   metric.Time().UnixNano() / unitsNanoseconds,
		}

		Metrics = append(Metrics, m)
	}

	// Setup request body to send metrics data
	var jsonReq struct {
		Metrics []Metric          `json:"metrics"`
		Data    map[string]string `json:"data"`
	}
	jsonReq.Metrics = Metrics
	if len(h.Data) > 0 {
		jsonReq.Data = h.Data
	}

	// Encode request body
	reqBody, err := json.Marshal(jsonReq)

	// Initialize HTTP(s) request
	req, err := http.NewRequest("POST", requestURL.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Errorf("Cannot setup HTTP request: %s", err)
	}

	// Add headers parameters
	for k, v := range h.Headers {
		req.Header.Add(k, v)
	}

	// Send HTTP(s) request
	client := http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()

	var parsedBody map[string]interface{}
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Errorf("Cannot real response body: %s", err)
	}

	err = json.Unmarshal([]byte(resBody), &parsedBody)
	if err != nil {
		fmt.Errorf("Cannot parse response body: %s", err)
	}

	return nil
}

func init() {
	outputs.Add("http_out", func() telegraf.Output {
		return &HttpOut{}
	})
}
