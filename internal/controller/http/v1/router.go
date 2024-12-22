package v1

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"medods-tz/internal/service"
	"os"
)

func NewRouter(handler *echo.Echo, service *service.Service, logPath string) {
	handler.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}", "method":"${method}","uri":"${uri}", "status":${status},"error":"${error}"}` + "\n",
		Output: setLogsFile(logPath),
	}))
	handler.Use(middleware.Recover())
	//handler.GET("/swagger/*", echoSwagger.WrapHandler)

	v1 := handler.Group("/api/v1")
	{
		newAuthRoutes(v1.Group("/auth"), service.AuthService)
	}
}

func setLogsFile(logPath string) *os.File {
	file, err := os.OpenFile(fmt.Sprintf("%s/requests.log", logPath), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}
