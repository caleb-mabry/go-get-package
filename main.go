package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Package struct {
	Version string `json:"version,omitempty"`
	Release string `json:"release,omitempty"`
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ApiResponse := events.APIGatewayProxyResponse{}

	packageName := request.QueryStringParameters["name"]
	packageSlice := []Package{}

	url := fmt.Sprintf("https://pkg.go.dev/%s%s", packageName, "?tab=versions")
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".Versions-list").Each(func(i int, s *goquery.Selection) {
		versionIndex := -1
		s.Children().Each(func(_ int, s *goquery.Selection) {
			className, _ := s.Attr("class")
			if className == "Version-tag" {
				versionIndex += 1
				versionNumber := string(s.Find("a").Text())
				packageSlice = append(packageSlice, Package{Version: versionNumber})
			} else if className == "Version-commitTime" {
				dateCommitted := strings.TrimSpace(string(s.Text()))
				packageSlice[versionIndex].Release = dateCommitted
			} else if strings.Contains(className, "Version-details") {
				dateCommitted := strings.TrimSpace(string(s.Find(".Version-summary").Text()))
				packageSlice[versionIndex].Release = dateCommitted
			}
		})
	})

	data, err := json.Marshal(packageSlice)
	if err != nil {
		log.Fatal(err)
	}
	ApiResponse = events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
		Body:       string(data),
		StatusCode: 200}

	return ApiResponse, nil

}

func main() {
	lambda.Start(HandleRequest)
}
