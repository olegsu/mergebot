package handlers

import (
	"context"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const userID = 1234

func Test_isAuthorized(t *testing.T) {
	type args struct {
		ctx  context.Context
		gh   func() GithubClient
		hook GithubWebhookBody
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "authorized for repo owner",
			args: args{
				ctx: context.Background(),
				gh: func() GithubClient {
					m := &MockGithubClient{}
					m.
						On("GetRepository", mock.Anything, mock.Anything, mock.Anything).
						Return(&github.Repository{
							Owner: &github.User{
								ID: github.Int64(userID),
							},
						}, nil, nil)

					return m
				},
				hook: GithubWebhookBody{Sender: Sender{ID: userID}},
			},
			wantErr: false,
			want:    true,
		},
		{
			name: "authorized member of org",
			args: args{
				ctx: context.Background(),
				gh: func() GithubClient {
					m := &MockGithubClient{}
					m.
						On("GetRepository", mock.Anything, mock.Anything, mock.Anything).
						Return(&github.Repository{
							Organization: &github.Organization{},
						}, nil, nil)

					m.
						On("ListOrganizationMembers", mock.Anything, mock.Anything, mock.Anything).
						Return([]*github.User{
							{
								ID: github.Int64(userID),
							},
						}, nil, nil)

					return m
				},
				hook: GithubWebhookBody{Sender: Sender{ID: userID}},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isAuthorized(tt.args.ctx, tt.args.gh(), tt.args.hook)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
