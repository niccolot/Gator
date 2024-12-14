package rss

import (
	"context"
	"database/sql"
	"encoding/xml"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/niccolot/BlogAggregator/internal/database"
	"github.com/niccolot/BlogAggregator/internal/state"
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

func ScrapeFeeds(s *state.State, ctx context.Context) error {
	feedToFetch, errFeed := s.Db.GetNextFeedToFetch(ctx)
	if errFeed != nil {
		return errFeed
	}

	// sql (possibly) null TIMESTAMP wants this kind of time object
	nullableTime := sql.NullTime{
		Time: time.Now(),
		Valid: true,
	}

	fetchedPars := &database.MarkFeedFetchedParams{
		ID: feedToFetch.ID,
		LastFetchedAt: nullableTime,
	}

	errMark := s.Db.MarkFeedFetched(ctx, *fetchedPars)
	if errMark != nil {
		return errMark
	}	

	feed, errFetching := FetchFeed(ctx, feedToFetch.Url)
	if errFetching != nil {
		return errFetching
	}

	for _, item := range(feed.Channel.Item) {
		pubTime, errPubtime := time.Parse(time.RFC1123, item.PubDate)
		if errPubtime != nil {
			return errPubtime
		}

		nullableTitle := sql.NullString{
			String: item.Title,
			Valid: true,
		}

		nullPubTime := sql.NullTime{
			Time: pubTime,
			Valid: true,
		}

		nullableDescription := sql.NullString{
			String: item.Description,
			Valid: true,
		}

		postPars := &database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: nullableTime.Time,
			UpdatedAt: nullableTime.Time,
			Title: nullableTitle,
			Url: item.Link,
			Description: nullableDescription,
			PublishedAt: nullPubTime,
			FeedID: feedToFetch.ID,
		}

		_, errPost := s.Db.CreatePost(ctx, *postPars)
		if errPost != nil {
			return errPost
		}
	}

	return nil
}