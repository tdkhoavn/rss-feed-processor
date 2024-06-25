package main

import (
	"encoding/xml"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	XMLName     xml.Name `xml:"channel"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Items       []Item   `xml:"item"`
}

type Item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	PubDate     string   `xml:"pubDate"`
	Guid        string   `xml:"guid"`
}

func fetchRSS(feedURL string) (RSS, error) {
	resp, err := http.Get(feedURL)
	if err != nil {
		return RSS{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return RSS{}, err
	}

	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		return RSS{}, err
	}

	for i := range rss.Channel.Items {
		// PubDate is from format like this string: "Tue, 31 October 2023 10:00:00 +0000"
		// we need convert it to RFC-822 format
		// TODO: convert the date to RFC-822 format
		//pubDate, err := time.Parse(time.RFC822, rss.Channel.Items[i].PubDate)

		rss.Channel.Items[i].Guid = uuid.New().String()
	}

	return rss, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	feedURL := "https://feeds.feedburner.com/blogspot/amDG"
	rss, err := fetchRSS(feedURL)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: err.Error()}, nil
	}

	rssBytes, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: err.Error()}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/rss+xml"},
		Body:       string(rssBytes),
	}, nil
}

func main() {
	lambda.Start(handler)
}
