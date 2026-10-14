package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestHandlerServesAssetsAndVueRoutes(t *testing.T) {
	assets := fstest.MapFS{
		"index.html":       &fstest.MapFile{Data: []byte("<main>app</main>")},
		"assets/app.js":    &fstest.MapFile{Data: []byte("console.log('app')")},
		"assets/style.css": &fstest.MapFile{Data: []byte("body { color: blue }")},
	}
	router := gin.New()
	router.NoRoute(Handler(assets))

	tests := []struct {
		name       string
		path       string
		statusCode int
		body       string
	}{
		{name: "root", path: "/", statusCode: http.StatusOK, body: "<main>app</main>"},
		{name: "vue route", path: "/orders/42", statusCode: http.StatusOK, body: "<main>app</main>"},
		{name: "asset", path: "/assets/app.js", statusCode: http.StatusOK, body: "console.log('app')"},
		{name: "stylesheet", path: "/assets/style.css", statusCode: http.StatusOK, body: "body { color: blue }"},
		{name: "missing asset", path: "/assets/missing.js", statusCode: http.StatusNotFound},
		{name: "unknown API", path: "/api/unknown", statusCode: http.StatusNotFound, body: "not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(recorder, request)
			if recorder.Code != tt.statusCode {
				t.Fatalf("status = %d, want %d; body: %s", recorder.Code, tt.statusCode, recorder.Body.String())
			}
			if tt.body != "" && !strings.Contains(recorder.Body.String(), tt.body) {
				t.Fatalf("body %q does not contain %q", recorder.Body.String(), tt.body)
			}
		})
	}
}

func TestHandlerExplainsMissingFrontendBuild(t *testing.T) {
	router := gin.New()
	router.NoRoute(Handler(fstest.MapFS{".keep": &fstest.MapFile{Data: nil}}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(recorder.Body.String(), "npm run build") {
		t.Fatalf("missing build instruction in response: %q", recorder.Body.String())
	}
}

func TestHandlerRejectsNonReadMethods(t *testing.T) {
	router := gin.New()
	router.NoRoute(Handler(fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("app")}}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/orders", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
