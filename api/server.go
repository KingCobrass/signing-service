package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// Response is the generic API response container.
type Response struct {
	Data interface{} `json:"data"`
}

// ErrorResponse is the generic error API response container.
type ErrorResponse struct {
	Errors []string `json:"errors"`
}

// Server manages HTTP requests and dispatches them to the appropriate services.
type Server struct {
	listenAddress string
	log           *log.Logger
}

// NewServer is a factory to instantiate a new Server.
func NewServer(listenAddress string) *Server {
	return &Server{
		listenAddress: listenAddress,
		// TODO: add services / further dependencies here ...
	}
}

// Run registers all HandlerFuncs for the existing HTTP routes and starts the Server.
func (s *Server) Run() error {
	mux := http.NewServeMux()

	opts := middleware.SwaggerUIOpts{SpecURL: "http://localhost:8080/swagger/swagger.json"}
	sh := middleware.SwaggerUI(opts, nil)
	mux.Handle("/docs", sh)

	// optsRedoc := middleware.RedocOpts{SpecURL: "http://localhost:8080/swagger/swagger2.json"}
	// shRedoc := middleware.Redoc(optsRedoc, nil)
	// mux.Handle("/redocs", shRedoc)

	mux.Handle("/api/v1/health", http.HandlerFunc(s.Health))

	// TODO: register further HandlerFuncs here ...
	mux.Handle("/api/v1/device/CreateSignatureDevice", http.HandlerFunc(s.CreateSignatureDevice))
	mux.Handle("/api/v1/device/SignTransaction", http.HandlerFunc(s.SignTransaction))

	return http.ListenAndServe(s.listenAddress, mux)
}

// WriteInternalError writes a default internal error message as an HTTP response.
func WriteInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
}

// WriteErrorResponse takes an HTTP status code and a slice of errors
// and writes those as an HTTP error response in a structured format.
func WriteErrorResponse(w http.ResponseWriter, code int, errors []string) {
	w.WriteHeader(code)

	errorResponse := ErrorResponse{
		Errors: errors,
	}

	bytes, err := json.Marshal(errorResponse)
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes)
}

// WriteAPIResponse takes an HTTP status code and a generic data struct
// and writes those as an HTTP response in a structured format.
func WriteAPIResponse(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)

	response := Response{
		Data: data,
	}

	bytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes)
}
