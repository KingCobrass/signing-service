package api

import (
	"log"
	"net/http"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// swagger:route POST health
// Get health status
//
// responses:
//
//	405: Method not allowed
//	200: Success
//
// Health evaluates the health of the service and writes a standardized response.
func (s *Server) Health(response http.ResponseWriter, request *http.Request) {

	log.Printf("Got the Health API Request\n")

	if request.Method != http.MethodGet {
		WriteErrorResponse(response, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	health := HealthResponse{
		Status:  "OK",
		Version: "v1",
	}

	WriteAPIResponse(response, http.StatusOK, health)
}
