package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"time"

	"github.com/JustinTulloss/hut"
	"github.com/facebookgo/httpcontrol"
	"github.com/temoto/robotstxt.go"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var ogmatcher = regexp.MustCompile("(^(og|airbedandbreakfast|twitter):|^description$)")
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
	node, err := html.Parse(r)
	if err != nil {
		return nil, &HttpError{err.Error(), 500}
	}
	var findmeta func(*html.Node)
	results := make(map[string]string)
	// Recursively goes through nodes looking for meta nodes that
	// have a property tag that matches the opengraph regex. If it does,
	// Saves the contents to the results map.
	findmeta = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Title {
			results["title"] = n.FirstChild.Data
		}
		if n.Type == html.ElementNode && n.DataAtom == atom.Meta {
			var content, property string
			save := false
			for _, a := range n.Attr {
				if (a.Key == "property" || a.Key == "name") && ogmatcher.FindStringIndex(a.Val) != nil {
					save = true
					property = html.UnescapeString(a.Val)
				}
				if a.Key == "content" {
					content = html.UnescapeString(a.Val)
				}
			}
			if save {
				results[property] = content
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findmeta(c)
		}
	}
	findmeta(node)
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
			fmt.Sprintf("Could not fetch %s", url),
			resp.StatusCode,
		}
	}
	return parseTags(resp.Body)
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
	protocol, err := service.Env.Get("protocol")
	if err != nil {
		log.Fatal(err)
	}
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
