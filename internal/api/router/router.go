package router

import (
	"fmt"
	"net/http"

	"github.com/askthechain/atcutils/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	ForYou = "/v1/forYou"
	Items  = "/v1/items"
)

type magicLog struct {
	logger atclogger.Logger
}

func (m magicLog) Write(args []byte) (int, error) {
	m.logger.Error(fmt.Sprint(string(args)), nil)

	return len(args), nil
}

type Handler interface {
	ListForYou(c *gin.Context)
	ListItems(c *gin.Context)
}

func New(
	logger atclogger.Logger,
	handler Handler,
) (http.Handler, error) {

	gin.SetMode(gin.DebugMode)
	//gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(CORSMiddleware())

	router.RedirectTrailingSlash = true
	router.Use(gin.RecoveryWithWriter(magicLog{logger: logger}))

	router.GET(ForYou, handler.ListForYou)
	router.GET(Items, handler.ListItems)

	return router, nil
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "x-access-token, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
