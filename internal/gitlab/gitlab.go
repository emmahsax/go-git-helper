package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"

	"github.com/emmahsax/go-git-helper/internal/configfile"
)

type GitLabClient struct{}

type Response struct {
	Message []string `json:"message"`
	WebURL  string   `json:"web_url"`
}

func NewGitLabClient() *GitLabClient {
	return &GitLabClient{}
}

func (c *GitLabClient) CreateMergeRequest(projectName string, options map[string]string) interface{} {
	return c.run("POST", fmt.Sprintf("/projects/%s/merge_requests%s", urlEncode(projectName), formatOptions(options)))
}

func (c *GitLabClient) run(requestType, curlURL string) interface{} {
	var result Response
	req, err := http.NewRequest(requestType, fmt.Sprintf("https://gitlab.com/api/v4%s", curlURL), nil)
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
		return result
	}

	req.Header.Set("PRIVATE-TOKEN", configfile.GitLabToken())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
		return result
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		debug.PrintStack()
		log.Fatal("Cannot unmarshal JSON")
		return err
	}

	return result
}

func urlEncode(input string) string {
	return url.PathEscape(input)
}

func formatOptions(options map[string]string) string {
	var optsAsString string
	for key, value := range options {
		if value != "" {
			optsAsString += fmt.Sprintf("%s=%s&", key, urlEncode(value))
		}
	}
	optsAsString = strings.TrimSuffix(optsAsString, "&")
	if optsAsString != "" {
		optsAsString = "?" + optsAsString
	}
	return optsAsString
}
