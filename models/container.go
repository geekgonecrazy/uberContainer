package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Container struct {
	ID       bson.ObjectId `bson:"_id" json:"-"`
	Key      string        `json:"key"`
	Filename string        `bson:"filename" json:"filename"`
	FileSize int64         `bson:"fileSize" json:"fileSize"`
	Empty    bool          `bson:"empty" json:"empty"`
	FileHash string        `bson:"fileHash" json:"fileHash"`
	MimeType string        `bson:"mimeType" json:"mimeType"`
	Width    int           `bson:"width" json:"width"`
	Height   int           `bson:"height" json:"height"`

	PreviewGenerated bool `bson:"previewGenerated" json:"previewGenerated"`

	CreatedAt  time.Time `bson:"createdAt" json:"createdAt"`
	ModifiedAt time.Time `bson:"modifiedAt" json:"modifiedAt"`
}
