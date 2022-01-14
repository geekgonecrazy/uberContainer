package models

type ContainerCreateUpdatePayload struct {
	ContainerKey string `form:"container_key" json:"containerKey"`
	DownloadUrl  string `form:"download_url" json:"downloadUrl"`
	Filename     string `form:"filename" json:"filename"`
	Callback     string `form:"callback" json:"callback"`
	IdOnly       bool   `json:"idOnly"`
}

type CreateRequest struct {
	DownloadUrl string `json:"downloadUrl,omitempty"`
	Filename    string `json:"filename"`
	Callback    string `json:"callback,omitempty"`
}
