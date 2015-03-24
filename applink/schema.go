// Package implementing applink, documented here: http://applinks.org/documentation/
package applink

import "net/url"

type Ios struct {
	Url *url.URL `json:"url"`
	// The only evidence i have that this is a string is here:
	// https://github.com/BoltsFramework/Bolts-iOS/blob/b72d5f2d6e0c418beea0a48da540a9eaf0768c0b/Bolts/iOS/BFAppLinkTarget.h#L28
	AppStoreId string `json:"app_store_id,omitempty"`
	AppName    string `json:"app_name,omitempty"`
}

type Android struct {
	Url     *url.URL `json:"url,omitempty"`
	Package string   `json:"package"`
	Class   string   `json:"class,omitempty"`
	AppName string   `json:"app_name,omitempty"`
}

type Windows struct {
	Url     *url.URL `json:"url"`
	AppId   string   `json:"app_id,omitempty"`
	AppName string   `json:"app_name,omitempty"`
}

type Web struct {
	Url            *url.URL `json:"url,omitempty"`
	ShouldFallback bool     `json:"should_fallback,omitempty"`
}

type AppLink struct {
	Ios              *Ios     `json:"ios,omitempty"`
	Iphone           *Ios     `json:"iphone,omitempty"`
	Ipad             *Ios     `json:"ipad,omitempty"`
	Android          *Android `json:"android,omitempty"`
	WindowsPhone     *Windows `json:"windows_phone,omitempty"`
	Windows          *Windows `json:"windows,omitempty"`
	WindowsUniversal *Windows `json:"windows_universal,omitempty"`
	Web              *Web
}
