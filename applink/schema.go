// Package implementing applink, documented here: http://applinks.org/documentation/
package applink

type Ios struct {
	Url string `json:"url" ogtag:"al:ios:url"`
	// The only evidence i have that this is a string is here:
	// https://github.com/BoltsFramework/Bolts-iOS/blob/b72d5f2d6e0c418beea0a48da540a9eaf0768c0b/Bolts/iOS/BFAppLinkTarget.h#L28
	AppStoreId string `json:"app_store_id,omitempty" ogtag:"al:ios:app_store_id"`
	AppName    string `json:"app_name,omitempty" ogtag:"al:ios:app_name"`
}

type Iphone struct {
	Url string `json:"url" ogtag:"al:iphone:url"`
	// The only evidence i have that this is a string is here:
	// https://github.com/BoltsFramework/Bolts-iOS/blob/b72d5f2d6e0c418beea0a48da540a9eaf0768c0b/Bolts/iOS/BFAppLinkTarget.h#L28
	AppStoreId string `json:"app_store_id,omitempty" ogtag:"al:iphone:app_store_id"`
	AppName    string `json:"app_name,omitempty" ogtag:"al:iphone:app_name"`
}

type Ipad struct {
	Url string `json:"url" ogtag:"al:ipad:url"`
	// The only evidence i have that this is a string is here:
	// https://github.com/BoltsFramework/Bolts-iOS/blob/b72d5f2d6e0c418beea0a48da540a9eaf0768c0b/Bolts/iOS/BFAppLinkTarget.h#L28
	AppStoreId string `json:"app_store_id,omitempty" ogtag:"al:ipad:app_store_id"`
	AppName    string `json:"app_name,omitempty" ogtag:"al:ipad:app_name"`
}

type Android struct {
	Url     string `json:"url,omitempty" ogtag:"al:android:url"`
	Package string `json:"package,omitempty"`
	Class   string `json:"class,omitempty"`
	AppName string `json:"app_name,omitempty"`
}

type Windows struct {
	Url     string `json:"url"`
	AppId   string `json:"app_id,omitempty"`
	AppName string `json:"app_name,omitempty"`
}

type Web struct {
	Url            string `json:"url,omitempty"`
	ShouldFallback bool   `json:"should_fallback,omitempty"`
}

type AppLink struct {
	Ios     *Ios     `json:"ios,omitempty" ogtag:",fill"`
	Iphone  *Iphone  `json:"iphone,omitempty" ogtag:",fill"`
	Ipad    *Ipad    `json:"ipad,omitempty" ogtag:",fill"`
	Android *Android `json:"android,omitempty" ogtag:",fill"`
	// TODO: Windows/web support
	WindowsPhone     *Windows `json:"windows_phone,omitempty"`
	Windows          *Windows `json:"windows,omitempty"`
	WindowsUniversal *Windows `json:"windows_universal,omitempty"`
	Web              *Web     `json:"web,omitempty"`
}
