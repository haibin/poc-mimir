package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
	"github.com/golang/protobuf/proto"
)

func main() {
	// URL of Grafana Mimir remote write endpoint
	url := "http://localhost:9009/api/v1/push"

	// Create a sample metric: example_metric{instance="example-instance", job="example-job"} = 42
	labels := []prompb.Label{
		{Name: "__name__", Value: "example_metric"},
		{Name: "instance", Value: "example-instance"},
		{Name: "job", Value: "example-job"},
	}

	// Current Unix timestamp in milliseconds
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Create a sample
	sample := prompb.Sample{
		Value:     42,
		Timestamp: timestamp,
	}

	// Create time series with labels and sample
	ts := prompb.TimeSeries{
		Labels:  labels,
		Samples: []prompb.Sample{sample},
	}

	// Create the WriteRequest
	req := &prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{ts},
	}

	// Marshal to protobuf bytes
	data, err := proto.Marshal(req)
	if err != nil {
		panic(fmt.Errorf("marshal error: %w", err))
	}

	// Compress with snappy
	compressed := snappy.Encode(nil, data)

	// Create POST request
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(compressed))
	if err != nil {
		panic(fmt.Errorf("error creating request: %w", err))
	}

	// Set required headers
	httpReq.Header.Set("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	httpReq.Header.Set("X-Scope-OrgID", "123")

	// Perform request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		panic(fmt.Errorf("request error: %w", err))
	}
	defer resp.Body.Close()

	fmt.Printf("Status code: %d\n", resp.StatusCode)
}