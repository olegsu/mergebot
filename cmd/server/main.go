package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/olegsu/pull-requests-bot/pkg/config"
	"github.com/olegsu/pull-requests-bot/pkg/http/handlers"
)

func main() {

	fmt.Println("Starting server")
	cnf := config.BuildConfig()
	http.HandleFunc("/github/webhook", handlers.GithubWebhook(cnf))
	http.HandleFunc("/github/marketplace", handlers.GithubMarketplaceWebhook(cnf))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}

func read(reader io.ReadCloser) []byte {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil
	}

	return b
}
