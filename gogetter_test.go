package gogetter

import (
	"reflect"

	"github.com/JustinTulloss/gogetter/applink"
	"github.com/JustinTulloss/gogetter/wildcard"
)
import "strings"
import (
	"testing"
	"time"
)

type doctest struct {
	doc  string
	tags interface{}
}

var testdocs = []doctest{
	{
		`<html>
			<head>
				<title>Not interesting</title>
				<meta property="og:title" content="More interesting" />
			</head>
		</html>`,
		&wildcard.LinkCard{
			Card: wildcard.Card{
				CardType: wildcard.LinkType,
			},
			Target: &wildcard.LinkTarget{
				GenericMetadata: wildcard.GenericMetadata{
					Title: "More interesting",
					AppLink: &applink.AppLink{
						Ios:     &applink.Ios{},
						Iphone:  &applink.Iphone{},
						Ipad:    &applink.Ipad{},
						Android: &applink.Android{},
					},
					Image:           &wildcard.ImageDetails{},
					PublicationDate: &time.Time{},
				},
			},
		},
	},
	{
		`random gobbleygook
		<title>not interesting</title>
		<meta property="og:title" content="relevant" />`,
		&wildcard.LinkCard{
			Card: wildcard.Card{
				CardType: wildcard.LinkType,
			},
			Target: &wildcard.LinkTarget{
				GenericMetadata: wildcard.GenericMetadata{
					Title: "relevant",
					AppLink: &applink.AppLink{
						Ios:     &applink.Ios{},
						Iphone:  &applink.Iphone{},
						Ipad:    &applink.Ipad{},
						Android: &applink.Android{},
					},
					Image:           &wildcard.ImageDetails{},
					PublicationDate: &time.Time{},
				},
			},
		},
	},
	/*
			{
				`<meta property="twitter:hello" content="twitter works" />`,
				map[string]string{
					"twitter:hello": "twitter works",
				},
			},
		{
			`<meta property="airbedandbreakfast:test" content="airbnb works" />`,
			map[string]string{
				"airbedandbreakfast:test": "airbnb works",
			},
		},*/
	{
		`<meta name="description" content="pod (plain old descriptions) work" />`,
		&wildcard.LinkCard{
			Card: wildcard.Card{
				CardType: wildcard.LinkType,
			},
			Target: &wildcard.LinkTarget{
				Description: "pod (plain old descriptions) work",
				GenericMetadata: wildcard.GenericMetadata{
					AppLink: &applink.AppLink{
						Ios:     &applink.Ios{},
						Iphone:  &applink.Iphone{},
						Ipad:    &applink.Ipad{},
						Android: &applink.Android{},
					},
					Image:           &wildcard.ImageDetails{},
					PublicationDate: &time.Time{},
				},
			},
		},
	},
}

func TestParseTags(t *testing.T) {
	t.Parallel()
	scraper, err := NewScraper("", false)
	if err != nil {
		t.Logf("Could not create scraper: %s\n", err)
		t.Fail()
	}
	for _, doctest := range testdocs {
		testresult, err := scraper.ParseTags(strings.NewReader(doctest.doc), "")
		if err != nil {
			t.Errorf(err.Error())
		}
		eq := reflect.DeepEqual(testresult, doctest.tags)
		if !eq {
			t.Errorf("%#v != %#v", testresult, doctest.tags)
		}
	}
}
