package router

import (
	"github.com/geekgonecrazy/uberContainer/controllers"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

func Start() {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	api := router.Group("/api")

	containers := api.Group("/containers")

	containers.GET("", controllers.GetContainersHandler)
	containers.POST("", controllers.ContainerCreateHandler)

	containers.GET("/*container_id", controllers.GetContainerHandler)
	containers.PUT("/*container_id", controllers.ContainerUpdateHandler)
	containers.POST("/*container_id", controllers.ContainerUpdateHandler)
	containers.DELETE("/*container_id", controllers.ContainerDeleteHandler)

	files := api.Group("/files")

	files.GET("/*container_id", controllers.ContainerDownloadHandler)
	files.DELETE("/*container_id", controllers.ContainerDeleteFileHandler)

	previews := api.Group("/previews")

	previews.GET("/*container_id", controllers.ContainerPreviewHandler)

	router.Use(static.Serve("/", static.LocalFile("app", true)))

	router.Run(":8080")
}
