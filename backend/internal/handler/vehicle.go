package handler

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
	"net/http"
)

type VehicleHandler struct{ decoder *service.VINDecoder }

func NewVehicleHandler(decoder *service.VINDecoder) *VehicleHandler {
	return &VehicleHandler{decoder: decoder}
}
func (h *VehicleHandler) DecodeVIN(w http.ResponseWriter, r *http.Request) {
	vehicle, err := h.decoder.Decode(r.Context(), chi.URLParam(r, "vin"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrVINFormat):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrVINNotDecoded):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		default:
			http.Error(w, "VIN lookup is unavailable. Enter vehicle details manually or try again.", http.StatusBadGateway)
		}
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, vehicle)
}
