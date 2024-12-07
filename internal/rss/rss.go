package rss

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, errReq := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if errReq != nil {
		return nil, errReq
	}

	req.Header.Add("User-Agent", "gator")

	client := &http.Client{}
	resp, errResp := client.Do(req)
	if errResp != nil {
		return nil, errResp
	}

	body, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return nil, errRead
	}

	feedStruct := &RSSFeed{}
	errUnmarshal := xml.Unmarshal(body, feedStruct)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}
	
	return feedStruct, nil
}