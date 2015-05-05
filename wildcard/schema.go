// Structs for wildcard schema, documented here:
// http://www.trywildcard.com/docs/schema/
package wildcard

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JustinTulloss/gogetter/applink"
)

type CardType string

const (
	ArticleType       CardType = "article"
	ImageType         CardType = "image"
	LinkType          CardType = "link"
	PlaceType         CardType = "place"
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
	Url             string   `json:"url"`
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
		&Article{
			Url: articleUrl,
		},
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
	PosterImageUrl    string `json:"poster_image_url,omitempty" ogtag:"og:image:url"`
	Creator           string `json:"creator,omitempty"`
	GenericMetadata   `ogtag:",squash"`
}

type VideoCard struct {
	Card
	Media *VideoMedia `json:"media" ogtag:",fill"`
}

func NewVideoCard(originalUrl string) *VideoCard {
	return &VideoCard{
		Card{
			CardType: VideoType,
			WebUrl:   originalUrl,
		},
		&VideoMedia{
			Type: VideoMediaType,
		},
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

// Like, where to send snail mail. Quite possibly a physical address.
type PostalAddress struct {
	StreetAddress       string `json:"street_address"`
	PostOfficeBoxNumber string `json:"post_office_box_number,omitempty"`
	// In the US, this is the city
	Locality string `json:"locality,omitempty"`
	// In the US, this is the state
	Region     string `json:"region,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

// Returns the address as a nicely formatted string on a single line.
// TODO: Make this deal with missing data better
func (a *PostalAddress) Formatted() string {
	addr := strings.Replace(a.StreetAddress, "\n", ", ", -1)
	if addr == "" {
		addr = a.PostOfficeBoxNumber
	}
	return fmt.Sprintf("%s, %s, %s, %s, %s", addr, a.Locality, a.Region, a.PostalCode, a.Country)
}

// TODO: Make this deal with missing data better
func (a *PostalAddress) MultiLineFormatted() string {
	addr := a.StreetAddress
	if addr == "" {
		addr = a.PostOfficeBoxNumber
	}
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s", addr, a.Locality, a.Region, a.PostalCode, a.Country)
}

type GeoCoordinates struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Elevation *float64 `json:"elevation,omitempty"`
}

type Rating struct {
	// What this is actually rated.
	Value string `json:"value"`

	// If this thing is perfect, this is what it would be rated.
	BestRating string `json:"best_rating,omitempty"`

	// This is almost always 1 (and should be assumed to be 1 if it's missing),
	// but it's the minimum rating.
	WorstRating string `json:"worst_rating,omitempty"`

	// Using an int32 here even though it limits things to 4 billion ratings.
	RatingCount int32 `json:"rating_count,omitempty"`
	ReviewCount int32 `json:"review_count,omitempty"`

	// An image that can be used to represent this rating.
	ImageUrl string `json:"image_url,omitempty"`
}

// This is different than the regular go time.Time because it serializes
// to a 24 hour clock instead of an actual point in time.
// Courtesy of @smagch -- https://gist.github.com/smagch/d2a55c60bbd76930c79f
var timeLayout = "15:04"
var TimeParseError = errors.New(`TimeParseError: should be a string formatted as "15:04:05"`)

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Format(timeLayout) + `"`), nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	s := string(b)
	// len(`"23:59"`) == 7
	if len(s) != 7 {
		return TimeParseError
	}
	ret, err := time.Parse(timeLayout, s[1:6])
	if err != nil {
		return err
	}
	t.Time = ret
	return nil
}

type TimeRange [2]Time

// This is mostly borrowed from foursquare, not schema.org
type Hours struct {
	Days []time.Weekday `json:"days"`
	// The indexes in the Open field match the indexes in the Days field.
	// Together they map when the place is open.
	Open []TimeRange `json:"open"`
}

type Place struct {
	Url         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`

	// Despite the "PostalAddress" type, this should be a physical address.
	Address              *PostalAddress  `json:"address,omitempty"`
	Location             *GeoCoordinates `json:"location,omitempty"`
	Rating               *Rating         `json:"rating,omitempty"`
	Hours                *Hours          `json:"hours,omitempty"`
	PhoneNumber          string          `json:"phone_number,omitempty"`
	FormattedPhoneNumber string          `json:"formatted_phone_number,omitempty"`
	GenericMetadata      `ogtag:",squash"`
}

func (p *Place) HasLocation() bool {
	return p.Location != nil &&
		p.Location.Latitude != nil &&
		p.Location.Longitude != nil
}

type PlaceCard struct {
	Card
	Place *Place `json:"place"`
}

func NewPlaceCard(webUrl string) *PlaceCard {
	return &PlaceCard{
		Card{
			CardType: PlaceType,
			WebUrl:   webUrl,
		},
		&Place{},
	}
}
