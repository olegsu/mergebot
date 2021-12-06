package handlers

import (
	"context"

	"github.com/google/go-github/v41/github"
)

type (
	GithubClient interface {
		GetRepository(ctx context.Context, owner string, name string) (*github.Repository, *github.Response, error)
		ListOrganizationMembers(ctx context.Context, org string, opt *github.ListMembersOptions) ([]*github.User, *github.Response, error)
		GetFileContent(ctx context.Context, owner string, name string, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
		EditIssueComment(ctx context.Context, owner string, name string, issue int64, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
		CreateIssueComment(ctx context.Context, owner string, name string, issue int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
		AddLabelsToIssue(ctx context.Context, owner string, name string, issue int, labels []string) ([]*github.Label, *github.Response, error)
		MergePullRequest(ctx context.Context, owner string, name string, issue int, commitMessage string, opt *github.PullRequestOptions) (*github.PullRequestMergeResult, *github.Response, error)
		CreateWorkflowDispatchEventByFileName(ctx context.Context, owner, repo, workflowFileName string, event github.CreateWorkflowDispatchEventRequest) (*github.Response, error)
	}

	gh struct {
		clinet *github.Client
	}

	Option func(g *gh)
)

func NewClient(opts ...Option) GithubClient {
	g := &gh{}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

func WithGithubClient(c *github.Client) Option {
	return func(g *gh) {
		g.clinet = c
	}
}

func (g *gh) GetRepository(ctx context.Context, owner string, name string) (*github.Repository, *github.Response, error) {
	return g.clinet.Repositories.Get(ctx, owner, name)
}
func (g *gh) ListOrganizationMembers(ctx context.Context, org string, opt *github.ListMembersOptions) ([]*github.User, *github.Response, error) {
	return g.clinet.Organizations.ListMembers(ctx, org, opt)
}
func (g *gh) GetFileContent(ctx context.Context, owner string, name string, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	return g.clinet.Repositories.GetContents(ctx, owner, name, path, opt)
}
func (g *gh) EditIssueComment(ctx context.Context, owner string, name string, issue int64, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	return g.clinet.Issues.EditComment(ctx, owner, name, issue, comment)
}
func (g *gh) CreateIssueComment(ctx context.Context, owner string, name string, issue int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	return g.clinet.Issues.CreateComment(ctx, owner, name, issue, comment)
}
func (g *gh) AddLabelsToIssue(ctx context.Context, owner string, name string, issue int, labels []string) ([]*github.Label, *github.Response, error) {
	return g.clinet.Issues.AddLabelsToIssue(ctx, owner, name, issue, labels)
}
func (g *gh) MergePullRequest(ctx context.Context, owner string, name string, issue int, commitMessage string, opt *github.PullRequestOptions) (*github.PullRequestMergeResult, *github.Response, error) {
	return g.clinet.PullRequests.Merge(ctx, owner, name, issue, commitMessage, opt)
}
func (g *gh) CreateWorkflowDispatchEventByFileName(ctx context.Context, owner, repo, workflowFileName string, event github.CreateWorkflowDispatchEventRequest) (*github.Response, error) {
	return g.clinet.Actions.CreateWorkflowDispatchEventByFileName(ctx, owner, repo, workflowFileName, event)
}
