package example

import (
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type HTTPServer struct {
	server *http.Server
	engine *gin.Engine
}

func NewHTTPServer(engine *gin.Engine) *HTTPServer {
	return &HTTPServer{engine: engine}
}

func (s *HTTPServer) Start() error {
	s.server = &http.Server{Handler: s.engine}
	ln, err := s.listen()
	if err != nil {
		return err
	}

	go func() {
		if err = s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Panic(err)
		}
	}()
	return nil
}

func (s *HTTPServer) listen() (net.Listener, error) {
	addr := s.server.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return ln, nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
