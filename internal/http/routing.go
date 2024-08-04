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
	if config.BehindProxy {
		r.Use(middleware.RealIP)
	}
	if len(config.AllowedHosts) > 0 {
		r.Use(HostFilter(config.AllowedHosts...))
	}
	r.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.CleanPath,
		middleware.RequestID,
		middleware.RedirectSlashes,
		middleware.Recoverer,
		middleware.Timeout(60*time.Second),
	)
	r.Route(contextRoot, func(r chi.Router) {
		r.Route("/api/v1", func(r chi.Router) {
			r.Get("/devices", s.getDevices())
			r.Get("/devices/{class}", s.getDevices())
			r.Get("/devices/RollerShutters/close", s.rollerShutter("close"))
			r.Get("/devices/RollerShutters/open", s.rollerShutter("open"))
		})
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Interface, config.Port),
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

func (s *Server) getDevices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		class := chi.URLParam(r, "class")
		render.JSON(w, r, s.overkiz.Devices(class))
	}
}

func (s *Server) rollerShutter(actionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		deviceCount, _ := s.overkiz.RollerShutters(actionName)
		if deviceCount == 0 {
			w.WriteHeader(404)
			_, err := w.Write([]byte("{\"error\":\"No RollerShutters found\"}"))
			if err != nil {
				log.Errorf("Error writing response: %v", err)
			}
		} else {
			w.WriteHeader(202)
			_, err := w.Write([]byte("{\"status\":\"Executing\"}"))
			if err != nil {
				log.Errorf("Error writing response: %v", err)
			}
		}
	}
}
