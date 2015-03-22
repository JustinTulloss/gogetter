package main

import (
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/JustinTulloss/hut"
	"github.com/PuerkitoBio/goquery"
	"github.com/facebookgo/httpcontrol"
	"github.com/temoto/robotstxt.go"
)

var ogPrefixes = []string{"og", "airbedandbreakfast", "twitter"}

var useragent = "Gogetter (https://github.com/JustinTulloss/gogetter) (like GoogleBot and facebookexternalhit)"
var service *hut.Service
var client *http.Client

type HttpError struct {
	msg        string
	StatusCode int
}

func (e *HttpError) Error() string { return e.msg }

func buildRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", useragent)
	return req, nil
}

func checkRobotsTxt(fullUrl string) (bool, error) {
	if service.Env.GetString("CHECK_ROBOTS_TXT") == "false" {
		return true, nil
	}
	parsed, err := url.Parse(fullUrl)
	if err != nil {
		return false, err
	}
	original := parsed.Path
	parsed.Path = "robots.txt"
	parsed.RawQuery = ""
	req, err := buildRequest(parsed.String())
	if err != nil {
		return false, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	robots, err := robotstxt.FromResponse(resp)
	if robots == nil {
		// Assume we can crawl if the robots.txt file doesn't work
		return true, nil
	}
	return robots.TestAgent(original, useragent), nil
}

func parseTags(r io.Reader) (map[string]string, *HttpError) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, &HttpError{err.Error(), 500}
	}
	results := make(map[string]string)
	// First we deal with a couple special tags to get the title
	// and the favicon
	title := doc.Find("title").Text()
	if title != "" {
		results["title"] = html.UnescapeString(title)
	}
	favicon, ok := doc.Find("link[rel~=icon]").Attr("href")
	if ok {
		results["favicon"] = html.UnescapeString(favicon)
	}
	// Find all meta tags for all different og prefixes we support
	tags := doc.Find(`meta[name="description"]`)
	for _, prefix := range ogPrefixes {
		tags = tags.Add(fmt.Sprintf(`meta[property^="%s:"]`, prefix))
		tags = tags.Add(fmt.Sprintf(`meta[name^="%s:"]`, prefix))
	}
	// For all the tags, extract the content
	tags.Each(func(i int, selection *goquery.Selection) {
		key, ok := selection.Attr("name")
		if !ok {
			key, _ = selection.Attr("property")
		}
		content, _ := selection.Attr("content")
		results[key] = html.UnescapeString(content)
	})
	return results, nil
}

func getTags(url string) (map[string]string, *HttpError) {
	permitted, err := checkRobotsTxt(url)
	if err != nil {
		return nil, &HttpError{err.Error(), 500}
	}
	if !permitted {
		msg := fmt.Sprintf("Not permitted to fetch %s as a robot", url)
		log.Println(msg)
		return nil, &HttpError{msg, 403}
	}
	req, _ := buildRequest(url)
	resp, err := client.Do(req)
	if err != nil {
		return nil, &HttpError{err.Error(), 500}
	}
	log.Printf("Fetched %s\n", url)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, &HttpError{
			fmt.Sprintf("Could not fetch %s, request was: %v", url, req),
			resp.StatusCode,
		}
	}
	// We can't really trust the Content-Type header, so we take
	// a look at what actually gets returned.
	contentStart, err := ioutil.ReadAll(io.LimitReader(resp.Body, 512))
	contentType := http.DetectContentType(contentStart)
	switch {
	case strings.Contains(contentType, "text/html"):
		return parseTags(resp.Body)
	default:
		return map[string]string{
			"mimeType": contentType,
		}, nil
	}
}

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
	tags, httpErr := getTags(decodedUrl)
	if httpErr != nil {
		service.HttpErrorReply(w, httpErr.Error(), httpErr.StatusCode)
		return
	}
	service.Reply(tags, w)
}

func main() {
	var err error
	service = hut.NewService(nil)
	jar, err := cookiejar.New(nil)
	if err != nil {
		service.Log.Error().Printf("Could not create cookie jar: %s\n", err)
	}
	service.Router.HandleFunc("/", handler)

	client = &http.Client{
		Transport: &httpcontrol.Transport{
			RequestTimeout: 10 * time.Second,
			MaxTries:       3,
		},
		Jar: jar,
	}

	flag.Parse()
	protocol := service.Env.Get("protocol")
	if protocol == "http" {
		service.Start()
	} else if len(flag.Args()) != 0 {
		tags, err := getTags(flag.Arg(0))
		if err != nil {
			fmt.Printf("Could not fetch: %s\n", err)
			return
		}
		for prop, val := range tags {
			fmt.Printf("%s -- %s\n", prop, val)
		}
	} else {
		panic("Need to use this properly and I need to print usage info!")
	}
}
