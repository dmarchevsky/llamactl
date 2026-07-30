package server

import (
	"fmt"
	"net/http"
)

// VersionHandler godoc
// @Summary Get llamactl version
// @Description Returns the version of the llamactl command
// @Tags System
// @Security ApiKeyAuth
// @Produces text/plain
// @Success 200 {string} string "Version information"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/version [get]
func (h *Handler) VersionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		versionInfo := fmt.Sprintf("Version: %s\nCommit: %s\nBuild Time: %s\n", h.cfg.Version, h.cfg.CommitHash, h.cfg.BuildTime)
		writeText(w, http.StatusOK, versionInfo)
	}
}

// healthResponse represents the health status returned by HealthHandler
type healthResponse struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Database string `json:"database"`
	Instances struct {
		Running int `json:"running"`
		Total   int `json:"total"`
	} `json:"instances"`
}

// HealthHandler godoc
// @Summary Get server health status
// @Description Returns the health status of the llamactl server, database, and instances
// @Tags System
// @Produces application/json
// @Success 200 {object} healthResponse "Server is healthy"
// @Failure 503 {object} healthResponse "Server is unhealthy"
// @Router /health [get]
func (h *Handler) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := healthResponse{
			Status:   "ok",
			Version:  h.cfg.Version,
			Database: "ok",
		}

		if err := h.authStore.HealthCheck(); err != nil {
			resp.Status = "error"
			resp.Database = err.Error()
		}

		if instances, err := h.InstanceManager.ListInstances(); err == nil {
			resp.Instances.Total = len(instances)
			for _, inst := range instances {
				if inst.IsRunning() {
					resp.Instances.Running++
				}
			}
		}

		status := http.StatusOK
		if resp.Status != "ok" {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, resp)
	}
}

// ConfigHandler godoc
// @Summary Get server configuration
// @Description Returns the current server configuration (sanitized)
// @Tags System
// @Security ApiKeyAuth
// @Produces application/json
// @Success 200 {object} config.AppConfig "Sanitized configuration"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/config [get]
func (h *Handler) ConfigHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sanitizedConfig, err := h.cfg.SanitizedCopy()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "sanitized_copy_error", "Failed to get sanitized config")
			return
		}
		writeJSON(w, http.StatusOK, sanitizedConfig)
	}
}
