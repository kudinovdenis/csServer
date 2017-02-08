package storage

import (
	"github.com/kudinovdenis/csServer/logger"
	"github.com/kudinovdenis/csServer/searchAPI"
)

// FindTopTags ... find top tags
func FindTopTags(limit int) []searchAPI.Tag {
	var tags []searchAPI.Tag
	results, err := internalDB.Query(`
		SELECT t.name, count(t.name)
		       FROM  photos_tags l
		       JOIN  tags t ON l.tag_id = t.id
		       JOIN  photos p ON l.photo_id =p.id
		       GROUP BY f.name`)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant search top tags %s", err.Error())
	}
	for results.Next() {
		columns, err := results.Columns()
		if err != nil {
			logger.Logf(logger.LogLevelError, "Cant get columns for top tags result %s", err.Error())
		}
		logger.Logf(logger.LogLevelDefault, "Results: %v", columns)
	}
	return tags
}

/*
SELECT t.name, count(t.name)
       FROM  photos_tags l
       JOIN  tags t ON l.tag_id = t.id
       JOIN  photos p ON l.photo_id =p.id
       GROUP BY f.name
*/
