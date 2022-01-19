package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/geekgonecrazy/uberContainer/core"
	"github.com/geekgonecrazy/uberContainer/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func GetContainersHandler(c *gin.Context) {
	containers, err := core.GetContainers()
	if err != nil {

	}

	c.JSON(http.StatusOK, containers)
}

func ContainerDownloadHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")

	fileLink, err := core.GetContainerFileLink(container_id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, fileLink)
}

func ContainerPreviewHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")

	default_size := "900"

	redirect_url := "/containers/" + container_id + "/preview/" + default_size

	http.Redirect(c.Writer, c.Request, redirect_url, 302)
}

func ContainerThumbnailHandler(c *gin.Context) {
	/*container_id := c.Params.ByName("container_id")
	size := c.Params.ByName("size")

	thumbPath, err := generateThumbnail(container_id, size)
	if err != nil {
		log.Println(err)
	}

	c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Writer.Header().Set("Pragma", "no-cache")
	c.Writer.Header().Set("Expires", "0")

	_, err = os.Stat(thumbPath)
	if err != nil {
		log.Println(err)
		c.String(404, "hello!")
	} else {
		http.ServeFile(c.Writer, c.Request, thumbPath)
	}*/
}

func ContainerCreateHandler(c *gin.Context) {
	form := models.ContainerCreateUpdatePayload{}

	c.Bind(&form)

	container, _ := core.GetContainer(form.ContainerKey)
	if container.Key != "" {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	fmt.Printf("%+v\n", form)
	if len(form.DownloadUrl) > 0 {

		container, err := core.ContainerFileUploadFromUrl(form)
		if err != nil {

		}

		c.JSON(201, container)
	} else {
		log.Println("File upload")

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			log.Println(err)
		}

		container, err := core.ContainerFileUploadFromForm(form, header, file)
		if err != nil {
			log.Fatalln(err)
		}

		c.JSON(201, container)
	}

}

func ContainerUpdateHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")
	form := models.ContainerCreateUpdatePayload{}

	c.BindWith(&form, binding.Form)

	fmt.Printf("%+v\n", form)

	form.ContainerKey = container_id

	if err := core.DeleteContainerFile(container_id); err != nil {

	}

	if len(form.DownloadUrl) > 0 {

		container, err := core.ContainerFileUploadFromUrl(form)
		if err != nil {

		}

		c.JSON(201, container)
	} else {
		log.Println("File upload")

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			log.Println(err)
		}

		container, err := core.ContainerFileUploadFromForm(form, header, file)
		if err != nil {

		}

		c.JSON(201, container)
	}
}

func GetContainerHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")

	container, err := core.GetContainer(container_id)
	if err != nil {
		log.Println(err.Error())
		if err.Error() == "not found" {
			c.JSON(404, gin.H{})
		} else {
			c.JSON(500, gin.H{})
		}

	} else {
		c.JSON(200, container)
	}
}

func ContainerDeleteFileHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")

	if err := core.DeleteContainerFile(container_id); err != nil {

	}

	c.JSON(200, gin.H{})
}

func ContainerDeleteHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")

	if err := core.DeleteContainer(container_id); err != nil {

	}

	c.String(200, "")
}
