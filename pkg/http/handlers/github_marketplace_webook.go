package handlers

import (
	"net/http"

	"github.com/olegsu/go-tools/pkg/logger"
	"github.com/olegsu/pull-requests-bot/pkg/config"
)

func GithubMarketplaceWebhook(cnf config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			w.Write([]byte("405 Method Not Allowed"))
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		lgr := logger.New()
		xhub := r.Header.Get("X-Hub-Signature-256")
		if xhub == "" {
			lgr.Info("X-Hub-Signature-256 was not provided")
			return
		}
		data := read(r.Body)
		if res := decodeSha(cnf.MarketplaceWebhookSecret, data); res != xhub {
			lgr.Info("Marketplace payload was signed by untrustred source", "x-hub", xhub, "result", res)
			return
		}
		lgr.Info("received marketplace webhook", "payload", string(data))
	}

}
