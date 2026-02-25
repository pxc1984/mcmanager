package manager

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mcmanager/internal/config"

	"github.com/gin-gonic/gin"
)

func TestUpdateHandlerUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{SecretToken: "secret"}
	mgr := New(cfg)

	called := false
	mgr.syncRepoFn = func() error { called = true; return nil }
	mgr.pluginDownloadFn = func() error { called = true; return nil }
	mgr.dataSyncFn = func() error { called = true; return nil }
	mgr.restartFn = func() error { called = true; return nil }

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/update", nil)

	mgr.UpdateHandler(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if called {
		t.Fatalf("expected no downstream calls on unauthorized request")
	}
}

func TestUpdateHandlerSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	mgr := New(cfg)

	mgr.syncRepoFn = func() error { return nil }
	mgr.pluginDownloadFn = func() error { return nil }
	mgr.dataSyncFn = func() error { return nil }

	restartCh := make(chan struct{})
	mgr.restartFn = func() error {
		close(restartCh)
		return nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/update", nil)

	mgr.UpdateHandler(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "update applied") {
		t.Fatalf("unexpected response body: %q", w.Body.String())
	}

	select {
	case <-restartCh:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected restart to be scheduled")
	}
}

func TestUpdateHandlerRepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	mgr := New(cfg)

	mgr.syncRepoFn = func() error { return errors.New("repo down") }
	mgr.pluginDownloadFn = func() error { t.Fatalf("plugin download should not run"); return nil }
	mgr.dataSyncFn = func() error { t.Fatalf("data sync should not run"); return nil }
	mgr.restartFn = func() error { t.Fatalf("restart should not run"); return nil }

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/update", nil)

	mgr.UpdateHandler(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "repo sync failed") {
		t.Fatalf("unexpected response body: %q", w.Body.String())
	}
}
