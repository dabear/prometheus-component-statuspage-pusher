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

	"gopkg.in/yaml.v2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)



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

	qcd, err := ioutil.ReadFile(*queryConfigFile)
	if err != nil {
		fatal("msg", "Couldn't read config file", "error", err.Error())
	}


	client, err := api.NewClient(api.Config{Address: *prometheusURL})
	if err != nil {
		fatal("msg", "Couldn't create Prometheus client", "error", err.Error())
	}
	api := prometheus.NewAPI(client)

	m := make(map[interface{}]map[interface{}]interface{})

  err2 :=yaml.Unmarshal([]byte(qcd), &m)
	if err2 != nil {
    fatal("error: %v", err2)
  }
	if *debugMode == true {
		fmt.Printf("%v\n", m)
	}


	for {

		for componentstatus, _ := range m {
		  //fmt.Printf("key[%s] value[%s]\n", componentstatus,entries)
			for componentID, query := range m[componentstatus] {
				if *debugMode == true {
					fmt.Printf("ComponentID: %s\nComponentstatus: %s\n", componentID, componentstatus)
					fmt.Printf("Statusquery: %s\n", query)
				}



				ts := time.Now()
 				resp,_, err := api.Query(context.Background(), query.(string), ts)
 				if err != nil {
 					level.Error(logger).Log("msg", "Couldn't query Prometheus", "error", err.Error())
 					continue
 				}
 				vec := resp.(model.Vector)
				l := vec.Len()
 				if l == 0 {
					fmt.Printf("Skipped sending status %s for component %s\n", componentstatus, componentID.(string))
					if *debugMode == true {
						fmt.Printf("Component query: %s", query.(string))
					}
					continue
				} else if l > 1{
					level.Error(logger).Log("msg", "Expected query to return single value", "samples", l)
					continue
				}

 				level.Info(logger).Log("metricID", componentID, "resp", vec[0].Value)
 				if err := sendComponentStatus(ts, componentID.(string), componentstatus.(string)	, float64(vec[0].Value)); err != nil {
 					level.Error(logger).Log("msg", "Couldn't send metric to Statuspage", "error", err.Error())
 					continue
 				}
			}

		 }

		 if *debugMode == true {
			 fmt.Printf("\n\n")
		 }
		 time.Sleep(*metricInterval)
   }

}

func sendComponentStatus(ts time.Time, componentID string, status string, value float64) error {

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
		fmt.Printf("postparams: %s\n", values.Encode())
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
			if *debugMode == true {
				fmt.Printf("Got response:  %s\n", respStr)
			}
			fmt.Printf("Setting status %s for component %s\n", status, componentID)


		}

	}
	return nil
}
