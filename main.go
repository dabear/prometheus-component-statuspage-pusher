package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type queryConfig map[string]string

var (
	prometheusURL   = flag.String("pu", "http://localhost:9090", "URL of Prometheus server")
	statusPageURL   = flag.String("su", "https://api.statuspage.io", "URL of Statuspage API")
	statusPageToken = flag.String("st", "", "Statuspage Oauth token")
	statusPageID    = flag.String("si", "", "Statuspage page ID")
	queryConfigFile = flag.String("c", "queries.yaml", "Query config file")
	metricInterval  = flag.Duration("i", 30*time.Second, "Metric push interval")
	debugMode = flag.Bool("debug", false, "run in debug mode")
	logger     = log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)), "caller", log.DefaultCaller)
	httpClient = &http.Client{}
)

func fatal(fields ...interface{}) {
	level.Error(logger).Log(fields...)
	os.Exit(1)
}
func main() {
	flag.Parse()
	qConfig := queryConfig{}
	qcd, err := ioutil.ReadFile(*queryConfigFile)
	if err != nil {
		fatal("msg", "Couldn't read config file", "error", err.Error())
	}
	if err := yaml.Unmarshal(qcd, &qConfig); err != nil {
		fatal("msg", "Couldn't parse config file", "error", err.Error())
	}

	client, err := api.NewClient(api.Config{Address: *prometheusURL})
	if err != nil {
		fatal("msg", "Couldn't create Prometheus client", "error", err.Error())
	}
	api := prometheus.NewAPI(client)

	for {
		for componentID, query := range qConfig {
			ts := time.Now()
			resp,_, err := api.Query(context.Background(), query, ts)
			if err != nil {
				level.Error(logger).Log("msg", "Couldn't query Prometheus", "error", err.Error())
				continue
			}
			vec := resp.(model.Vector)
			if l := vec.Len(); l != 1 {
				level.Error(logger).Log("msg", "Expected query to return single value", "samples", l)
				continue
			}

			level.Info(logger).Log("metricID", componentID, "resp", vec[0].Value)
			if err := sendComponentStatus(ts, componentID, float64(vec[0].Value)); err != nil {
				level.Error(logger).Log("msg", "Couldn't send metric to Statuspage", "error", err.Error())
				continue
			}
		}
		time.Sleep(*metricInterval)
	}
}

func sendComponentStatus(ts time.Time, componentID string, value float64) error {
	status := "operational"
	if value > 100 {
		status = "partial_outage"
	}
	values := url.Values{
		"component[status]": []string{status},

	}
	url := *statusPageURL + path.Join("/v1", "pages", *statusPageID, "components", componentID+ ".json")
	req, err := http.NewRequest("PATCH", url, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	if *debugMode == true {
		fmt.Printf("statuspagetoken: %s\n", *statusPageToken)
		fmt.Printf("url: %s\n", url)
		fmt.Printf("values: %s\n", values.Encode())
	}
	req.Header.Set("Authorization", "OAuth "+*statusPageToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respStr, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Empty API Error")
		}
		return errors.New("API Error: " + string(respStr))
	} else {
		respStr, err := ioutil.ReadAll(resp.Body)
		if err == nil{
			fmt.Printf("Got success: %s", respStr)
		}

	}
	return nil
}
