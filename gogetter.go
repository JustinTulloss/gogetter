package gogetter

import (
	"errors"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/facebookgo/httpcontrol"
	"github.com/temoto/robotstxt.go"
)

// There are potentially a ton of these as any facebook app can enter their own
// prefixes.
var ogPrefixes = []string{"og", "airbedandbreakfast", "twitter"}

const DEFAULT_UA = "Gogetter (https://github.com/JustinTulloss/gogetter) (like GoogleBot and facebookexternalhit)"

// A Scraper instance can be used to scrape webpages for metadata.
type Scraper struct {
	useragent            string
	shouldCheckRobotsTxt bool
	client               *http.Client
}

func (s *Scraper) buildRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.useragent)
	req.Header.Set("Accept", "*/*")
	return req, nil
}

func (s *Scraper) checkRobotsTxt(fullUrl string) (bool, error) {
	if !s.shouldCheckRobotsTxt {
		return true, nil
	}
	parsed, err := url.Parse(fullUrl)
	if err != nil {
		return false, err
	}
	original := parsed.Path
	parsed.Path = "robots.txt"
	parsed.RawQuery = ""
	req, err := s.buildRequest(parsed.String())
	if err != nil {
		return false, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	robots, err := robotstxt.FromResponse(resp)
	if robots == nil {
		// Assume we can crawl if the robots.txt file doesn't work
		return true, nil
	}
	return robots.TestAgent(original, s.useragent), nil
}

func (s *Scraper) ParseTags(r io.Reader) (map[string]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
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

func (s *Scraper) ScrapeTags(url string) (map[string]string, error) {
	permitted, err := s.checkRobotsTxt(url)
	if err != nil {
		return nil, err
	}
	if !permitted {
		return nil, errors.New(fmt.Sprintf("Not permitted to fetch %s", url))
	}
	req, err := s.buildRequest(url)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		errMsg := fmt.Sprintf("Could not fetch %s: %s", url, string(body))
		return nil, errors.New(errMsg)
	}
	contentType := resp.Header.Get("Content-Type")
	switch {
	case contentType == "":
		fallthrough
	case strings.Contains(contentType, "text/html"):
		return s.ParseTags(resp.Body)
	default:
		// We can't really trust the Content-Type header, so we take
		// a look at what actually gets returned.
		contentStart, err := ioutil.ReadAll(io.LimitReader(resp.Body, 512))
		if err != nil {
			contentType = http.DetectContentType(contentStart)
		}
		return map[string]string{
			"mimeType": contentType,
		}, nil
	}
}

// Creates a new scraper. If no user agent is provided, DEFAULT_UA is used.
func NewScraper(ua string, shouldCheckRobotsTxt bool) (*Scraper, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &httpcontrol.Transport{
			RequestTimeout: 10 * time.Second,
			MaxTries:       3,
		},
		Jar: jar,
	}
	if ua == "" {
		ua = DEFAULT_UA
	}
	return &Scraper{
		useragent:            ua,
		shouldCheckRobotsTxt: shouldCheckRobotsTxt,
		client:               client,
	}, nil
}
