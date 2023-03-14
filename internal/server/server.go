package server

import (
	"fmt"
	"internal/storage"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type gauge float64
type counter int64

type Server struct {
	chiRouter chi.Router
	Storage   *storage.MemStorage
}

func NewServer() *Server {
	return &Server{
		Storage:   storage.NewMemStorage(),
		chiRouter: chi.NewRouter(),
	}
}

func (s Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	return
}

func (s Server) listHandler(w http.ResponseWriter, r *http.Request) {
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

func (s Server) GetValueHandler(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.Path, "/")
	l := len(paths)

	if l >= 2 {
		var metricType, metricName string
		metricType, metricName = paths[(l-2)], paths[(l-1)]

		v, err := s.Storage.Get(metricName)
		if err == nil {
			switch metricType {
			case "gauge":
				if gaugeType(v) {
					w.Write([]byte(fmt.Sprintf("%f", v.(gauge))))
					return
				}
			case "counter":
				if counterType(v) {
					w.Write([]byte(fmt.Sprintf("%d", v.(counter))))
					return
				}
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
	return
}

func (s Server) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.Path, "/")
	l := len(paths)

	if l >= 3 {
		var metricType, metricName, metricValue string
		metricType, metricName, metricValue = paths[(l-3)], paths[(l-2)], paths[(l-1)]

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
	//router := chi.NewRouter()
	s.chiRouter.Route("/", func(router chi.Router) {
		s.chiRouter.Get("/", s.listHandler)
		s.chiRouter.Get("/value/{metricType}/{metricName}", s.GetValueHandler)
		s.chiRouter.Post("/update/{metricType}/{metricName}/{metricValue}", s.UpdateHandler)
	})
	http.ListenAndServe(addr, s.chiRouter)
}

func getObjectType(i interface{}) string {
	return fmt.Sprintf("%T", i)
}

func gaugeType(i interface{}) bool {
	if getObjectType(i) == "server.gauge" {
		return true
	}
	return false
}

func counterType(i interface{}) bool {
	if getObjectType(i) == "server.counter" {
		return true
	}
	return false
}
