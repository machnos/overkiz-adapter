package http

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"net/http"
	"overkiz-adapter/internal/config"
	"overkiz-adapter/internal/domain"
	"overkiz-adapter/internal/log"
	"strings"
	"time"
)

type Server struct {
	server  *http.Server
	overkiz *domain.Overkiz
}

func NewServer(config *config.Http, overkiz *domain.Overkiz) (*Server, error) {
	s := &Server{
		overkiz: overkiz,
	}

	contextRoot := config.ContextRoot
	if contextRoot == "" {
		contextRoot = "/"
	} else {
		if !strings.HasPrefix(contextRoot, "/") {
			contextRoot = "/" + contextRoot
		}
		if len(contextRoot) > 1 && strings.HasSuffix(contextRoot, "/") {
			contextRoot = contextRoot[0 : len(contextRoot)-1]
		}
	}

	r := chi.NewRouter()
	r.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.CleanPath,
		middleware.RequestID,
		middleware.RealIP,
		middleware.RedirectSlashes,
		middleware.Recoverer,
		middleware.Timeout(60*time.Second),
	)

	r.Route(contextRoot, func(r chi.Router) {
		r.Route("/api/v1", func(r chi.Router) {
			r.Get("/shutters", s.getShutters())
		})
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: r,
	}
	return s, nil
}

func (s *Server) Start() error {
	log.Infof("Starting http server at %v", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Info("Shutting down http server")
	return s.server.Shutdown(ctx)
}

func (s *Server) getShutters() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, fmt.Sprintf("Hello, %s", "Mark"))
	}
}
