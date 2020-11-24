package metrics

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Prometheus struct {
	client  v1.API
	metrics []string
}

func NewPrometheus(address string) (p *Prometheus, err error) {
	resp, err := http.Get(address + "/api/v1/label/__name__/values")
	if err != nil {
		return
	}
	type Metrics struct {
		Status string
		Data   []string
	}
	var metrics Metrics
	err = json.NewDecoder(resp.Body).Decode(&metrics)
	if err != nil {
		return
	}
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return
	}
	p = &Prometheus{
		v1.NewAPI(client),
		metrics.Data,
	}
	return
}

func (p *Prometheus) Query(query string, start time.Time, end time.Time, step time.Duration) (model.Value, error) {
	query = strings.ReplaceAll(query, "%s", strconv.Itoa(int(step.Seconds()))+"s")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rng := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}
	res, _, err := p.client.QueryRange(ctx, query, rng)
	return res, err
}

func (p *Prometheus) PreciseQuery(query string, start time.Time, end time.Time) (val model.Value, err error) {
	step := 15 * time.Second
	for {
		val, err = p.Query(query, start, end, step)
		if err == nil {
			return
		}
		if strings.Index(err.Error(), "exceeded maximum resolution of") < 0 {
			return
		}
		step *= step
	}
}

func ValuesToFloatArray(val model.Value) TaggedValueSlice {
	var values TaggedValueSlice

	switch val.Type() {
	case model.ValVector:
		values := make([]float64, len(val.(model.Vector)))
		for i, sample := range val.(model.Vector) {
			values[i] = float64(sample.Value)
		}
	case model.ValScalar:
		values = TaggedValueSlice{WithTag(float64(val.(*model.Scalar).Value), strconv.FormatInt(int64(val.(*model.Scalar).Timestamp), 10))}
	case model.ValMatrix:
		values = make(TaggedValueSlice, 0)
		for _, stream := range val.(model.Matrix) {
			for _, sample := range stream.Values {
				values = append(values, WithTag(float64(sample.Value), strconv.FormatInt(int64(sample.Timestamp), 10)))
			}
		}
	}

	return values
}
