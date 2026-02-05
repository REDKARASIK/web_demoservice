package router

import (
	"net/http"
	"web_demoservice/internal/transport/http/v1/handlers"

	"github.com/gorilla/mux"
)

func RegisterOrderRoutes(r *mux.Router, handler *handlers.OrderHandler) {
	or := r.PathPrefix("/order").Subrouter()
	or.HandleFunc("/{order_id}", handler.GetOrder).Methods(http.MethodGet)
}
