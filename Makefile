build: clean
	go build -a -tags netgo -ldflags '-w'

clean:
	rm -f solr-prometheus-exporter
