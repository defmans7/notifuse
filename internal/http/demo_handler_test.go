package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Notifuse/notifuse/internal/service"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
)

// We only cover negative/edge flows to increase coverage for DemoHandler quickly.

func TestDemoHandler_MethodNotAllowed(t *testing.T) {
	mockSvc := &service.DemoService{}
	mockLogger := pkgmocks.NewMockLogger(nil)
	h := NewDemoHandler(mockSvc, mockLogger)

	req := httptest.NewRequest(http.MethodPost, "/api/demo.reset", nil)
	w := httptest.NewRecorder()
	h.handleResetDemo(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, w.Code)
	}
}
