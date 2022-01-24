package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/geekgonecrazy/uberContainer/core"
	"github.com/geekgonecrazy/uberContainer/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func GetContainersHandler(c *gin.Context) {
	if valid := checkValidAuthentication("", c); !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	containers, err := core.GetContainers()
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, containers)
}

func ContainerDownloadHandler(c *gin.Context) {
	containerKey := c.Params.ByName("container_key")
	returnLink, _ := strconv.ParseBool(c.Query("r"))

	if valid := checkValidAuthentication(containerKey, c); !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	fileLink, err := core.GetContainerFileLink(containerKey)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if returnLink {
		c.JSON(http.StatusOK, gin.H{"downloadLink": fileLink})
	} else {
		c.Redirect(http.StatusTemporaryRedirect, fileLink)
	}
}

func ContainerPreviewHandler(c *gin.Context) {
	containerKey := c.Params.ByName("container_key")

	previewLink, err := core.GetContainerPreviewLink(containerKey)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, previewLink)
}

func ContainerCreateHandler(c *gin.Context) {
	form := models.ContainerCreateUpdatePayload{}

	c.Bind(&form)

	if valid := checkValidAuthentication(form.ContainerKey, c); !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	container, _ := core.GetContainer(form.ContainerKey)
	if container.Key != "" {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	fmt.Printf("%+v\n", form)
	if len(form.DownloadUrl) > 0 {

		container, err := core.ContainerFileUploadFromUrl(form)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(201, container)
	} else {
		log.Println("File upload")

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		container, err := core.ContainerFileUploadFromForm(form, header, file)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(201, container)
	}

}

func ContainerUpdateHandler(c *gin.Context) {
	containerKey := c.Params.ByName("container_key")

	if valid := checkValidAuthentication(containerKey, c); !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	form := models.ContainerCreateUpdatePayload{}

	c.BindWith(&form, binding.Form)

	fmt.Printf("%+v\n", form)

	form.ContainerKey = containerKey

	if err := core.DeleteContainerFile(containerKey); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(form.DownloadUrl) > 0 {

		container, err := core.ContainerFileUploadFromUrl(form)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
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
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(201, container)
	}
}

func GetContainerHandler(c *gin.Context) {
	containerKey := c.Params.ByName("container_key")

	if valid := checkValidAuthentication(containerKey, c); !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	container, err := core.GetContainer(containerKey)
	if err != nil {
		log.Println(err.Error())
		if err.Error() == "record not found" {
			c.JSON(404, gin.H{})
		} else {
			c.JSON(500, gin.H{})
		}

	}

	c.JSON(200, container)
}

func GetContainerMetaHandler(c *gin.Context) {
	containerKey := c.Params.ByName("container_key")

	if valid := checkValidAuthentication(containerKey, c); !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	container, err := core.GetContainer(containerKey)
	if err != nil {
		log.Println(err.Error())
		if err.Error() == "record not found" {
			c.JSON(404, gin.H{})
			return
		} else {
			c.JSON(500, gin.H{})
			return
		}
	}

	c.Writer.Header().Add("X-Uber-Container-Filename", container.Filename)
	c.Writer.Header().Add("X-Uber-Container-Filesize", fmt.Sprintf("%d", container.FileSize))
	c.Writer.Header().Add("Last-Modified", container.ModifiedAt.String())
	c.AbortWithStatus(http.StatusNoContent)
}

func ContainerDeleteFileHandler(c *gin.Context) {
	containerKey := c.Params.ByName("container_key")

	if valid := checkValidAuthentication(containerKey, c); !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err := core.DeleteContainerFile(containerKey); err != nil {

	}

	c.JSON(200, gin.H{})
}

func ContainerDeleteHandler(c *gin.Context) {
	containerKey := c.Params.ByName("container_key")

	if valid := checkValidAuthentication(containerKey, c); !valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err := core.DeleteContainer(containerKey); err != nil {

	}

	c.String(200, "")
}
