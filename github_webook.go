package function

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v40/github"
	"github.com/olegsu/go-tools/pkg/logger"
	"github.com/tidwall/gjson"
)

func GithubWebhook(w http.ResponseWriter, r *http.Request) {
	lgr := logger.New()
	lgr.Info("received webhook")
	cnf := BuildConfig()
	data := read(r.Body)
	bodyAsStr := string(data)
	if gjson.Get(bodyAsStr, "sender.type").String() != "User" {
		return
	}
	if gjson.Get(bodyAsStr, "action").String() != "created" {
		return
	}
	repo := gjson.Get(bodyAsStr, "repository.owner.login").String()
	name := gjson.Get(bodyAsStr, "repository.name").String()
	issueStr := gjson.Get(bodyAsStr, "issue.number").String()
	issue, _ := strconv.Atoi(issueStr)

	comment := gjson.Get(bodyAsStr, "comment.body").String()
	if comment == "" {
		return
	}

	installation, _ := strconv.Atoi(gjson.Get(bodyAsStr, "installation.id").String())
	itr, err := ghinstallation.New(http.DefaultTransport, int64(cnf.ApplicationID), int64(installation), cnf.ApplicationPrivateKey)
	if err != nil {
		lgr.Info("failed to create installation transport", "error", err.Error())
		w.WriteHeader(500)
		return
	}
	// Use installation transport with client.
	client := github.NewClient(&http.Client{Transport: itr})
	ctx := context.Background()

	fileContent, _, _, err := client.Repositories.GetContents(ctx, repo, name, ".mergebot.yaml", nil)
	if err != nil {
		lgr.Info("failed to get .mergebot.yaml file content", "error", err.Error())
		return
	}
	content, err := fileContent.GetContent()
	if err != nil {
		w.WriteHeader(500)
		return
	}

	mf, err := UnmarshalMergebotFile([]byte(content))
	if err != nil {
		w.WriteHeader(500)
		return
	}

	tokens := strings.Split(comment, " ")
	if len(tokens) == 1 {
		return
	}
	root := tokens[0]
	if root != fmt.Sprintf("/%s", mf.Use) {
		return
	}

	cmd := tokens[1]
	fmt.Println(cmd)
	switch cmd {
	case "label":
		if len(tokens) < 3 {
			return
		}
		labels := tokens[2:]
		_, _, err := client.Issues.AddLabelsToIssue(ctx, repo, name, issue, labels)
		if err != nil {
			lgr.Info("failed to add issues", "error", err.Error())
			w.WriteHeader(500)
			return
		}
		return
	case "merge":
		message := gjson.Get(bodyAsStr, "issue.title").String()
		_, _, err := client.PullRequests.Merge(ctx, repo, name, issue, message, &github.PullRequestOptions{
			MergeMethod: "squash",
		})
		if err != nil {
			lgr.Info("failed to merge", "error", err.Error())
			w.WriteHeader(500)
			return
		}
	default:
		return
	}
}
