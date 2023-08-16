package server

import (
	"internal/storage"
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

type Server struct {
	Storage *storage.MemStorage
}

func NewServer() *Server {
	return &Server{
		Storage: storage.NewMemStorage(),
	}
}

func (s Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	return
}

func (s Server) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	paths := strings.Split(r.URL.Path, "/")

	if len(paths) == 5 {
		var metricType, metricName, metricValue string
		metricType = paths[2]
		metricName = paths[3]
		metricValue = paths[4]

		switch metricType {
		case "gauge":
			f, _ := strconv.ParseFloat(metricValue, 64)
			err := s.Storage.Set(metricName, gauge(f))
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
			}
		case "counter":
			c, _ := strconv.ParseInt(metricValue, 0, 64)

			// get previous counter value
			prev_counter, err := s.Storage.Get(metricName)
			if err != nil {
				prev_counter = counter(0)
			}
			err = s.Storage.Set(metricName, counter(c)+prev_counter.(counter))
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
			}
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusForbidden)
	return
}

func (s Server) ListenAndServe(addr string) {
	http.HandleFunc("/update/", s.UpdateHandler)
	http.HandleFunc("/", s.defaultHandler)
	http.ListenAndServe(addr, nil)
}
