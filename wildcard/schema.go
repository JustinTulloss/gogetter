// Structs for wildcard schema, documnted here:
// http://www.trywildcard.com/docs/schema/
package wildcard

import (
	"time"

	"github.com/JustinTulloss/gogetter/applink"
)

type CardType string

const (
	ArticleType       CardType = "article"
	ImageType         CardType = "image"
	LinkType          CardType = "link"
	ProductSearchType CardType = "product_search"
	ProductType       CardType = "product"
	ReviewType        CardType = "review"
	VideoType         CardType = "video"
)

type MediaType string

const (
	ImageMediaType MediaType = "image"
	VideoMediaType MediaType = "video"
)

// Every card must implement this interface
type Wildcard interface{}

// Every card has these
type Card struct {
	CardType CardType `json:"card_type"`
	WebUrl   string   `json:"web_url"`
}

// Metadata that pretty much every topic has
type GenericMetadata struct {
	Title           string     `json:"title,omitempty" ogtag:"og:title"`
	PublicationDate *time.Time `json:"publication_date,omitempty"`
	Source          string     `json:"source,omitempty"`
	Keywords        []string   `json:"keywords,omitempty"`

	// Our own addition, wildcard has a neutered version
	AppLink *applink.AppLink `json:"al,omitempty"`

	// Our own addition, is usually the favicon
	SourceIcon string `json:"source_icon,omitempty" ogtag:"favicon"`
	ImageUrl   string `json:"image_url,omitempty" ogtag:"og:image"`
}

type Article struct {
	AbstractContent string   `json:"abstract_content"`
	IsBreaking      bool     `json:"is_breaking,omitempty"`
	Contributors    []string `json:"contributors,omitempty"`
	GenericMetadata `ogtag:",squash"`
}

type ArticleCard struct {
	Card
	Article *Article
}

type VideoMedia struct {
	Type MediaType `json:"type"`

	// XXX: Perhaps pull these out for other embed types
	EmbeddedUrl       string `json:"embedded_url"`
	EmbeddedUrlWidth  string `json:"embedded_url_width" ogtag:"og:video:width"`
	EmbeddedUrlHeight string `json:"embedded_url_height" ogtag:"og:video:height"`

	// Optional
	StreamUrl         string `json:"stream_url,omitempty" ogtag:"og:video:url"`
	StreamContentType string `json:"stream_content_type,omitempty" ogtag:"og:video:type"`
	PosterImageUrl    string `json:"poster_image_url,omitempty"`
	Creator           string `json:"creator,omitempty"`
	GenericMetadata   `ogtag:",squash"`
}

type VideoCard struct {
	Card
	Media *VideoMedia `json:"media"`
}

func NewVideoCard(originalUrl string) *VideoCard {
	return &VideoCard{
		Card{
			CardType: VideoType,
			WebUrl:   originalUrl,
		},
		&VideoMedia{},
	}
}

type ImageMedia struct {
	Type     MediaType `json:"type"`
	ImageUrl string    `json:"image_url"`

	//Optional
	ImageCaption    string `json:"image_caption,omitempty"`
	Author          string `json:"author,omitempty"`
	Width           int    `json:"width,omitempty"`
	Height          int    `json:"height,omitempty"`
	GenericMetadata `ogtag:",squash"`

	// Added by us
	ImageContentType string `json:"image_content_type,omitempty"`
}

type ImageCard struct {
	Card
	Media *ImageMedia `json:"media"`
}

func NewImageCard(originalUrl, src string) *ImageCard {
	return &ImageCard{
		Card{
			CardType: ImageType,
			WebUrl:   originalUrl,
		},
		&ImageMedia{
			Type:     ImageMediaType,
			ImageUrl: src,
		},
	}
}

type LinkTarget struct {
	Url             string `json:"url"`
	Description     string `json:"description,omitempty" ogtag:"og:description"`
	GenericMetadata `ogtag:",squash"`
}

type LinkCard struct {
	Card
	Target *LinkTarget `json:"target"`
}

func NewLinkCard(originalUrl, linkUrl string) *LinkCard {
	return &LinkCard{
		Card{
			CardType: LinkType,
			WebUrl:   originalUrl,
		},
		&LinkTarget{
			Url: linkUrl,
		},
	}
}
