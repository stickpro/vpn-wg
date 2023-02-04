package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"vpn-wg/internal/delivery/http/handlers"
)

type Router struct {
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Init() *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.initApi(router)

	return router
}

func (r *Router) initApi(router *gin.Engine) {
	handlerV1 := handlers.NewHandler()
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
