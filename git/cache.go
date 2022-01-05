package git

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/adrg/xdg"
	log "github.com/sirupsen/logrus"
)

func userCacheDump(cache map[string]string) {
	// dont forget to import "encoding/json"
	cacheJSON, err := json.MarshalIndent(cache, "", "    ")
	if err != nil {
		panic(err)
	}

	cachePath := cacheLocation()

	err = os.WriteFile(cachePath, cacheJSON, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"err":       err,
			"cachePath": cachePath,
		}).Error("cannot write cache file")
	}
}

func userCacheLoad() map[string]string {
	cachePath := cacheLocation()

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return map[string]string{}
	}

	data, err := ioutil.ReadFile(cachePath)
	if err != nil {
		panic(err)
	}

	var cache map[string]string

	err = json.Unmarshal(data, &cache)
	if err != nil {
		panic(err)
	}

	return cache
}

func cacheLocation() string {
	cachePath, err := xdg.CacheFile("git-reviewers/emails.json")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Panic("cannot get a cache path")
	}
	log.WithFields(log.Fields{
		"cachePath": cachePath,
	}).Debug("emails cache path")

	return cachePath
}
