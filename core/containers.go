package core

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"image"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/FideTechSolutions/uberContainer/models"
	"github.com/FideTechSolutions/uberContainer/store"
)

func GetContainer(container_id string) (models.Container, error) {
	container, err := _store.GetContainer(container_id)
	if err != nil {
		return models.Container{}, err
	}

	return container, nil
}

func GetContainers() ([]models.Container, error) {
	return _store.GetContainers()
}

func GetContainerFileLink(container_id string) (string, error) {
	container, err := GetContainer(container_id)
	if err != nil {
		return "", err
	}

	return _storage.GetDownloadLink(fmt.Sprintf("%s/%s", container.Key, container.Filename))
}

func GetContainerPreviewLink(container_id string) (string, error) {
	container, err := GetContainer(container_id)
	if err != nil {
		return "", err
	}

	return _storage.GetDownloadLink(fmt.Sprintf("%s/preview.png", container.Key))
}

func CreateContainerFromFileUploadFromForm(form models.ContainerCreateUpdatePayload, fileHeader *multipart.FileHeader, file io.Reader) (*models.Container, error) {
	fileExt := filepath.Ext(fileHeader.Filename)
	mimeType := mime.TypeByExtension(fileExt)

	if !strings.HasPrefix(form.ContainerKey, "/") {
		form.ContainerKey = "/" + form.ContainerKey
	}

	container := models.Container{
		Key:      form.ContainerKey,
		Filename: fileHeader.Filename,
		FileSize: fileHeader.Size,
		//FileHash: fileHash,
		Empty:    false,
		MimeType: mimeType,
		Height:   0,
		Width:    0,
	}

	log.Println(container)

	if err := _store.CreateContainer(&container); err != nil {
		return nil, err
	}

	if err := uploadAndGenerateThumbnail(&container, file, "250", mimeType); err != nil {
		return nil, err
	}

	if err := _store.UpdateContainer(&container); err != nil {
		return nil, err
	}

	if len(form.Callback) > 0 {
		if err := uploadCallback(&container, form.Callback); err != nil {
			return &container, err
		}
	}

	return nil, nil
}

