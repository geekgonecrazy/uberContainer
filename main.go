package main

import (
	"bytes"
	"crypto"
	"crypto/tls"
	//"encoding/json"
	"fmt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"image"
	_ "image/png"
	"io"
	"log"
	"menteslibres.net/gosexy/checksum"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
)

var (
	mgoSession         *mgo.Session
	databaseName       = os.Getenv("DATABASE")
	databaseHost       = os.Getenv("HOST")
	databaseUser       = os.Getenv("USER")
	databasePasswd     = os.Getenv("PASSWD")
	containerDirectory = "/Volumes/Containers"
)

type Container struct {
	Id       bson.ObjectId `bson:"_id" json:"container_id"`
	Filename string        `bson:"filename" json:"filename"`
	Empty    bool          `bson:"empty" json:"empty"`
	FileHash string        `bson:"fileHash" json:"fileHash"`
	MimeType string        `bson:"mimeType" json:"mimeType"`
	Width    int           `bson:"width" json:"width"`
	Height   int           `bson:"height" json:"height"`
}

func (c *Container) String() string {
	oid := bson.ObjectId(c.Id)
	return fmt.Sprintf(`<Container Id:"%s" Filename:"%s" Empty:"%s" FileHash:"%s"`, oid.String(), c.Filename, c.Empty, c.FileHash)
}

type CreateRequest struct {
	Download_url string `json:"download_url,omitempty"`
	Filename     string `json:"filename"`
	Callback     string `json:"callback,omitempty"`
}

func getSession() *mgo.Session {
	if mgoSession == nil {
		var err error

		dialInfo := &mgo.DialInfo{
			Addrs:    []string{databaseHost},
			Database: databaseName,
			Username: databaseUser,
			Password: databasePasswd,
		}

		mgoSession, err = mgo.DialWithInfo(dialInfo)
		if err != nil {
			panic(err)
		}
	}

	return mgoSession.Clone()
}

func HomeHandler(c *gin.Context) {
	c.String(200, "Hail Hydra!")
}

func getContainer(container_id string) (Container, error) {

	session := getSession()
	defer session.Close()

	c := session.DB(databaseName).C("containers")

	result := Container{}
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(container_id)}).One(&result)
	if err != nil {
		return Container{}, err
	}

	return result, nil
}

