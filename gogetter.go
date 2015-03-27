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
	"reflect"
	"strings"
	"time"

	"github.com/JustinTulloss/gogetter/wildcard"
	"github.com/PuerkitoBio/goquery"
	"github.com/facebookgo/httpcontrol"
	"github.com/mitchellh/mapstructure"
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

var tagAliases = map[string][]string{
	"al:android:url":         {"twitter:app:url:googleplay"},
	"al:ipad:url":            {"twitter:app:url:ipad"},
	"al:iphone:url":          {"twitter:app:url:iphone"},
	"article:published_time": {"article:published"},
	"og:description":         {"twitter:description", "description"},
	"og:image":               {"twitter:image"},
	"og:site_name":           {"cre"},
	"og:title":               {"twitter:title", "title"},
}

// Finds other names for the same value and puts it in the map
// under the name we prefer.
func resolveAliases(tags map[string]string) {
	for tag, aliases := range tagAliases {
		_, ok := tags[tag]
		if ok {
			continue
		}
		for _, alias := range aliases {
			val, ok := tags[alias]
			if ok {
				tags[tag] = val
				break
			}
		}
	}
}

// Used by recusivelyDecode to look at every field in a struct
func iterateOverFields(value reflect.Value, tags map[string]string) error {
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		structField := value.Type().Field(i)
		if field.Kind() == reflect.Ptr && field.CanSet() && structField.Tag.Get("ogtag") != "" {
			if field.IsNil() && field.CanSet() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			err := recursivelyDecode(tags, field.Interface())
			if err != nil {
				return err
			}
		} else if field.Kind() == reflect.Struct && structField.Anonymous {
			err := iterateOverFields(field, tags)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// This is a hacky way of transferring a flat map of tags to values into a
// nested structure. Since each leaf in the structure has a unique key in the
// flat map, we can just iterate through every struct that might potentially
// have tags (as indicated by the `ogtag` struct tag) and try to match the
// tags to the fields using mapstructure.
func recursivelyDecode(tags map[string]string, result interface{}) error {
	decoderConfig := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		TagName:          "ogtag",
		Result:           result,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}
	err = decoder.Decode(tags)
	if err != nil {
		return err
	}
	value := reflect.Indirect(reflect.ValueOf(result))
	return iterateOverFields(value, tags)
}

func convertTagsToCard(tags map[string]string, webUrl string) (wildcard.Wildcard, error) {
	resolveAliases(tags)
	ogType, ok := tags["og:type"]
	if !ok {
		ogType = "website"
	}
	var card wildcard.Wildcard
	url, ok := tags["og:url"]
	if !ok {
		url = webUrl
	}
	switch ogType {
	case "article":
		card = wildcard.NewArticleCard(webUrl, url)
	default:
		card = wildcard.NewLinkCard(webUrl, url)
	}
	err := recursivelyDecode(tags, card)
	if err != nil {
		return nil, err
	}
	return card, nil
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

var rawMetaTags = []string{"cre"}

func (s *Scraper) ParseTags(r io.Reader, url string) (wildcard.Wildcard, error) {
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
	for _, metaTag := range rawMetaTags {
		tags = tags.Add(fmt.Sprintf(`meta[name="%s"]`, metaTag))
	}
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
		// Open graph defers to the first tag that we understand.
		_, alreadySet := results[key]
		if !alreadySet {
			results[key] = html.UnescapeString(content)
		}
	})
	card, err := convertTagsToCard(results, url)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (s *Scraper) ScrapeTags(url string) (interface{}, error) {
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
	if contentType == "" || strings.Contains(contentType, "text/html") {
		return s.ParseTags(resp.Body, url)
	}
	// We can't really trust the Content-Type header, so we take
	// a look at what actually gets returned.
	contentStart, err := ioutil.ReadAll(io.LimitReader(resp.Body, 512))
	if err != nil {
		contentType = http.DetectContentType(contentStart)
	}
	switch {
	case strings.HasPrefix(contentType, "image"):
		card := wildcard.NewImageCard(url, url)
		card.Media.ImageContentType = contentType
		return card, nil
	case strings.HasPrefix(contentType, "video"):
		card := wildcard.NewVideoCard(url)
		card.Media.StreamUrl = url
		card.Media.StreamContentType = contentType
		return card, nil
	default:
		card := wildcard.NewLinkCard(url, url)
		return card, nil
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
