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

type GitLabClient struct {
	Debug bool
}

type Response struct {
	Message []string `json:"message"`
	WebURL  string   `json:"web_url"`
}

func NewGitLabClient(debug bool) *GitLabClient {
	return &GitLabClient{
		Debug: debug,
	}
}

func (c *GitLabClient) CreateMergeRequest(projectName string, options map[string]string) interface{} {
	return c.run("POST", fmt.Sprintf("/projects/%s/merge_requests%s", c.urlEncode(projectName), c.formatOptions(options)))
}

func (c *GitLabClient) run(requestType, curlURL string) interface{} {
	var result Response
	req, err := http.NewRequest(requestType, fmt.Sprintf("https://gitlab.com/api/v4%s", curlURL), nil)
	if err != nil {
		if c.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return result
	}

	cf := configfile.NewConfigFileClient(c.Debug)
	req.Header.Set("PRIVATE-TOKEN", cf.GitLabToken())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if c.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return result
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		if c.Debug {
			debug.PrintStack()
		}
		log.Fatal("Cannot unmarshal JSON")
		return err
	}

	return result
}

func (c *GitLabClient) urlEncode(input string) string {
	return url.PathEscape(input)
}

func (c *GitLabClient) formatOptions(options map[string]string) string {
	var optsAsString string
	for key, value := range options {
		if value != "" {
			optsAsString += fmt.Sprintf("%s=%s&", key, c.urlEncode(value))
		}
	}
	optsAsString = strings.TrimSuffix(optsAsString, "&")
	if optsAsString != "" {
		optsAsString = "?" + optsAsString
	}
	return optsAsString
}
