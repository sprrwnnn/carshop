package handlers

import (
	"net/http"
	"time"
)

// @Summary Panic
// @Description Panics every time
// @Tags Quirky
// @Router /api/v1/quirky/panic [get]
func (h *Handler) PanicHandler(w http.ResponseWriter, req *http.Request) {
	panic("this is a panic")
}

// @Summary Slow
// @Description Executes a slow operation
// @Tags Quirky
// @Router /api/v1/quirky/slow [get]
func (h *Handler) SlowHandler(w http.ResponseWriter, req *http.Request) {
	time.Sleep(time.Second * 5)
	w.WriteHeader(http.StatusOK)
}
