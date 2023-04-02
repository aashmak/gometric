package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func (s HTTPServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
}

func (s HTTPServer) listHandler(w http.ResponseWriter, r *http.Request) {
	var varList string

	for _, metricName := range s.Storage.List() {
		v, err := s.Storage.Get(metricName)
		if err == nil {
			if gaugeType(v) {
				varList += fmt.Sprintf("%s (type: gauge): %f<br>\n", metricName, v.(gauge))
			} else if counterType(v) {
				varList += fmt.Sprintf("%s (type: counter): %d<br>\n", metricName, v.(counter))
			}
		}
	}

	fmt.Fprintf(w, "<html>\n<title>Metric Dump</title>\n"+
		"<body>\n<h2>Metric Dump</h2>\n"+
		"%s\n"+
		"</body>\n</html>", varList)
}

func (s HTTPServer) GetValueHandler(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") == "application/json" {
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("server: could not read request body: %s\n", err)
		}

		var metric Metrics
		if err := json.Unmarshal(reqBody, &metric); err != nil {
			log.Printf("Error: %s", err.Error())
			return
		}

		v, err := s.Storage.Get(metric.ID)
		if err == nil {
			switch metric.MType {
			case "gauge":
				if gaugeType(v) {
					v1 := v.(gauge)
					metric.Value = (*float64)(&v1)

					ret, err := json.Marshal(v)
					if err != nil {
						log.Printf("Error: %s", err.Error())
					}
					w.Write(ret)
					return
				}
			case "counter":
				if counterType(v) {
					v1 := v.(counter)
					metric.Delta = (*int64)(&v1)

					ret, err := json.Marshal(v)
					if err != nil {
						log.Printf("Error: %s", err.Error())
					}
					w.Write(ret)
					return
				}
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (s HTTPServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("server: could not read request body: %s\n", err)
	}

	var metric Metrics
	if err := json.Unmarshal(reqBody, &metric); err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	switch metric.MType {
	case "gauge":
		if metric.ID != "" && metric.Value != nil {
			err := s.Storage.Set(metric.ID, gauge(*metric.Value))
			if err == nil {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	case "counter":
		// get previous counter value
		prevCounter, err := s.Storage.Get(metric.ID)
		if err != nil {
			prevCounter = counter(0)
		}

		if metric.ID != "" && metric.Delta != nil {
			err = s.Storage.Set(metric.ID, counter(*metric.Delta)+prevCounter.(counter))
			if err == nil {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	}

	w.WriteHeader(http.StatusForbidden)
}
