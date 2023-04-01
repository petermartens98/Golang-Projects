package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	// Prompt user for website URL
	var website string
	fmt.Print("Enter website URL: ")
	fmt.Scanln(&website)

	// Make HTTP GET request
	response, err := http.Get(website)
	if err != nil {
		fmt.Println("Error fetching page:", err)
		return
	}
	defer response.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Convert byte slice to string
	content := string(body)

	// Find all links on the page
	links := getLinks(content)

	// Print links
	for _, link := range links {
		fmt.Println(link)
	}
}

func getLinks(content string) []string {
	links := make([]string, 0)

	// Find all occurrences of '<a href="...">' in the content
	startTag := "<a href=\""
	startIndex := 0
	for {
		index := strings.Index(content[startIndex:], startTag)
		if index == -1 {
			break
		}
		index += startIndex + len(startTag)

		// Find end of link URL
		endIndex := strings.Index(content[index:], "\"")
		if endIndex == -1 {
			break
		}
		endIndex += index

		// Extract link URL and append to list of links
		link := content[index:endIndex]
		links = append(links, link)

		startIndex = endIndex
	}

	return links
}