func CreateContainerFromFileUploadFromUrl(form models.ContainerCreateUpdatePayload) (*models.Container, error) {
	log.Println(form.DownloadUrl)

	downloadFile(form.ContainerKey, form.DownloadUrl, form.Filename)

	fileHash := getFileHash(form.ContainerKey, form.Filename)

	filePath := path.Join(previewTempDirectory, form.ContainerKey, form.Filename)
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

	file, fileSize, err := downloadFile(form.ContainerKey, form.DownloadUrl, form.Filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	container.FileSize = fileSize

	if err := _store.CreateContainer(&container); err != nil {
		return nil, err
	}

	if err := uploadAndGenerateThumbnail(&container, file, "250", mimeType); err != nil {
		return nil, err
	}

	if err := _store.UpdateContainer(&container); err != nil {
		return nil, err
	}

	if len(form.Callback) > 0 {
		if err := uploadCallback(&container, form.Callback); err != nil {
			return &container, err
		}
	}

	return &container, nil
}

func UpdateContainerFromFileUploadFromForm(form models.ContainerCreateUpdatePayload, fileHeader *multipart.FileHeader, file io.Reader) (*models.Container, error) {
	container, err := GetContainer(form.ContainerKey)
	if err != nil {
		return nil, err
	}

	fileExt := filepath.Ext(fileHeader.Filename)
	mimeType := mime.TypeByExtension(fileExt)

	if !strings.HasPrefix(form.ContainerKey, "/") {
		form.ContainerKey = "/" + form.ContainerKey
	}

	container.Filename = fileHeader.Filename
	container.FileSize = fileHeader.Size
	//container.FileHash = fileHash
	container.Empty = false
	container.MimeType = mimeType

	if err := uploadAndGenerateThumbnail(&container, file, "250", mimeType); err != nil {
		return nil, err
	}

	log.Println(container)

	if err := _store.UpdateContainer(&container); err != nil {
		return nil, err
	}

	if len(form.Callback) > 0 {
		if err := uploadCallback(&container, form.Callback); err != nil {
			return &container, err
		}
	}

	return nil, nil
}

func UpdateContainerFromFileUploadFromUrl(form models.ContainerCreateUpdatePayload) (*models.Container, error) {
	container, err := GetContainer(form.ContainerKey)
	if err != nil {
		return nil, err
	}

	log.Println(form.DownloadUrl)

	downloadFile(form.ContainerKey, form.DownloadUrl, form.Filename)

	fileHash := getFileHash(form.ContainerKey, form.Filename)

	filePath := path.Join(previewTempDirectory, form.ContainerKey, form.Filename)
	fileExt := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(fileExt)

	container.Filename = form.Filename
	container.FileSize = 0
	container.FileHash = fileHash
	container.Empty = false
	container.MimeType = mimeType

	file, fileSize, err := downloadFile(form.ContainerKey, form.DownloadUrl, form.Filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	container.FileSize = fileSize

	if err := uploadAndGenerateThumbnail(&container, file, "250", mimeType); err != nil {
		return nil, err
	}

	if err := _store.UpdateContainer(&container); err != nil {
		return nil, err
	}

	if len(form.Callback) > 0 {
		if err := uploadCallback(&container, form.Callback); err != nil {
			return &container, err
		}
	}

	return &container, nil
}

func DeleteContainerFile(container_id string) error {
	container, err := GetContainer(container_id)
	if err != nil && err != store.ErrNotFound {
		return err
	}

	if err == store.ErrNotFound {
		return nil
	}

	_storage.Delete(fmt.Sprintf("%s/%s", container.Key, container.Filename))

	if container.PreviewGenerated {
		_storage.Delete(fmt.Sprintf("%s/preview.png", container.Key))
	}

	container.Empty = true
	container.FileHash = ""
	container.MimeType = ""
	container.Width = 0
	container.Height = 0
	container.Filename = ""
	container.PreviewGenerated = false

	_store.UpdateContainer(&container)

	return nil
}

func DeleteContainer(container_id string) error {
	DeleteContainerFile(container_id)
	_store.DeleteContainer(container_id)

	return nil
}

func uploadCallback(container *models.Container, callback string) error {
	resp, err := http.Post(callback, "application/json", bytes.NewReader([]byte(`{"container_id": "`+container.Key+`"}`)))
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
	}

	return nil
}

func getFileHash(container_id string, filename string) string {
	//sha1 := checksum.File(path.Join(containerDirectory, container_id, filename), crypto.SHA1)
	//log.Println(sha1)

	return "sha1"
}

func getImageDimension(filePath string) (int, int) {
	log.Println("Getting image dimensions for", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		log.Println(err)
	}

	log.Println("image meta from file", image)
	return image.Width, image.Height
}

func downloadFile(container_id string, download_url string, filename string) (io.ReadCloser, int64, error) {
	log.Println("Download Url Passed")
	log.Println("URL: " + download_url)
	log.Println("Filename: " + filename)

	u, err := url.Parse(download_url)
	if err != nil {
		log.Println(err)
	}

	client := &http.Client{}

	if u.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client.Transport = tr
	}

	resp, err := client.Get(download_url)
	if err != nil {
		log.Println(err)
	}

	log.Println(resp.Status)

	return resp.Body, resp.ContentLength, nil
}

func uploadAndGenerateThumbnail(container *models.Container, file io.Reader, size string, mimeType string) error {

	containerFilePrefix := strings.ReplaceAll(container.Key, "/", "-")

	thumbPath := path.Join(previewTempDirectory, fmt.Sprintf("%s-preview.png", containerFilePrefix))
	filePath := path.Join(previewTempDirectory, fmt.Sprintf("%s-%s", containerFilePrefix, container.Filename))

	out, err := os.Create(filePath)
	defer out.Close()
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, file); err != nil {
		return err
	}

	log.Println(" Thumbnail has been requested for container: " + container.Key)

	_, err = os.Stat(thumbPath)
	if err != nil {
		log.Println("Preview image needs updated or Created")

		fileExt := filepath.Ext(container.Filename)

		if fileExt == ".psd" {
			filePath += "[0]"
		}

		skipPreview := false

		if fileExt == ".zip" {
			skipPreview = true
		}

		if err := _storage.Upload(fmt.Sprintf("%s/%s", container.Key, container.Filename), filePath, mimeType); err != nil {
			return err
		}

		if !skipPreview {
			_, err := exec.Command("/usr/bin/convert", filePath, "-thumbnail", size+"x"+size, thumbPath).Output()
			if err != nil {
				log.Println("Failed to generate preview", err)
				return nil
			}

			if _, err := os.Stat(thumbPath); err != nil {
				log.Println("File missing from preview", err)
				return err
			}

			if err := _storage.Upload(fmt.Sprintf("%s/preview.png", container.Key), thumbPath, "image/png"); err != nil {
				return err
			}

			container.PreviewGenerated = true
		}

		width, height := getImageDimension(filePath)
		container.Width = width
		container.Height = height

		err = os.Remove(thumbPath)

		err = os.Remove(filePath)

		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
