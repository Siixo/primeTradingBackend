package handler

import (
	"backend/internal/domain/model"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// CommodityServicePort defines the contract the handler depends on.
// This decouples the handler from the concrete *application.CommodityService.
type CommodityServicePort interface {
	GetCommodityByType(commodityType string) (*model.Commodity, error)
	GetHistory(name string, limit int) ([]model.Commodity, error)
	GetStatuses() ([]model.CommodityStatus, error)
}

type CommodityHandler struct {
	commodityService CommodityServicePort
}

func NewCommodityHandler(commodityService CommodityServicePort) *CommodityHandler {
	return &CommodityHandler{commodityService: commodityService}
}

func (h *CommodityHandler) GetCommodityHandler(w http.ResponseWriter, r *http.Request) {
	commodityType := r.URL.Query().Get("type")
	if commodityType == "" {
		jsonError(w, "'type' query parameter is required", http.StatusBadRequest)
		return
	}

	commodity, err := h.commodityService.GetCommodityByType(commodityType)
	if err != nil {
		if err.Error() == "unknown commodity type" {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(commodity); err != nil {
		jsonError(w, "failed to encode response", http.StatusInternalServerError)
	}
}
func (h *CommodityHandler) GetCommodityHistoryHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		jsonError(w, "commodity name is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	history, err := h.commodityService.GetHistory(name, limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure we return [] instead of null for empty results
	if history == nil {
		history = []model.Commodity{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		jsonError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *CommodityHandler) GetCommodityStatusHandler(w http.ResponseWriter, r *http.Request) {
	statuses, err := h.commodityService.GetStatuses()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		jsonError(w, "failed to encode response", http.StatusInternalServerError)
	}
}
