package github

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/emmahsax/go-git-helper/internal/configfile"
)

type GitHub struct {
	Debug bool
}

type Response struct {
	HtmlURL string `json:"html_url"`
	Message string `json:"message"`
	Errors  []struct {
		Resource string `json:"resource"`
		Code     string `json:"code"`
		Message  string `json:"message"`
	} `json:"errors"`
}

func NewGitHub(debug bool) *GitHub {
	return &GitHub{
		Debug: debug,
	}
}

func (c *GitHub) CreatePullRequest(repoName string, options map[string]interface{}) interface{} {
	return c.run(repoName, "POST", "/repos/"+repoName+"/pulls", options)
}

func (c *GitHub) run(username, requestType, curlURL string, payload map[string]interface{}) interface{} {
	var result Response
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		if c.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return result
	}
	req, err := http.NewRequest(requestType, "https://api.github.com"+curlURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		if c.Debug {
			debug.PrintStack()
		}
		log.Fatal(err)
		return result
	}
	cf := configfile.NewConfigFile(c.Debug)
	req.Header.Set("Authorization", "token "+cf.GitHubToken())
	req.Header.Set("Content-Type", "application/json")

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
