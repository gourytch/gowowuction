package fetcher

import (
	//	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	config "github.com/gourytch/gowowuction/config"
)

type FDesc struct {
	Url string `json:"url"`
	Lmt int64  `json:"lastModified"`
}

type Rec1 struct {
	Files []FDesc `json:"files"`
}

type Session struct {
	Config *config.Config
	Client *http.Client
}

func (s *Session) Get(url string) (body []byte) {
	if s.Client == nil {
		s.Client = new(http.Client)
	}
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("Accept-Encoding", "gzip")
	response, err := s.Client.Do(request)
	if err != nil {
		log.Fatalf(".. request failed: %s", url, err)
	}
	defer response.Body.Close()

	// Check that the server actually sent compressed data
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			log.Fatalf(".. gzip reader failed: %s", url, err)
		}
		defer reader.Close()
	default:
		reader = response.Body
	}
	body, err = ioutil.ReadAll(reader)
	if err != nil {
		log.Fatalf(".. request read failed: %s", url, err)
	}
	return body
}

func (s *Session) Fetch_FileURL(realm string, locale string) (url string, ts time.Time) {
	v := strings.Split(realm, ":")
	if len(v) != 2 {
		log.Fatalln("realm is in bad format: '" + realm + "'")
	}
	var data []byte
	url = fmt.Sprintf("https://%s.api.battle.net/wow/auction/data/%s?locale=%s&apikey=%s",
		v[0], v[1], locale, s.Config.APIKey)
	log.Printf("GET %s ...", url)
	data = s.Get(url)
	log.Println("parse auction file metainfo ...")

	var p1 Rec1
	if err := json.Unmarshal(data, &p1); err != nil {
		log.Fatalf("... json failed: %s", err)
	}
	url = p1.Files[0].Url
	lmt := p1.Files[0].Lmt
	ts = time.Unix(lmt/1000, lmt%1000).UTC()
	log.Printf("... url=%s, mtime=%s", url, ts)
	return
}
