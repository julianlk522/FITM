package handler

import (
	"io"

	"golang.org/x/net/html"
)

type HTMLMeta struct {
	Title         string
	Description   string
	OGTitle       string
	OGDescription string
	OGImage       string
	OGAuthor      string
	OGPublisher   string
	OGSiteName    string
}

func MetaFromHTMLTokens(resp io.Reader) (hm HTMLMeta) {
	z := html.NewTokenizer(resp)

	for {
		tt := z.Next()
		switch tt {
			case html.ErrorToken:
				return
			case html.SelfClosingTagToken, html.StartTagToken:
				t := z.Token()
				if t.Data == "meta" {
					AssignTokenPropertyToHTMLMeta(t, &hm)
				}
		}
	}
}

func AssignTokenPropertyToHTMLMeta(t html.Token, hm *HTMLMeta) {
	desc, ok := ExtractMetaProperty(t, "description")
	if ok {
		hm.Description = desc
	}

	ogTitle, ok := ExtractMetaProperty(t, "og:title")
	if ok {
		hm.OGTitle = ogTitle
	}

	ogDesc, ok := ExtractMetaProperty(t, "og:description")
	if ok {
		hm.OGDescription = ogDesc
	}

	ogImage, ok := ExtractMetaProperty(t, "og:image")
	if ok {
		hm.OGImage = ogImage
	}

	ogAuthor, ok := ExtractMetaProperty(t, "og:author")
	if ok {
		hm.OGAuthor = ogAuthor
	}

	ogPublisher, ok := ExtractMetaProperty(t, "og:publisher")
	if ok {
		hm.OGPublisher = ogPublisher
	}

	ogSiteName, ok := ExtractMetaProperty(t, "og:site_name")
	if ok {
		hm.OGSiteName = ogSiteName
	}
}

func ExtractMetaProperty(t html.Token, prop string) (content string, ok bool) {
	for _, attr := range t.Attr {
		if attr.Key == "property" && attr.Val == prop {
			ok = true
		}

		if attr.Key == "content" {
			content = attr.Val
		}
	}

	return
}