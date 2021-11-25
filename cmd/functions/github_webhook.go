// Package p contains an HTTP Cloud Function.
package functions

import (
	"net/http"

	"github.com/olegsu/mergebot/pkg/functions"
)

func GithubWebhook(w http.ResponseWriter, r *http.Request) {
	functions.GithubWebhook(w, r)
}
