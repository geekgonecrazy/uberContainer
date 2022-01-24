package router

import (
	"github.com/geekgonecrazy/uberContainer/controllers"
	m "github.com/geekgonecrazy/uberContainer/router/middleware"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

func Start() {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.LoadHTMLGlob("templates/*")

	api := router.Group("/api")

	containers := api.Group("/containers")

	containers.GET("", controllers.GetContainersHandler)
	containers.POST("", controllers.ContainerCreateHandler)

	containers.GET("/*container_key", controllers.GetContainerHandler)
	containers.HEAD("/*container_key", controllers.GetContainerMetaHandler)

	containers.PUT("/*container_key", controllers.ContainerUpdateHandler)
	containers.POST("/*container_key", controllers.ContainerUpdateHandler)
	containers.DELETE("/*container_key", controllers.ContainerDeleteHandler)

	files := api.Group("/files")

	files.GET("/*container_key", controllers.ContainerDownloadHandler)
	files.DELETE("/*container_key", controllers.ContainerDeleteFileHandler)

	previews := api.Group("/previews")

	previews.GET("/*container_key", controllers.ContainerPreviewHandler)

	router.Use(static.Serve("/", static.LocalFile("web/public", true)))

	router.NoRoute(m.ShowIndex)
	router.Run(":8080")
}
