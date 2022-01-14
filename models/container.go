package models

import "time"

type Container struct {
	Key      string `json:"key"`
	Filename string `bson:"filename" json:"filename"`
	Empty    bool   `bson:"empty" json:"empty"`
	FileHash string `bson:"fileHash" json:"fileHash"`
	MimeType string `bson:"mimeType" json:"mimeType"`
	Width    int    `bson:"width" json:"width"`
	Height   int    `bson:"height" json:"height"`

	CreatedAt  time.Time `bson:"createdAt"`
	ModifiedAt time.Time `bson:"modifiedAt"`
}
