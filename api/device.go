package api

import "net/http"

// TODO: REST endpoints ...
func (s *Server) Device(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		WriteErrorResponse(response, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	WriteAPIResponse(response, http.StatusOK, "Not completed Yet")
}
