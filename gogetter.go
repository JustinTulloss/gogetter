package main

import "code.google.com/p/go.net/html"
import "code.google.com/p/go.net/html/atom"
import "encoding/json"
import "flag"
import "fmt"
import "github.com/temoto/robotstxt.go"
import "log"
import "net/http"
import "net/url"
import "regexp"
import "strings"

var ogmatcher = regexp.MustCompile("^(og|airbedandbreakfast|twitter):")
var useragent = "Gogetter (https://github.com/JustinTulloss/gogetter) (like GoogleBot)"

type HttpError struct {
	msg string
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
	client := http.Client{}
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
	client := http.Client{}
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
	node, err := html.Parse(resp.Body)
	if err != nil {
		return nil, &HttpError{err.Error(), 500}
	}
	var findmeta func (*html.Node)
	results := make(map[string]string)
	// Recursively goes through nodes looking for meta nodes that
	// have a property tag that matches the opengraph regex. If it does,
	// Saves the contents to the results map.
	findmeta = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Meta  {
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
			if (save) {
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

func startHttpServer(address string) {
	log.Printf("Starting http server on %s\n", address)
	var handler http.HandlerFunc = func (w http.ResponseWriter, r *http.Request) {
		tags, err := getTags(strings.TrimPrefix(r.URL.Path, "/"))
		if err != nil {
			http.Error(w, err.Error(), err.StatusCode)
			return
		}
		encoded, err2 := json.Marshal(tags)
		if err2 != nil {
			http.Error(w, err2.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(encoded)
	}
	http.ListenAndServe(address, handler)
}

func main() {
	// First we deal with getting the arguments out
	address := flag.String("address", ":8080", "The address to bind to. Defaults to ':8080'")
	protocol := flag.String("protocol", "none", "The protocol to use. Can be 'http' or blank to start in command line mode")
	flag.Parse()

	if *protocol == "http" {
		startHttpServer(*address);
	} else if len(flag.Args()) != 0 {
		tags, _ := getTags(flag.Arg(0))
		for prop, val := range tags {
			fmt.Printf("%s -- %s\n", prop, val)
		}
	}
}
