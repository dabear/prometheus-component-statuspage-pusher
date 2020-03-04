# prometheus-statuspage-pusher

# Usage
```
Usage of ./prometheus-statuspage-pusher:
  -c string
    	Query config file (default "queries.yaml")
  -i duration
    	Components push interval (default 30s)
  -pu string
    	URL of Prometheus API (default "http://localhost:9091/prometheus")
  -si string
    	Statuspage page ID
  -st string
    	Statuspage Oauth token
  -su string
    	URL of Statuspage API (default "https://api.statuspage.io")
  -debug=false|true
        Whether or not to run in debug mode
```

## Config:
Syntax:
```
somestatus:
  componentID: prometheus-expression
  componentID: promethues-expression
somestatus:
  componentID: prometheus-expression
  componentID: promethues-expression  
```

The prometheus-expression needs to return a single element vector, like:
```
abcdef: avg(up{job="web"})
```
componentID refers to a component's ID as defined in your statuspage.io management interface.

somestatus are placeholders and refers to statuspage.io componentstatus, e.g. one of "operational","under_maintenance", "degraded_performance", "partial_outage", "major_outage" or ""
