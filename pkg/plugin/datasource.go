package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/feldera/feldera/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/infinity-libs/lib/go/jsonframer"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(_ context.Context, setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, err := models.LoadPluginSettings(setting)
	if err != nil {
		return nil, err
	}

	return &Datasource{
		client: http.Client{},
		pipeline: settings.Pipeline,
		baseUrl: settings.BaseUrl,
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct{
	client http.Client
	baseUrl string
	pipeline string
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct{
	QueryText string		`json:"queryText"`
}

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	// TODO: query.TimeRange
	// query.MaxDataPoints
	// query.Interval
	
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusValidationFailed, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	sql := qm.QueryText

	if sql == "" {
		return response
	}

	sql = strings.ReplaceAll(sql, "$__timeFrom()", fmt.Sprintf("'%s'", query.TimeRange.From.UTC().Format(time.RFC3339)))
	sql = strings.ReplaceAll(sql, "$__timeTo()", fmt.Sprintf("'%s'", query.TimeRange.To.UTC().Format(time.RFC3339)))

	println(sql)

	params := url.Values{}
	params.Add("format", "json")
	params.Add("sql", sql)

	url := fmt.Sprintf("%s/v0/pipelines/%s/query?%s", d.baseUrl, d.pipeline, params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("failed to create http request: %v", err.Error()))
	}

	if pCtx.DataSourceInstanceSettings != nil {
		config, err := models.LoadPluginSettings(*pCtx.DataSourceInstanceSettings)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("failed to load http config: %v", err.Error()))
		}

		apiKey := config.Secrets.ApiKey
		if apiKey != "" {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		}
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadGateway, fmt.Sprintf("feldera error: %v", err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return backend.ErrDataResponse(
				backend.StatusBadGateway,
				fmt.Sprintf("err: query failed, status: %s", resp.Status))
		}
		errMsg := string(msg)
		
		return backend.ErrDataResponse(backend.StatusBadRequest, 
			fmt.Sprintf("err: query failed, status: '%s', error: %s", resp.Status, errMsg))
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal,
			fmt.Sprintf("err: query failed, status: '%s', error: %s", resp.Status, err.Error()))
	}


	jsonstr := "[" + strings.TrimSpace(string(contents))
	jsonstr = strings.ReplaceAll(jsonstr, "\n", ", ") + "]"

	frame, err := jsonframer.ToFrame(jsonstr, jsonframer.FramerOptions{})
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal,
			fmt.Sprintf("err: query failed, status: '%s', error: %s", resp.Status, err.Error()))
	}

	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)

	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		return res, nil
	}

	apiKey := config.Secrets.ApiKey
	if apiKey != "" {
	}

	url := fmt.Sprintf("%s/v0/pipelines", d.baseUrl)

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to create HTTP request"
		return res, nil
	}

	resp, err := d.client.Do(r)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Data source unavailable"
		return res, nil
	}

	if resp.StatusCode >= 400 || resp.StatusCode < 200 {
		res.Status = backend.HealthStatusError
		res.Message = "Invalid response from data source"
		return res, nil
	}


	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
