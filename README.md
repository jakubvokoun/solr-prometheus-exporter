# SOLR Prometheus exporter

## Requirements

- Go 1.13+
- Make

## Build

Simply run `make`.

## Options

- `-port` (default 2112)
- `-scrape-interval` (default 2 seconds)
- `-sorl-url` (default http://localhost:8983/solr/admin/info/system?wt=json)
