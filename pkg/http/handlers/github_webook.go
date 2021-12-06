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
* "/? help"
* "/? label {name}" - adds a label, creating new one if not exists
* "/? merge" - squash merge the pull request
* "/? workflow {name}" - uses workflow dispatch event api to trigger worklfow. The workflow must have "on: workflow_dispatch".
`

func GithubWebhook(cnf config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Write([]byte("405 Method Not Allowed"))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		lgr := logger.New()
		if err := githubWebhook(r.Context(), lgr, cnf, read(r.Body), r.Header); err != nil {
			lgr.Error(err, "failed to handler webhook")
			w.WriteHeader(500)
			return
		}

	}
}

func githubWebhook(ctx context.Context, lgr *logger.Logger, cnf config.Config, data []byte, headers http.Header) error {
	lgr.Info("received webhook")

	if !cnf.SkipPayloadValidation {
		if !isValidBody(headers.Get("X-Hub-Signature-256"), data, cnf.WebhookSecret) {
			lgr.Info("Payload was signed by untrustred source")
			return nil
		}
	}
	body, err := UnmarshalGithubWebhookBody(data)
	if err != nil {
		lgr.Info("failed to unmarshal body into struct", "error", err)
		return err
	}
	if body.Sender.Type != "User" {
		return nil
	}

	if body.Action != "created" {
		return nil
	}

	comment := body.Comment.Body
	if comment == "" {
		return nil
	}
	installation := body.Installation.ID
	itr, err := ghinstallation.New(http.DefaultTransport, int64(cnf.ApplicationID), installation, cnf.ApplicationPrivateKey)
	if err != nil {
		lgr.Info("failed to create installation transport", "error", err.Error())
		return err
	}
	// Use installation transport with client.
	client := github.NewClient(&http.Client{Transport: itr})
	gh := NewClient(WithGithubClient(client))

	authorized, err := isAuthorized(ctx, gh, body)
	if err != nil {
		return err
	}
	if !authorized {
		lgr.Info("user is not allowed to perform the command", "user", body.Sender.Login)
		return nil
	}

	prbot := getPRBotDefinition(ctx, lgr, cnf, gh, body)
	if err := processComment(ctx, lgr, body, gh, prbot); err != nil {
		lgr.Info("comment process failed", "errors", err.Error())
		return nil
	}
	return nil
}

func isAuthorized(ctx context.Context, gh GithubClient, hook GithubWebhookBody) (bool, error) {
	repo := hook.Repository.Owner.Login
	name := hook.Repository.Name
	repository, _, err := gh.GetRepository(ctx, repo, name)
	if err != nil {
		return false, fmt.Errorf("failed to get repository %s: %w", repo+"/"+name, err)
	}
	allowed := false
	if repository.Owner != nil && repository.Owner.ID != nil {
		if *repository.Owner.ID == hook.Sender.ID {
			allowed = true
		}
	}

	if repository.Organization != nil {
		members, _, err := gh.ListOrganizationMembers(ctx, repo, &github.ListMembersOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to list organization (%s) members: %s", repo, err)
		}
		for _, m := range members {
			if m != nil && m.ID != nil && *m.ID == hook.Sender.ID {
				allowed = true
			}
		}
	}
	return allowed, nil
}

func getPRBotDefinition(ctx context.Context, lgr *logger.Logger, cnf config.Config, gh GithubClient, hook GithubWebhookBody) PrBotFile {
	repo := hook.Repository.Owner.Login
	name := hook.Repository.Name
	prbot := PrBotFile{
		Version: "1.0.0",
		Use:     cnf.DefaultRootCmd,
	}
	fileContent, _, _, err := gh.GetFileContent(ctx, repo, name, ".prbot.yaml", nil)
	if err != nil {
		lgr.Info("failed to get .prbot.yaml file content, using default config", "error", err.Error())
		return prbot
	}
	content, err := fileContent.GetContent()
	if err != nil {
		lgr.Info("failed to get .prbot.yaml content, using default config", "error", err.Error(), "content", content)
		return prbot
	}

	prbot, err = UnmarshalPrBotFile([]byte(content))
	if err != nil {
		lgr.Info("failed to unmarshal .prbot.yaml, using default config", "error", err.Error(), "content", content)
		return prbot
	}

	return prbot
}

func processComment(ctx context.Context, lgr *logger.Logger, body GithubWebhookBody, gh GithubClient, prbot PrBotFile) error {
	errs := []error{}
	lines := strings.Split(body.Comment.Body, "\n")
	for _, l := range lines {
		if l == "\r" {
			continue
		}
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
			if err := onHelp(ctx, gh, body, prbot); err != nil {
				errs = append(errs, err)
			}
		case "label":
			lgr.Info("labeling", "tokens", tokens)
			if err := onLabel(ctx, lgr, gh, body, prbot, tokens); err != nil {
				errs = append(errs, err)
			}
		case "merge":
			lgr.Info("merging")
			if err := onMerge(ctx, gh, body, prbot); err != nil {
				errs = append(errs, err)
			}
		case "workflow":
			file := tokens[2]
			inputs := []string{}
			if len(tokens) >= 3 {
				inputs = tokens[3:]
			}
			lgr.Info("starting workflow", "file", file)
			if err := onWorkflow(ctx, lgr, gh, body, prbot, file, inputs); err != nil {
				errs = append(errs, err)
			}
		default:
			lgr.Info("unknown command", "cmd", cmd)
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

func onHelp(ctx context.Context, gh GithubClient, body GithubWebhookBody, prbot PrBotFile) error {
	_, _, err := gh.EditIssueComment(ctx, body.Repository.Owner.Login, body.Repository.Name, int64(body.Comment.ID), &github.IssueComment{
		Body: github.String(fmt.Sprintf(help, prbot.Use, prbot.Use, prbot.Use)),
	})
	return err
}

func onLabel(ctx context.Context, lgr *logger.Logger, gh GithubClient, body GithubWebhookBody, prbot PrBotFile, tokens []string) error {
	if len(tokens) < 2 {
		lgr.Info("not enough arguments to label")
		return nil
	}
	labels := tokens[2:]
	lgr.Info("adding labels", "labels", labels)
	_, _, err := gh.AddLabelsToIssue(ctx, body.Repository.Owner.Login, body.Repository.Name, int(body.Issue.Number), labels)
	return err
}

func onMerge(ctx context.Context, gh GithubClient, body GithubWebhookBody, prbot PrBotFile) error {
	message := body.Issue.Title
	_, _, err := gh.MergePullRequest(ctx, body.Repository.Owner.Login, body.Repository.Name, int(body.Issue.Number), message, &github.PullRequestOptions{
		MergeMethod: "squash",
	})
	return err
}

func onWorkflow(ctx context.Context, lgr *logger.Logger, gh GithubClient, body GithubWebhookBody, prbot PrBotFile, file string, inputs []string) error {
	repo := body.Repository.Owner.Login
	name := body.Repository.Name
	j := map[string]interface{}{}

	if len(inputs) > 0 {
		for _, in := range inputs {
			kv := strings.Split(in, "=")
			if (len(kv)) == 2 {
				j[kv[0]] = kv[1]
			}
		}
	}
	_, err := gh.CreateWorkflowDispatchEventByFileName(ctx, repo, name, file, github.CreateWorkflowDispatchEventRequest{
		Ref:    body.Repository.DefaultBranch,
		Inputs: j,
	})
	if err != nil {
		return err
	}
	_, _, err = gh.CreateIssueComment(ctx, repo, name, int(body.Issue.Number), &github.IssueComment{
		Body: github.String(fmt.Sprintf("Workflow %s started", file)),
	})
	if err != nil {
		return err
	}
	return nil
}
