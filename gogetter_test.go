package main

import "reflect"
import "strings"
import "testing"

type doctest struct {
	doc  string
	tags map[string]string
}

var testdocs = []doctest{
	{
		`<html>
			<head>
				<title>Not interesting</title>
				<meta property="og:title" content="More interesting" />
			</head>
		</html>`,
		map[string]string{
			"og:title": "More interesting",
			"title":    "Not interesting",
		},
	},
	{
		`random gobbleygook
		<title>not interesting</title>
		<meta property="og:title" content="relevant" />`,
		map[string]string{
			"og:title": "relevant",
			"title":    "not interesting",
		},
	},
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
	},
	{
		`<meta property="description" content="pod (plain old descriptions) work" />`,
		map[string]string{
			"description": "pod (plain old descriptions) work",
		},
	},
}

func TestParseTags(t *testing.T) {
	t.Parallel()
	for _, doctest := range testdocs {
		testresult, err := parseTags(strings.NewReader(doctest.doc))
		if err != nil {
			t.Errorf(err.Error())
		}
		eq := reflect.DeepEqual(testresult, doctest.tags)
		if !eq {
			t.Errorf("%v != %v", testresult, doctest.tags)
		}
	}
}
