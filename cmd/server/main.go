package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	functions "github.com/olegsu/mergebot"
)

func main() {

	fmt.Println("Starting server")

	http.HandleFunc("/github/webhook", functions.GithubWebhook)
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
