package storage

import (
	"database/sql"
	"os"

	"github.com/kudinovdenis/csServer/logger"
	"github.com/kudinovdenis/csServer/searchAPI"
	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

var internalDB *sql.DB

// InitDB ... initialize Database
func InitDB(name string) {
	mysqlIP := os.Getenv("MYSQL_IP_SERVER")
	logger.Logf(logger.LogLevelDefault, "MYSQL_IP_SERVER variable is %s", mysqlIP)
	if mysqlIP == "" {
		logger.Log(logger.LogLevelError, "MYSQL_IP_SERVER variable is not set")
		return
	}
	db, error := sql.Open("mysql", "root:bb5ih2xK3q@tcp("+mysqlIP+":3306)/")
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant create sql. %s", error.Error())
		return
	}
	error = db.Ping()
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant Ping sql. %s", error.Error())
		return
	}
	_, error = db.Exec("CREATE DATABASE IF NOT EXISTS " + name)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant create database %s. %s", name, error.Error())
		return
	}
	error = db.Close()
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant Ping Close sql. %s", error.Error())
		return
	}
	db, error = sql.Open("mysql", "root:bb5ih2xK3q@tcp("+mysqlIP+":3306)/"+name)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant use database storage. %s", error.Error())
		return
	}
	internalDB = db
	createTables()
}

func createTables() {
	_, error := internalDB.Exec(`
	CREATE TABLE Photos
	(
		id INT NOT NULL AUTO_INCREMENT,
		PRIMARY KEY (id),
		assetID VARCHAR(100) NOT NULL,
		localURL VARCHAR(100) NOT NULL
	)`)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant create table photos. %s", error.Error())
	}

	_, error = internalDB.Exec(`
	CREATE TABLE Tags
	(
		id INT NOT NULL AUTO_INCREMENT,
		PRIMARY KEY (id),
		name VARCHAR(255) NOT NULL
	)`)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant create table tags. %s", error.Error())
	}

	_, error = internalDB.Exec(`
	CREATE TABLE Photos_Tags
	(
		photo_id INT NOT NULL,
		tag_id INT NOT NULL,
		PRIMARY KEY (photo_id, tag_id),
		FOREIGN KEY (photo_id) REFERENCES Photos(id) ON UPDATE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES Tags(id) ON UPDATE CASCADE
	)`)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant create table tags. %s", error.Error())
	}
}

// SavePhoto ... add photo and tags to MySQL
func SavePhoto(assetID string, localURL string, tags []string) {
	if IsPhotoExists(assetID) {
		logger.Logf(logger.LogLevelDefault, "Trying to save Asset %s which is already in database", assetID)
		return
	}
	// tagsString := strings.Join(tags, ",")
	photoID := insertPhoto(assetID, localURL)
	insertTags(tags)
	//
	// make many to many relationships
	for i := 0; i < len(tags); i++ {
		tag := tags[i]
		result, error := internalDB.Query("SELECT id FROM Tags WHERE name = ?", tag)
		if error != nil {
			logger.Logf(logger.LogLevelError, "Cant select tags. %s", error.Error())
		}
		if result.Next() {
			var tagID int64
			result.Scan(&tagID)
			logger.Logf(logger.LogLevelDefault, "MANY TO MANY: %d, %d", photoID, tagID)
			_, error = internalDB.Exec(`
			INSERT INTO Photos_Tags (photo_id, tag_id) VALUES (?, ?)`, photoID, tagID)
			if error != nil {
				logger.Logf(logger.LogLevelError, "Cant save photo. %s", error.Error())
			}
		}
	}
}

func insertTags(tags []string) {
	// insert tags
	logger.Logf(logger.LogLevelDefault, "Inserting tags: %v ...", tags)
	for i := 0; i < len(tags); i++ {
		tag := tags[i]
		if isTagExists(tag) == false {
			_, error := internalDB.Exec("INSERT INTO Tags (name) VALUES (?);", tag)
			if error != nil {
				logger.Logf(logger.LogLevelError, "Cant insert tag: %s. %s", tag, error.Error())
			} else {
				logger.Logf(logger.LogLevelDefault, "Inserted tag tag: %s.", tag)
			}
		}
	}
}

func isTagExists(tag string) bool {
	logger.Logf(logger.LogLevelDefault, "Finding tag: %s", tag)
	result, error := internalDB.Query("SELECT * FROM Tags WHERE name = ?", tag)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant start finding tag. %s", error.Error())
		return false
	}
	if result.Next() {
		logger.Logf(logger.LogLevelDefault, "Tag found: %s", tag)
		return true
	}
	logger.Logf(logger.LogLevelDefault, "Tag not found: %s", tag)
	return false
}

func insertPhoto(assetID string, localURL string) int64 {
	logger.Logf(logger.LogLevelDefault, "Inserting photo %s with local URL %s", assetID, localURL)
	if IsPhotoExists(assetID) {
		return iDForPhoto(assetID)
	}
	insertPhotoRes, error := internalDB.Exec("INSERT INTO Photos (assetID, localURL) VALUES (?, ?);", assetID, localURL)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant insert photo. %s", error.Error())
		return -1
	}
	id, error := insertPhotoRes.LastInsertId()
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant save photo. %s", error.Error())
		return -1
	}
	logger.Logf(logger.LogLevelDefault, "Photo %s inserted with id %d", assetID, id)
	return id
}

// IsPhotoExists ... check if photo exists
func IsPhotoExists(assetID string) bool {
	logger.Logf(logger.LogLevelDefault, "Finding Photo: %s", assetID)
	result, error := internalDB.Query("SELECT * FROM Photos WHERE assetID = ?", assetID)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant start finding photo. %s", error.Error())
		return false
	}
	if result.Next() {
		logger.Logf(logger.LogLevelDefault, "Photo found: %s", assetID)
		return true
	}
	logger.Logf(logger.LogLevelDefault, "Photo not found: %s", assetID)
	return false
}

func iDForPhoto(assetID string) int64 {
	result, error := internalDB.Query("SELECT id FROM Photos WHERE assetID = ?", assetID)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Error in iDForPhoto: %s", error.Error())
		return -1
	}
	if result.Next() {
		var id int64
		result.Scan(&id)
		return id
	}
	return -1
}

// TagsForPhoto ... return tags for photo
func TagsForPhoto(assetID string) []searchAPI.Tag {
	logger.Logf(logger.LogLevelDefault, "Searching for tags for asset %s", assetID)
	var tags []searchAPI.Tag
	photoID := iDForPhoto(assetID)
	result, error := internalDB.Query("SELECT tag_id FROM Photos_Tags WHERE photo_id = ?", photoID)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cannot retreive tags from photo %s", assetID)
		return tags
	}
	for result.Next() {
		var tagID int64
		result.Scan(&tagID)
		tags = append(tags, findTagByID(tagID))
	}
	logger.Logf(logger.LogLevelDefault, "[LOCAL] Found tags: %v for asset %s", tags, assetID)
	return tags
}

func findTagByID(tagID int64) searchAPI.Tag {
	logger.Logf(logger.LogLevelDefault, "Searching for tag with ID %d", tagID)
	var tag searchAPI.Tag
	result, error := internalDB.Query("SELECT name FROM Tags WHERE id = ?", tagID)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant start finding tag with ID %d", tagID)
	}
	if result.Next() {
		var name string
		result.Scan(&name)
		tag.Name = name
		return tag
	}
	logger.Logf(logger.LogLevelError, "Cant find tag with ID %d", tagID)
	return tag
}
