package rss

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
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

func FetchAndStoreFeed(s *state.State, feedToFetch *database.Feed, ctx context.Context) error {
	select {
	case <- ctx.Done():
		return fmt.Errorf("warning: fetch time exceeded time between request, timeout")
	default:
		feed, err := fetchFeed(ctx, feedToFetch.Url)
		if err != nil {
			log.Printf("Error: failed to fetch feed '%s': %v\n", feedToFetch.Url, err)
			return err
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

		err = s.Db.MarkFeedFetched(ctx, *fetchedPars)
		if err != nil {
			log.Printf("Error: failed to mark feed '%s' as fetched: %v", feedToFetch.Url, err)
			return err
		}

		for _, item := range(feed.Channel.Item) {
			processFeedItem(s, feedToFetch, &item, nullableTime.Time)
		}

		return err
	}	
}

func processFeedItem(s *state.State, feedToFetch *database.Feed, item *RSSItem, fetchTime time.Time) {
	pubTime, errTime := parseTime(item.PubDate)
	if errTime != nil {
		log.Printf("Warning: couldn't parse time for post '%s': %v\n", item.Title, errTime)
	}

	nullableTitle := sql.NullString{
		String: item.Title,
		Valid: true,
	}

	nullPubTime := sql.NullTime{
		Time: pubTime,
		Valid: true,
	}
	
	nullableDescription := getDescription(item)
	
	postPars := &database.CreatePostParams{
		ID: uuid.New(),
		CreatedAt: fetchTime,
		UpdatedAt: fetchTime,
		Title: nullableTitle,
		Url: item.Link,
		Description: *nullableDescription,
		PublishedAt: nullPubTime,
		FeedID: feedToFetch.ID,
	}

	_, errPost := s.Db.CreatePost(context.Background(), *postPars)
	if errPost != nil {
		log.Printf("Warning: failed to save post '%s' in database: %v\n", 
			nullableTitle.String, errPost)
	}
}

func ScrapeFeeds(s *state.State, ctx context.Context, batchSize int32) ([]database.Feed, error) {
	feeds, err := s.Db.GetNextFeedsToFetch(ctx, batchSize)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func getDescription(item *RSSItem) *sql.NullString {
	/*
	some blogs do not have a proper 'description' rss field
	but contains just some html, these cases are discarded
	*/
	var nullableDescription sql.NullString
	hasLeftHTML := strings.Contains(item.Description, "<")
	hasRightHTML := strings.Contains(item.Description, ">")
	hasHTML := hasLeftHTML && hasRightHTML

	if !hasHTML {
		nullableDescription = sql.NullString{
			String: item.Description,
			Valid: true,
		}
	} else {
		nullableDescription = sql.NullString{Valid: false}
	}

	return &nullableDescription
}

func parseTime(timeStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC822,
		time.RFC822Z,
		// add more formats as needed
	}

	for _, format := range(formats) {
		if t, errTime := time.Parse(format, timeStr); errTime == nil {
			return t, errTime
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}