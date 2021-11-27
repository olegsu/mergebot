package function

import (
	"net/http"

	"github.com/olegsu/go-tools/pkg/logger"
)

func GithubMarketplaceWebhook(w http.ResponseWriter, r *http.Request) {
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
	cnf := BuildConfig()
	if res := decodeSha(cnf.MarketplaceWebhookSecret, data); res != xhub {
		lgr.Info("Marketplace payload was signed by untrustred source", "x-hub", xhub, "result", res)
		return
	}
	lgr.Info("received marketplace webhook", "payload", string(data))
}
