package handler

import (
	"backend/internal/domain/model"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

type CorrelationServicePort interface {
	GetCorrelationByType(correlationType string) (*model.Correlation, error)
	GetHistory(commodityA, commodityB string, limit int) ([]*model.Correlation, error)
}

type CorrelationHandler struct {
	correlationService CorrelationServicePort
}

func NewCorrelationHandler(correlationService CorrelationServicePort) *CorrelationHandler {
	return &CorrelationHandler{correlationService: correlationService}
}

func (h *CorrelationHandler) GetCorrelationHandler(w http.ResponseWriter, r *http.Request) {
	correlationType := r.URL.Query().Get("type")
	if correlationType == "" {
		jsonError(w, "'type' query parameter is required", http.StatusBadRequest)
		return
	}

	correlation, err := h.correlationService.GetCorrelationByType(correlationType)
	if err != nil {
		if err.Error() == "unknown correlation type" {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sanitizeCorrelation(correlation)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(correlation); err != nil {
		jsonError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *CorrelationHandler) GetCorrelationHistoryHandler(w http.ResponseWriter, r *http.Request) {
	commodityA := r.URL.Query().Get("a")
	commodityB := r.URL.Query().Get("b")

	if commodityA == "" || commodityB == "" {
		jsonError(w, "'a' and 'b' query parameters are required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	history, err := h.correlationService.GetHistory(commodityA, commodityB, limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, c := range history {
		sanitizeCorrelation(c)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		jsonError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func sanitizeCorrelation(c *model.Correlation) {
	if c == nil {
		return
	}
	if math.IsNaN(c.PearsonR) || math.IsInf(c.PearsonR, 0) {
		c.PearsonR = 0
	}
	if math.IsNaN(c.SpearmanRho) || math.IsInf(c.SpearmanRho, 0) {
		c.SpearmanRho = 0
	}
}