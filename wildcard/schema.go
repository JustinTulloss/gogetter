// Structs for wildcard schema, documented here:
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
	PublicationDate *time.Time `json:"publication_date,omitempty" ogtag:"article:published_time"`
	Source          string     `json:"source,omitempty" ogtag:"og:site_name"`
	Keywords        []string   `json:"keywords,omitempty"`

	// Our own addition, wildcard has a neutered version
	AppLink *applink.AppLink `json:"app_link,omitempty" ogtag:",fill"`

	// Our own addition, is usually the favicon
	SourceIcon string `json:"source_icon,omitempty" ogtag:"favicon"`

	// Our own addition since why wouldn't everything have an image?
	Image *ImageDetails `json:"image,omitempty" ogtag:",fill"`
}

type Article struct {
	AbstractContent string   `json:"abstract_content" ogtag:"og:description"`
	IsBreaking      bool     `json:"is_breaking,omitempty"`
	Contributors    []string `json:"contributors,omitempty"`
	GenericMetadata `ogtag:",squash"`
}

type ArticleCard struct {
	Card
	Article *Article `json:"article" ogtag:",fill"`
}

func NewArticleCard(webUrl, articleUrl string) *ArticleCard {
	return &ArticleCard{
		Card{
			CardType: ArticleType,
			WebUrl:   webUrl,
		},
		&Article{},
	}
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

type ImageDetails struct {
	ImageUrl string `json:"image_url" ogtag:"og:image"`
	Width    int    `json:"width,omitempty" ogtag:"og:image:width"`
	Height   int    `json:"height,omitempty" ogtag:"og:image:height"`

	// Added by us
	ImageContentType string `json:"image_content_type,omitempty"`
}

type ImageMedia struct {
	Type MediaType `json:"type"`
	ImageDetails

	//Optional
	ImageCaption string `json:"image_caption,omitempty"`
	Author       string `json:"author,omitempty"`
	GenericMetadata
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
			Type: ImageMediaType,
			ImageDetails: ImageDetails{
				ImageUrl: src,
			},
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
	Target *LinkTarget `json:"target" ogtag:",fill"`
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
