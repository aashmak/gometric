module gometric

go 1.18

require internal/metric v1.0.0

replace internal/metric => ./internal/metric

require internal/server v1.0.0

replace internal/server => ./internal/server

require internal/storage v1.0.0

require github.com/go-chi/chi/v5 v5.0.8 // indirect

replace internal/storage => ./internal/storage
