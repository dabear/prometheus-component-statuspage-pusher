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
  
```

The prometheus-expression needs to return a single element vector, like:
```
abcdef: avg(up{job="web"})
```
Values above 100 are considered as "partial outage"
Values below 100 are considered as "operational"
