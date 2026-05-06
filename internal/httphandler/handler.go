// Package httphandler provides the HTTP convenience handler for CodeValdDT.
//
// Currently exposes one route:
//
//	GET /{agencyId}/dt/schema/dtdl — DTDL v3 export of the active schema.
package httphandler

import (
	"errors"
	"net/http"
	"strings"

	codevalddt "github.com/aosanya/CodeValdDT"
	"github.com/aosanya/CodeValdDT/internal/dtdl"
	"github.com/aosanya/CodeValdSharedLib/entitygraph"
)

// Handler serves CodeValdDT HTTP convenience routes. Construct via New and
// register as an http.Handler on the cmux HTTP listener.
type Handler struct {
	sm codevalddt.DTSchemaManager
}

// New returns an initialised Handler.
func New(sm codevalddt.DTSchemaManager) *Handler {
	return &Handler{sm: sm}
}

// ServeHTTP routes incoming requests.
//
//	GET /{agencyId}/dt/schema/dtdl — DTDL v3 export
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Path format: /{agencyId}/dt/schema/dtdl
	// SplitN("/agency-1/dt/schema/dtdl", "/", 4) → ["", "agency-1", "dt", "schema/dtdl"]
	parts := strings.SplitN(r.URL.Path, "/", 4)
	if len(parts) == 4 && parts[2] == "dt" && parts[3] == "schema/dtdl" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.handleExportDTDL(w, r, parts[1])
		return
	}

	http.NotFound(w, r)
}

func (h *Handler) handleExportDTDL(w http.ResponseWriter, r *http.Request, agencyID string) {
	if agencyID == "" {
		http.Error(w, "missing agencyId", http.StatusBadRequest)
		return
	}

	schema, err := h.sm.GetActive(r.Context(), agencyID)
	if err != nil {
		if errors.Is(err, entitygraph.ErrSchemaNotFound) {
			http.Error(w, "schema not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data, err := dtdl.ExportSchema(agencyID, schema)
	if err != nil {
		http.Error(w, "export error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}