func getFileHash(container_id string, filename string) string {
	sha1 := checksum.File(path.Join(containerDirectory, container_id, filename), crypto.SHA1)
	log.Println(sha1)
	return sha1
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

func getNewContainerId() string {
	container_id := bson.NewObjectId()
	return container_id.Hex()
}

func createContainer(container_id string, filename string, empty bool, fileHash string, mimeType string) Container {
	session := getSession()
	defer session.Close()

	c := session.DB(databaseName).C("containers")

	var width, height = 0, 0

	log.Println(path.Join(containerDirectory, container_id, filename))
	if mimeType == "image/png" {
		width, height = getImageDimension(path.Join(containerDirectory, container_id, filename))
	}

	newContainer := Container{
		Id:       bson.ObjectIdHex(container_id),
		Filename: filename,
		Empty:    empty,
		FileHash: fileHash,
		MimeType: mimeType,
		Width:    width,
		Height:   height,
	}

	c.Insert(newContainer)

	log.Println(container_id)

	return newContainer
}

func updateContainer(container_id string, filename string, empty bool, fileHash string, mimeType string) Container {
	session := getSession()
	defer session.Close()

	c := session.DB(databaseName).C("containers")

	var width, height = 0, 0

	if mimeType == "image/png" {
		width, height = getImageDimension(path.Join(containerDirectory, container_id, filename))
	}

	change := Container{
		Id:       bson.ObjectIdHex(container_id),
		Filename: filename,
		Empty:    empty,
		FileHash: fileHash,
		MimeType: mimeType,
		Width:    width,
		Height:   height,
	}

	query := bson.M{"_id": bson.ObjectIdHex(container_id)}

	err := c.Update(query, change)
	if err != nil {
		panic(err)
	}

	return change
}

func deleteContainer(container_id string) {
	session := getSession()
	defer session.Close()

	c := session.DB(databaseName).C("containers")

	err := c.RemoveId(bson.ObjectIdHex(container_id))
	if err != nil {
		log.Println(err)
	}
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

	container, err := getContainer(container_id)
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

	container, err := getContainer(container_id)
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

	var form struct {
		Download_url string `form:"download_url"`
		Filename     string `form:filename`
		Callback     string `form:callback`
	}

	c.BindWith(&form, binding.Form)

	cleanThumbnails(container_id)

	container, err := getContainer(container_id)
	if err != nil {
		log.Println(err)
	}

	err = os.Mkdir(path.Join(containerDirectory, container_id), 0777)
	if err != nil {
		log.Println(err)
	}

	if len(form.Download_url) > 0 {
		downloadFile(container_id, form.Download_url, form.Filename)

		fileHash := getFileHash(container_id, form.Filename)

		filePath := path.Join(containerDirectory, container_id, form.Filename)
		fileExt := filepath.Ext(filePath)
		mimeType := mime.TypeByExtension(fileExt)

		updateContainer(container_id, form.Filename, false, fileHash, mimeType)
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

		updateContainer(container_id, header.Filename, false, fileHash, mimeType)
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

	container, err := getContainer(container_id)
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

	container, err := getContainer(container_id)
	if err != nil {
		log.Println(err)
	}

	log.Println("DELETE: " + container_id)

	err = os.Remove(path.Join(containerDirectory, container_id, container.Filename))
	if err != nil {
		log.Println(err)
	}

	cleanThumbnails(container_id)

	updateContainer(container_id, "", true, "", "")

	c.JSON(200, gin.H{})
}

func ContainerCreateHandler(c *gin.Context) {
	log.Println("Creating New Container..")

	var form struct {
		Download_url string `json:"download_url"`
		Filename     string `form:"filename" json:"filename"`
		Callback     string `form:"callback" json:"callback"`
		IdOnly       bool   `json:"id_only"`
	}

	c.Bind(&form)

	fmt.Printf("%+v\n", form)
	if form.IdOnly {
		container_id := getNewContainerId()
		container := createContainer(container_id, "", true, "", "")

		c.JSON(201, container)
	} else if len(form.Download_url) > 0 {
		container_id := getNewContainerId()

		log.Println(form.Download_url)

		downloadFile(container_id, form.Download_url, form.Filename)

		fileHash := getFileHash(container_id, form.Filename)

		filePath := path.Join(containerDirectory, container_id, form.Filename)
		fileExt := filepath.Ext(filePath)
		mimeType := mime.TypeByExtension(fileExt)

		container := createContainer(container_id, form.Filename, false, fileHash, mimeType)

		if len(form.Callback) > 0 {
			resp, err := http.Post(form.Callback, "application/json", bytes.NewReader([]byte(`{"container_id": "`+container.Id+`"}`)))
			defer resp.Body.Close()
			if err != nil {
				log.Println(err)
			}

			c.String(201, "Done!")
		} else {
			c.JSON(201, container)
		}
	} else if _, _, err := c.Request.FormFile("file"); err == nil {
		log.Println("File upload")

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			log.Println(err)
		}

		log.Println(header.Filename)

		container_id := getNewContainerId()

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

		container := createContainer(container_id, header.Filename, false, fileHash, mimeType)

		if len(form.Callback) > 0 {
			log.Println("Callback: " + form.Callback)
			resp, err := http.Post(form.Callback, "application/json", bytes.NewReader([]byte(`{"container_id": "`+container_id+`"}`)))
			defer resp.Body.Close()
			if err != nil {
				log.Println(err)
			}
		}

		c.JSON(201, container)
	} else {
		container_id := getNewContainerId()
		container := createContainer(container_id, "", true, "", "")

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

	deleteContainer(container_id)

	c.String(200, "")
}

func main() {

	log.Println(fmt.Sprintf(`Database:"%s"`, databaseName))

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	api := router.Group("/api")

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
