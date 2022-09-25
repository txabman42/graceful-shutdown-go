package example

import (
	"testing"

	"github.com/gin-gonic/gin"
	gracefulShutdown "github.com/txabman42/graceful-shutdown-go"
)

func Test_Shutdown(t *testing.T) {
	shutdown := gracefulShutdown.NewGracefulShutdown()

	ginEngine := gin.Default()
	server := NewHTTPServer(ginEngine)
	shutdown.Register(server, gracefulShutdown.HIGH)

	shutdown.Run()
}
