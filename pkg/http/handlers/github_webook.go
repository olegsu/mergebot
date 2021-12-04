package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v41/github"
	"github.com/olegsu/go-tools/pkg/logger"
	"github.com/olegsu/pull-requests-bot/pkg/config"
)

var help = `I am here to do all the boring stuff for you!
Here is what I can do:
/%s help - show this message
/%s label {name} - to add label
/%s merge - to merge the pull request
`

func GithubWebhook(cnf config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Write([]byte("405 Method Not Allowed"))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		lgr := logger.New()
		lgr.Info("received webhook")

		data := read(r.Body)
		if !cnf.SkipPayloadValidation {
			if !isValidBody(r.Header.Get("X-Hub-Signature-256"), data, cnf.WebhookSecret) {
				lgr.Info("Payload was signed by untrustred source")
				return
			}
		}
		body, err := UnmarshalGithubWebhookBody(data)
		if err != nil {
			lgr.Info("failed to unmarshal body into struct", "error", err)
			return
		}
		if body.Sender.Type != "User" {
			return
		}

		if body.Action != "created" {
			return
		}

		repo := body.Repository.Owner.Login
		name := body.Repository.Name
		comment := body.Comment.Body
		if comment == "" {
			return
		}
		installation := body.Installation.ID
		itr, err := ghinstallation.New(http.DefaultTransport, int64(cnf.ApplicationID), installation, cnf.ApplicationPrivateKey)
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
			Use:     "bot-local",
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
		if err := processComment(ctx, lgr, body, client, prbot); err != nil {
			return
		}
		return
	}
}

func processComment(ctx context.Context, lgr *logger.Logger, body GithubWebhookBody, client *github.Client, prbot PrBotFile) error {
	errs := []error{}
	lines := strings.Split(body.Comment.Body, "\n")
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
		lgr.Info("parsing command", "cmd", l)
		switch cmd {
		case "help":
			if err := onHelp(ctx, client, body, prbot); err != nil {
				errs = append(errs, err)
			}
		case "label":
			if err := onLabel(ctx, client, body, prbot, tokens); err != nil {
				errs = append(errs, err)
			}
		case "merge":
			if err := onMerge(ctx, client, body, prbot); err != nil {
				errs = append(errs, err)
			}
		case "workflow":
			file := tokens[2]
			if err := onWorkflow(ctx, client, body, prbot, file); err != nil {
				errs = append(errs, err)
			}
		default:
			continue
		}
	}

	if len(errs) > 0 {
		msg := strings.Builder{}
		for _, err := range errs {
			msg.WriteString(fmt.Sprintf("%s\n", err.Error()))
		}
		return errors.New(msg.String())
	}
	return nil
}

func onHelp(ctx context.Context, client *github.Client, body GithubWebhookBody, prbot PrBotFile) error {
	_, _, err := client.Issues.EditComment(ctx, body.Repository.Owner.Login, body.Repository.Name, int64(body.Comment.ID), &github.IssueComment{
		Body: github.String(fmt.Sprintf(help, prbot.Use, prbot.Use, prbot.Use)),
	})
	return err
}

func onLabel(ctx context.Context, client *github.Client, body GithubWebhookBody, prbot PrBotFile, tokens []string) error {
	if len(tokens) < 3 {
		return nil
	}
	labels := tokens[2:]
	_, _, err := client.Issues.AddLabelsToIssue(ctx, body.Repository.Owner.Login, body.Repository.Name, int(body.Issue.ID), labels)
	return err
}

func onMerge(ctx context.Context, client *github.Client, body GithubWebhookBody, prbot PrBotFile) error {
	message := body.Issue.Title
	_, _, err := client.PullRequests.Merge(ctx, body.Repository.Owner.Login, body.Repository.Name, int(body.Issue.ID), message, &github.PullRequestOptions{
		MergeMethod: "squash",
	})
	return err
}

func onWorkflow(ctx context.Context, client *github.Client, body GithubWebhookBody, prbot PrBotFile, file string) error {
	repo := body.Repository.Owner.Login
	name := body.Repository.Name
	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(ctx, repo, name, file, github.CreateWorkflowDispatchEventRequest{
		Ref: body.Repository.DefaultBranch,
	})
	if err != nil {
		return err
	}
	_, _, err = client.Issues.CreateComment(ctx, repo, name, int(body.Issue.ID), &github.IssueComment{
		Body: github.String(fmt.Sprintf("Workflow %s started", file)),
	})
	if err != nil {
		return err
	}
	return nil
}
