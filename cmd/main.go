package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"github.com/JustinTulloss/gogetter"
	"github.com/JustinTulloss/hut"
)

var service *hut.Service
var scraper *gogetter.Scraper

func handler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		service.ErrorReply(err, w)
		return
	}
	decodedUrl, err := url.QueryUnescape(r.Form.Get("url"))
	if err != nil {
		service.ErrorReply(err, w)
		return
	}
	tags, httpErr := scraper.ScrapeTags(decodedUrl)
	if httpErr != nil {
		service.HttpErrorReply(w, httpErr.Error(), http.StatusInternalServerError)
		return
	}
	service.Reply(tags, w)
}

func main() {
	var err error
	service = hut.NewService(nil)
	service.Router.HandleFunc("/", handler)
	scraper, err = gogetter.NewScraper("", service.Env.GetBool("check_robots_txt"))
	if err != nil {
		service.Log.Error().Fatalln(err)
	}

	flag.Parse()
	protocol := service.Env.GetString("protocol")
	if protocol == "http" {
		service.Start()
	} else if len(flag.Args()) != 0 {
		tags, err := scraper.ScrapeTags(flag.Arg(0))
		if err != nil {
			fmt.Printf("Could not fetch: %s\n", err)
			return
		}
		jsonTags, err := json.MarshalIndent(tags, "", "\t")
		if err != nil {
			fmt.Printf("Could not marshal json: %s\n", err)
		}
		fmt.Println(jsonTags)
	} else {
		panic("Need to use this properly and I need to print usage info!")
	}
}
