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

var help = `I am here to do all the boring stuff for you!
Here is what I can do:
/%s help - show this message
/%s label {name} - to add label
/%s merge - to merge the pull request
`

func GithubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Write([]byte("405 Method Not Allowed"))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	lgr := logger.New()
	lgr.Info("received webhook")
	xhub := r.Header.Get("X-Hub-Signature-256")
	if xhub == "" {
		lgr.Info("X-Hub-Signature-256 was not provided")
		return
	}
	cnf := BuildConfig()
	data := read(r.Body)
	if res := decodeSha(cnf.WebhookSecret, data); res != xhub {
		lgr.Info("Payload was signed by untrustred source", "x-hub", xhub, "result", res)
		return
	}
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

	commentStr := gjson.Get(bodyAsStr, "comment.id").String()
	commentID, _ := strconv.Atoi(commentStr)

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

	prbot := PrBotFile{
		Version: "1.0.0",
		Use:     "bot",
	}
	fileContent, _, _, err := client.Repositories.GetContents(ctx, repo, name, ".prbot.yaml", nil)
	if err != nil {
		lgr.Info("failed to get .prbot.yaml file content, using default config", "error", err.Error())
	} else {
		content, err := fileContent.GetContent()
		if err != nil {
			lgr.Info("failed to get .prbot.yaml content, using default config", "error", err.Error())
		}

		prbot, err = UnmarshalPrBotFile([]byte(content))
		if err != nil {
			lgr.Info("failed to unmarshal .prbot.yaml, using default config", "error", err.Error())
		}

	}
	lines := strings.Split(comment, "\n")
	for _, l := range lines {
		tokens := strings.Split(l, " ")
		if len(tokens) == 1 {
			continue
		}
		root := tokens[0]
		if root != fmt.Sprintf("/%s", prbot.Use) {
			continue
		}

		cmd := tokens[1]
		switch cmd {
		case "help":
			_, _, err := client.Issues.EditComment(ctx, repo, name, int64(commentID), &github.IssueComment{
				Body: github.String(fmt.Sprintf(help, prbot.Use, prbot.Use, prbot.Use)),
			})
			if err != nil {
				lgr.Info("failed to edit comment", "error", err.Error())
				w.WriteHeader(500)
				return
			}
		case "label":
			if len(tokens) < 3 {
				continue
			}
			labels := tokens[2:]
			_, _, err := client.Issues.AddLabelsToIssue(ctx, repo, name, issue, labels)
			if err != nil {
				lgr.Info("failed to add issues", "error", err.Error())
				w.WriteHeader(500)
				return
			}
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
			lgr.Info("unknown command", "cmd", cmd)
			continue
		}
	}
	return
}
