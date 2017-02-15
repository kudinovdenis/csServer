package newStorage

import (
	"os"
	"github.com/jinzhu/gorm"
	"github.com/kudinovdenis/csServer/logger"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/kudinovdenis/csServer/searchAPI"
)

// Tag ... tag
type Tag struct {
	ID        uint `gorm:"primary_key"`
	Name       string
	Confidence float64
	Image      Image
}

// Image ... image
type Image struct {
	ID        uint `gorm:"primary_key"`
	Hash  string
	LocalURL string
	Tags     []Tag `gorm:"many2many:image_tags;"`
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
	db.AutoMigrate(&Image{})
	db.AutoMigrate(&Tag{})
}

// IsImageExists ... check if image is already processed
func IsImageExists(hash string) bool {
	logger.Logf(logger.LogLevelDefault, "Checking if image %s exists", hash)
	if FindImageWithAssetID(hash).Hash == "" {
		logger.Logf(logger.LogLevelDefault, "Image %s is not exists", hash)
		return false
	}
	logger.Logf(logger.LogLevelDefault, "Image %s is exists", hash)
	return true
}

// FindImageWithAssetID ... find image with hash
func FindImageWithAssetID(hash string) Image {
	logger.Logf(logger.LogLevelDefault, "Getting image with ID %s", hash)
	var image Image
	internalDB.Where("hash = ?", hash).Preload("Tags").First(&image)
	logger.Logf(logger.LogLevelDefault, "Found image with ID %s: %v", hash, image)
	return image
}

func FindTagsForImage(hash string) []Tag {
	logger.Logf(logger.LogLevelDefault, "Finding tags for image with ID %s", hash)
	var tags []Tag
	if IsImageExists(hash) {
		tags = FindImageWithAssetID(hash).Tags
	}
	logger.Logf(logger.LogLevelDefault, "Found tags for image with ID %s: %+v", hash, tags)
	return tags
}

func SaveTags(tagsIN []searchAPI.Tag) []Tag {
	logger.Logf(logger.LogLevelDefault, "Saving tags %+v", tagsIN)
	tagsOut := []Tag{}
	for i := 0; i < len(tagsIN); i++ {
		tagIN := tagsIN[i]
		tagOut := Tag{Name: tagIN.Name, Confidence: tagIN.Confidence}
		if internalDB.NewRecord(tagOut) {
			internalDB.Create(&tagOut)
			logger.Logf(logger.LogLevelDefault, "Tag created: %+v", tagOut)
		} else {
			internalDB.Save(&tagOut)
			logger.Logf(logger.LogLevelDefault, "Tag updated: %+v", tagOut)
		}
		tagsOut = append(tagsOut, tagOut)
	}
	return tagsOut
}

// SaveImage ... add photo and tags to MySQL
func SaveImage(hash string, localURL string, tags []Tag) {
	logger.Logf(logger.LogLevelDefault, "Saving image %s", hash)
	image := Image{Hash: hash, LocalURL: localURL}
	if IsImageExists(hash) {
		internalDB.Create(&image)
		logger.Logf(logger.LogLevelDefault, "Created image %+v", image)
	} else {
		logger.Logf(logger.LogLevelDefault, "Updated image %+v", image)
	}
	logger.Logf(logger.LogLevelDefault, "Trying to link tags %+v", tags)
	for i := 0; i < len(tags); i++ {
		tag := tags[i]
		tag.Image = image
		logger.Logf(logger.LogLevelDefault, "Linking tag %+v", tag)
		if internalDB.NewRecord(tag) {
			internalDB.Create(&tag)
			logger.Logf(logger.LogLevelDefault, "Tag created %+v", tag)
		} else {
			internalDB.Save(&tag)
			logger.Logf(logger.LogLevelDefault, "Tag updated %+v", tag)
		}
		logger.Logf(logger.LogLevelDefault, "Linking tags %+v to image %s", tags, hash)
		image.Tags = append(image.Tags, tag)
	}
	internalDB.Save(&image)
}
