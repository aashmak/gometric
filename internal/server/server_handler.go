package server

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gometric/internal/logger"
	"gometric/internal/metrics"
	"gometric/internal/postgres"
)

// defaultHandler стандарный handler, возвращает статус http.StatusForbidden.
func (s HTTPServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
}

// listHandler выводит все существующие метрики в виде html.
func (s HTTPServer) listHandler(w http.ResponseWriter, r *http.Request) {
	var varList string

	for _, metricName := range s.Storage.List() {
		v, err := s.Storage.Get(metricName)
		if err == nil {
			if gaugeType(v) {
				varList += fmt.Sprintf("%s (type: gauge): %f<br>\n", metricName, v.(float64))
			} else if counterType(v) {
				varList += fmt.Sprintf("%s (type: counter): %d<br>\n", metricName, v.(int64))
			}
		}
	}

	w.Header().Set("Content-Type", "text/html")

	fmt.Fprintf(w, "<html>\n<title>Metric Dump</title>\n"+
		"<body>\n<h2>Metric Dump</h2>\n"+
		"%s\n"+
		"</body>\n</html>", varList)
}

// GetValueHandler извлекает метрики из key-value бэкенда и отсылает в формате json.
// Функция также подписывает сообщение перед отправкой с помощью функции Sign().
func (s HTTPServer) GetValueHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("server could not read request body", err)
	}
	logger.Debug("request Body: " + string(reqBody))

	var metric metrics.Metrics
	if err = json.Unmarshal(reqBody, &metric); err != nil {
		logger.Error("", err)
		return
	}
	logger.Debug(fmt.Sprintf("unmarshall succefull: %v", metric))

	v, err := s.Storage.Get(metric.ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metric.MType {
	case "gauge":
		if gaugeType(v) {
			v1 := v.(float64)
			metric.Value = (*float64)(&v1)

			// sign if key is not empty
			if s.KeySign != "" {
				metric.Sign(s.KeySign)
			}

			ret, err := json.Marshal(metric)
			if err != nil {
				logger.Error("", err)
				return
			}
			logger.Debug(fmt.Sprintf("marshall succefull: %s", ret))

			w.Header().Set("Content-Type", "application/json")
			w.Write(ret)
			return
		}
	case "counter":
		if counterType(v) {
			v1 := v.(int64)
			metric.Delta = (*int64)(&v1)

			// sign if key is not empty
			if s.KeySign != "" {
				metric.Sign(s.KeySign)
			}

			ret, err := json.Marshal(metric)
			if err != nil {
				logger.Error("", err)
				return
			}
			logger.Debug(fmt.Sprintf("marshall succefull: %s", ret))

			w.Header().Set("Content-Type", "application/json")
			w.Write(ret)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

// UpdateHandler принимает метрики в формате json и сохраняет в key-value бэкенд.
// Функция также проверяет подпись с помощью ValidMAC().
func (s HTTPServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("server could not read request body", err)
	}

	var metric metrics.Metrics
	if err := json.Unmarshal(reqBody, &metric); err != nil {
		logger.Error("", err)
		return
	}

	// ValidMAC if key is not epmty
	if s.KeySign != "" {
		if !metric.ValidMAC(s.KeySign) {
			logger.Debug("invalid HMAC of the data")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	switch metric.MType {
	case "gauge":
		if metric.ID != "" && metric.Value != nil {
			err := s.Storage.Set(metric.ID, float64(*metric.Value))
			if err == nil {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	case "counter":
		// get previous counter value
		prevCounter, err := s.Storage.Get(metric.ID)
		if err != nil {
			prevCounter = int64(0)
		}

		if metric.ID != "" && metric.Delta != nil {
			err = s.Storage.Set(metric.ID, (*metric.Delta + prevCounter.(int64)))
			if err == nil {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	}

	logger.Debug("response status is Forbidden")
	w.WriteHeader(http.StatusForbidden)
}

func (s HTTPServer) UpdatesHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("server could not read request body", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var metrics []metrics.Metrics
	if err = json.Unmarshal(reqBody, &metrics); err != nil {
		logger.Error("", err)
		return
	}

	data := make(map[string]interface{})

	for _, metric := range metrics {
		// ValidMAC if key is not epmty
		if s.KeySign != "" {
			if !metric.ValidMAC(s.KeySign) {
				logger.Debug("invalid HMAC of the data")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		switch metric.MType {
		case "gauge":
			if metric.ID != "" && metric.Value != nil {
				data[metric.ID] = float64(*metric.Value)
			}

		case "counter":
			// get previous counter value
			var prevCounter interface{}
			prevCounter, err = s.Storage.Get(metric.ID)
			if err != nil {
				prevCounter = int64(0)
			}

			if metric.ID != "" && metric.Delta != nil {
				data[metric.ID] = (*metric.Delta + prevCounter.(int64))
			}
		}
	}

	err = s.Storage.MSet(data)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	logger.Debug("response status is Forbidden")
	w.WriteHeader(http.StatusForbidden)
}

// unzipBodyHandler используется для распаковки сжатого с помощью gzip тела сообщения.
func unzipBodyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncodingValues := r.Header.Values("Content-Encoding")

		if contentEncodingContains(contentEncodingValues, "gzip") {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}
			defer reader.Close()

			r.Body = io.NopCloser(reader)
		}

		next.ServeHTTP(w, r)
	})
}

// pingHandler используется для проверки доступности БД.
// Используется только с бэкендом Postgres.
func (s HTTPServer) pingHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.Storage.(*postgres.Postgres); ok {
		if err := s.Storage.(*postgres.Postgres).Ping(); err == nil {
			logger.Debug("database is reachable")
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	logger.Debug("database is unreachable")
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}
