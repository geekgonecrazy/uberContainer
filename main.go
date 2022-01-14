package main

import (
	"bytes"
	"crypto/tls"

	//"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"

	"github.com/geekgonecrazy/uberContainer/models"
	"github.com/geekgonecrazy/uberContainer/store"
	"github.com/geekgonecrazy/uberContainer/store/boltdb"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/mgo.v2/bson"
)

var (
	containerDirectory = "/Volumes/Containers"
)

var _store store.Store

func HomeHandler(c *gin.Context) {
	c.String(200, "Hail Hydra!")
}

func getFileHash(container_id string, filename string) string {
	//sha1 := checksum.File(path.Join(containerDirectory, container_id, filename), crypto.SHA1)
	//log.Println(sha1)

	return "sha1"
}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Println(err)
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		log.Println(err)
	}

	log.Println(image)
	return image.Width, image.Height
}

func downloadFile(container_id string, download_url string, filename string) {
	log.Println("Download Url Passed")
	log.Println("URL: " + download_url)
	log.Println("Filename: " + filename)

	err := os.Mkdir(path.Join(containerDirectory, container_id), 0777)
	if err != nil {
		log.Println(err)
	}

	out, err := os.Create(path.Join(containerDirectory, container_id, filename))
	defer out.Close()
	if err != nil {
		log.Println(err)
	}

	u, err := url.Parse(download_url)
	if err != nil {
		log.Println(err)
	}

	if u.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client := &http.Client{Transport: tr}

		resp, err := client.Get(download_url)
		defer resp.Body.Close()
		if err != nil {
			log.Println(err)
		}

		log.Println(resp.Status)

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Println(err)
		}
	} else {
		resp, err := http.Get(download_url)
		defer resp.Body.Close()
		if err != nil {
			log.Println(err)
		}

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Println(err)
		}
	}

	_, err = os.Stat(path.Join(containerDirectory, container_id, filename))
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Finished Downloading")
	}
}

func cleanThumbnails(container_id string) {
	log.Println("Removing all thumbnails from container: " + container_id)

	directory := path.Join(containerDirectory, container_id)

	d, err := os.Open(directory)
	if err != nil {
		log.Println(err)
	}

	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		log.Println(err)
	}

	r, err := regexp.Compile("preview-[0-9]+.png")
	if err != nil {
		log.Println(err)
	}

	for _, file := range files {
		if r.MatchString(file.Name()) {
			err := os.Remove(directory + "/" + file.Name())

			if err != nil {
				log.Println(err)
			}
		}
	}
}

func generateThumbnail(container_id string, size string) (string, error) {

	container, err := _store.GetContainer(container_id)
	if err != nil {
		log.Println(err)
	}

	filePath := path.Join(containerDirectory, container_id, container.Filename)
	thumbPath := path.Join(containerDirectory, container_id, "preview-"+size+".png")

	log.Println(size + " x " + size + " Thumbnail has been requested for container: " + container_id)

	out, err := os.Stat(thumbPath)
	if err != nil {
		log.Println(out)
		log.Println("Preview image needs updated or Created")

		fileExt := filepath.Ext(filePath)

		if fileExt == ".psd" {
			filePath += "[0]"
		}

		_, err := exec.Command("/usr/local/bin/convert", filePath, "-thumbnail", size+"x"+size, thumbPath).Output()
		if err != nil {
			log.Println(err)
		}

		log.Println(thumbPath)
	}

	return thumbPath, nil
}

func ContainerDownloadHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")

	container, err := _store.GetContainer(container_id)
	if err != nil {
		log.Println(err)
	}

	log.Println(container)

	filePath := path.Join(containerDirectory, container_id, container.Filename)

	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+container.Filename)

	fileExt := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(fileExt)

	log.Println(fileExt + "   ----    " + mimeType)

	c.Writer.Header().Set("Content-Type", mimeType)

	http.ServeFile(c.Writer, c.Request, filePath)
}

func ContainerPreviewHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")

	default_size := "900"

	redirect_url := "/containers/" + container_id + "/preview/" + default_size

	http.Redirect(c.Writer, c.Request, redirect_url, 302)
}

func ContainerThumbnailHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")
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
	}
}

func ContainerUpdateHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")

	form := models.ContainerCreateUpdatePayload{}

	c.BindWith(&form, binding.Form)

	cleanThumbnails(container_id)

	container, err := _store.GetContainer(container_id)
	if err != nil {
		log.Println(err)
	}

	err = os.Mkdir(path.Join(containerDirectory, container_id), 0777)
	if err != nil {
		log.Println(err)
	}

	if len(form.DownloadUrl) > 0 {
		downloadFile(container_id, form.DownloadUrl, form.Filename)

		fileHash := getFileHash(container_id, form.Filename)

		filePath := path.Join(containerDirectory, container_id, form.Filename)
		fileExt := filepath.Ext(filePath)
		mimeType := mime.TypeByExtension(fileExt)

		container.MimeType = mimeType
		container.FileHash = fileHash

		_store.UpdateContainer(&container)
	} else {

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			log.Println(err)
		}

		log.Println(header.Filename)

		if len(container.Filename) > 0 {
			err = os.Remove(path.Join(containerDirectory, container_id, container.Filename))
			if err != nil {
				log.Println(err)
			}
		}

		out, err := os.Create(path.Join(containerDirectory, container_id, header.Filename))
		if err != nil {
			log.Println(err)
		}

		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			log.Println(err)
		}

		fileHash := getFileHash(container_id, header.Filename)
		filePath := path.Join(containerDirectory, container_id, header.Filename)
		fileExt := filepath.Ext(filePath)
		mimeType := mime.TypeByExtension(fileExt)

		container.FileHash = fileHash
		container.MimeType = mimeType

		_store.UpdateContainer(&container)
	}

	if len(form.Callback) > 0 {
		log.Println("Callback: " + form.Callback)
		resp, err := http.Post(form.Callback, "application/json", bytes.NewReader([]byte(`{"container_id": "`+container_id+`"}`)))
		defer resp.Body.Close()
		if err != nil {
			log.Println(err)
		}

	}

	c.JSON(200, gin.H{"container_id": container_id})
}

func GetContainerHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")

	container, err := _store.GetContainer(container_id)
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

	container, err := _store.GetContainer(container_id)
	if err != nil {
		log.Println(err)
	}

	log.Println("DELETE: " + container_id)

	err = os.Remove(path.Join(containerDirectory, container_id, container.Filename))
	if err != nil {
		log.Println(err)
	}

	cleanThumbnails(container_id)

	container.Empty = true
	container.FileHash = ""
	container.MimeType = ""
	container.Width = 0
	container.Height = 0
	container.Filename = ""

	_store.UpdateContainer(&container)

	c.JSON(200, gin.H{})
}

func ContainerCreateHandler(c *gin.Context) {
	log.Println("Creating New Container..")

	container_id := bson.NewObjectId().Hex()

	form := models.ContainerCreateUpdatePayload{}

	c.Bind(&form)

	fmt.Printf("%+v\n", form)
	if len(form.DownloadUrl) > 0 {

		log.Println(form.DownloadUrl)

		downloadFile(form.ContainerKey, form.DownloadUrl, form.Filename)

		fileHash := getFileHash(form.ContainerKey, form.Filename)

		filePath := path.Join(containerDirectory, form.ContainerKey, form.Filename)
		fileExt := filepath.Ext(filePath)
		mimeType := mime.TypeByExtension(fileExt)

		container := models.Container{
			Key:      form.ContainerKey,
			Filename: form.Filename,
			FileHash: fileHash,
			Empty:    false,
			MimeType: mimeType,
			Height:   0,
			Width:    0,
		}

		_store.CreateContainer(&container)

		if len(form.Callback) > 0 {
			resp, err := http.Post(form.Callback, "application/json", bytes.NewReader([]byte(`{"container_id": "`+container.Key+`"}`)))
			defer resp.Body.Close()
			if err != nil {
				log.Println(err)
			}

			c.String(201, "Done!")
		} else {
			c.JSON(201, container)
		}
	} else {
		log.Println("File upload")

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			log.Println(err)
		}

		log.Println(header.Filename)

		err = os.Mkdir(path.Join(containerDirectory, container_id), 0777)
		if err != nil {
			log.Println(err)
		}

		out, err := os.Create(path.Join(containerDirectory, container_id, header.Filename))
		if err != nil {
			log.Println(err)
		}

		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			log.Println(err)
		}

		log.Println("Finished Uploading")

		fileHash := getFileHash(container_id, header.Filename)

		filePath := path.Join(containerDirectory, container_id, header.Filename)
		fileExt := filepath.Ext(filePath)
		mimeType := mime.TypeByExtension(fileExt)

		container := models.Container{
			Key:      form.ContainerKey,
			Filename: header.Filename,
			FileHash: fileHash,
			Empty:    false,
			MimeType: mimeType,
			Height:   0,
			Width:    0,
		}

		_store.CreateContainer(&container)

		if len(form.Callback) > 0 {
			log.Println("Callback: " + form.Callback)
			resp, err := http.Post(form.Callback, "application/json", bytes.NewReader([]byte(`{"container_id": "`+container_id+`"}`)))
			defer resp.Body.Close()
			if err != nil {
				log.Println(err)
			}
		}

		c.JSON(201, container)
	}

}

func ContainerDeleteHandler(c *gin.Context) {
	container_id := c.Params.ByName("container_id")

	log.Println("DELETE: " + container_id)

	err := os.RemoveAll(path.Join(containerDirectory, container_id))
	if err != nil {
		log.Println(err)
	}

	_store.DeleteContainer(container_id)

	c.String(200, "")
}

func TestHandler(c *gin.Context) {
	key := c.Params.ByName("key")

	log.Println("test", key)

	c.String(201, key)
}

func main() {

	//connectionString := "mongodb://localhost:27017/uber"

	/*mongoStore, err := mongo.New(connectionString)
	if err != nil {
		panic(err)
	}

	_store = mongoStore*/

	boltStore, err := boltdb.New("./bolt.db")
	if err != nil {
		panic(err)
	}

	_store = boltStore

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	api := router.Group("/api")

	api.GET("/bob/*key", TestHandler)

	containers := api.Group("/containers")

	containers.GET("", GetContainerHandler)
	containers.POST("", ContainerCreateHandler)

	containers.GET("/:container_id", GetContainerHandler)
	containers.PUT("/:container_id", ContainerUpdateHandler)
	containers.POST("/:container_id", ContainerUpdateHandler)

	containers.DELETE("/:container_id", ContainerDeleteHandler)

	containers.GET("/:container_id/file", ContainerDownloadHandler)
	containers.DELETE("/:container_id/file", ContainerDeleteFileHandler)

	containers.GET("/:container_id/preview", ContainerPreviewHandler)
	containers.GET("/:container_id/preview/:size", ContainerThumbnailHandler)

	router.Use(static.Serve("/", static.LocalFile("app", false)))

	router.Run(":8080")
}
