package newStorage

import (
	"os"

	"github.com/jinzhu/gorm"
	"github.com/kudinovdenis/csServer/logger"
)

// Tag ... tag
type Tag struct {
	gorm.Model
	Name       string
	Confidence float64
	Images     []Image
}

// Image ... image
type Image struct {
	gorm.Model
	AssetID  string
	LocalURL string
	Tags     []Tag
}

var internalDB *gorm.DB

// InitDB .. initialize Database
func InitDB(name string) {
	mysqlIP := os.Getenv("MYSQL_IP_SERVER")
	logger.Logf(logger.LogLevelDefault, "MYSQL_IP_SERVER variable is %s", mysqlIP)
	if mysqlIP == "" {
		logger.Log(logger.LogLevelError, "MYSQL_IP_SERVER variable is not set")
		return
	}
	db, err := gorm.Open("mysql", "root:bb5ih2xK3q@tcp("+mysqlIP+":3306)/")
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant open database. %s", err.Error())
		return
	}
	db = db.Exec("CREATE DATABASE IF NOT EXISTS " + name)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant create database %s. %s", name, err.Error())
		return
	}
	db, err = gorm.Open("mysql", "root:bb5ih2xK3q@tcp("+mysqlIP+":3306)/"+name)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant open database. %s", err.Error())
		return
	}
	internalDB = db
	if !db.HasTable(&Image{}) {
		db.CreateTable(&Image{})
	}
	if !db.HasTable(&Tag{}) {
		db.CreateTable(&Tag{})
	}
}

// IsImageExists ... check if image is already processed
func IsImageExists(assetID string) bool {
	if FindImageWithAssetID(assetID).AssetID == "" {
		return false
	}
	return true
}

// FindImageWithAssetID ... find image with asset id
func FindImageWithAssetID(assetID string) Image {
	var image Image
	internalDB.Where("assetID = ?", assetID).First(&image)
	return image
}

// SavePhoto ... add photo and tags to MySQL
func SavePhoto(assetID string, localURL string, tags []Tag) {
	for i := 0; i < len(tags); i++ {
		tag := tags[i]
		if internalDB.NewRecord(tag) {
			internalDB.Create(tag)
		}
	}
}
