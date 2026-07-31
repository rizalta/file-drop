package handler

import "net/http"

type DropsService interface{}

type dropsHandler struct {
	dropsService DropsService
}

func NewHandler(s DropsService) *dropsHandler {
	return &dropsHandler{s}
}

func (h *dropsHandler) UploadDrop(w http.ResponseWriter, r *http.Request) {
}
