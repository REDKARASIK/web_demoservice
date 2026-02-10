package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"web_demoservice/internal/domain"
	"web_demoservice/internal/transport/http/v1/dto"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type OrderService interface {
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error)
}

type OrderHandler struct {
	service OrderService
}

func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := mux.Vars(r)["order_id"]
	if orderID == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)
		return
	}

	uuid, err := uuid.Parse(orderID)
	if err != nil {
		http.Error(w, "invalid order_id", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrder(r.Context(), uuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	orderDTO := dto.MapToOrderDTO(order)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(orderDTO); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
