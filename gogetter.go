package main

import "code.google.com/p/go.net/html"
import "code.google.com/p/go.net/html/atom"
import "encoding/json"
import "flag"
import "fmt"
import "log"
import "net/http"
import "regexp"
import "strings"

var ogmatcher = regexp.MustCompile("^(og|airbedandbreakfast):")

func getTags(url string) (map[string]string, error) {
	resp, err := http.Get(url)
	if (err != nil) {
		return nil, err
	}
	log.Printf("Fetched %s\n", url)
	defer resp.Body.Close()
	node, err := html.Parse(resp.Body)
	if (err != nil) {
		log.Fatal(err)
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
				if a.Key == "property" && ogmatcher.FindStringIndex(a.Val) != nil {
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
			// TODO: 500 isn't always correct, getTags needs to do better
			http.Error(w, err.Error(), 500)
			return
		}
		encoded, err := json.Marshal(tags)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(encoded)
	}
	http.ListenAndServe(address, handler)
}

func main() {
	// First we deal with getting the arguments out
	address := flag.String("address", ":8080", "The address to bind to. Defaults to 'localhost:8080'")
	protocol := flag.String("protocol", "none", "The protocol to use. Can be 'http' or blank to start in command line mode")
	flag.Parse()

	if *protocol == "http" {
		startHttpServer(*address);
	} else if len(flag.Args()) != 0 {
		tags, _ := getTags("https://www.airbnb.com/rooms/339470")
		for prop, val := range tags {
			fmt.Printf("%s -- %s\n\n", prop, val)
		}
	}
}
