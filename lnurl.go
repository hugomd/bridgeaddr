package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/fiatjaf/go-lnurl"
	"github.com/gorilla/mux"
	nostr "github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/tidwall/sjson"
)

func handleLNURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	username := mux.Vars(r)["username"]
	domain := r.Host

	var nostr_pubkey string
	if v, err := net.LookupTXT("_nostr_pubkey." + domain); err == nil && len(v) > 0 {
		nostr_pubkey = v[0]
	}

	if err != nil {
		log.Info().Err(err).Msg("Failed to retrieve nostr pubkey")
		return
	}

	log.Info().Str("username", username).Str("domain", domain).
		Msg("got lnurl request")

	if amount := r.URL.Query().Get("amount"); amount == "" {
		// check if the receiver accepts comments
		var commentLength int64 = 0
		if v, err := net.LookupTXT("_webhook." + domain); err == nil && len(v) > 0 {
			commentLength = 500
		}

		json.NewEncoder(w).Encode(lnurl.LNURLPayParams{
			LNURLResponse:   lnurl.LNURLResponse{Status: "OK"},
			Callback:        fmt.Sprintf("https://%s/.well-known/lnurlp/%s", domain, username),
			MinSendable:     1000,
			MaxSendable:     100000000,
			EncodedMetadata: makeMetadata(username, domain),
			CommentAllowed:  commentLength,
			Tag:             "payRequest",
			AllowsNostr:     nostr_pubkey != "",
			NostrPubkey:     nostr_pubkey,
		})

	} else {
		msat, err := strconv.Atoi(amount)
		if err != nil {
			json.NewEncoder(w).Encode(lnurl.ErrorResponse("amount is not integer"))
			return
		}

		zapReqStr, _ := url.QueryUnescape(r.URL.Query().Get("nostr"))

		// TODO: better zap validation
		var zapReq nostr.Event
		if err := json.Unmarshal([]byte(zapReqStr), &zapReq); err != nil {
			log.Warn().Err(err).Msg("Failed to unmarshal zap request")
			return
		}
		valid, err := zapReq.CheckSignature()
		if !valid {
			log.Info().Msg("Zap request signature invalid")
			return
		}

		log.Info().Interface("zap request", zapReq).Msg("Parsed zap request")

		bolt11, err := makeInvoice(username, domain, msat, zapReq.String())
		if err != nil {
			json.NewEncoder(w).Encode(
				lnurl.ErrorResponse("failed to create invoice: " + err.Error()))
			return
		}

		json.NewEncoder(w).Encode(lnurl.LNURLPayValues{
			LNURLResponse: lnurl.LNURLResponse{Status: "OK"},
			PR:            bolt11,
			Routes:        make([][]interface{}, 0),
			Disposable:    lnurl.FALSE,
			SuccessAction: lnurl.Action("Payment received!", ""),
		})

		go func() {
			inv, err := decodepay.Decodepay(bolt11)
			if err != nil {
				return
			}
			WaitForZap(inv.PaymentHash, domain, zapReq)
		}()

		// send webhook
		go func() {
			if v, err := net.LookupTXT("_webhook." + domain); err == nil && len(v) > 0 {
				body, _ := sjson.Set("{}", "pr", bolt11)
				body, _ = sjson.Set(body, "amount", msat)
				if comment := r.URL.Query().Get("comment"); comment != "" {
					body, _ = sjson.Set(body, "comment", comment)
				}

				(&http.Client{Timeout: 5 * time.Second}).
					Post(v[0], "application/json", bytes.NewBufferString(body))
			}
		}()
	}
}
