package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	b64 "encoding/base64"
	"encoding/hex"
	"github.com/gorilla/websocket"
	nostr "github.com/nbd-wtf/go-nostr"
	"net"
	"net/http"
)

type LNDResponse struct {
	Result Invoice `json:"result"`
}

type Invoice struct {
	Memo           string `json:"memo"`
	State          string `json:"state"`
	SettleDate     int64  `json:"settle_date,string"`
	CreationDate   int64  `json:"creation_date,string"`
	PaymentRequest string `json:"payment_request"`
	PreImage       string `json:"r_preimage"`
}

func WaitForZap(r_hash, domain string, zapReq nostr.Event) {
	log.Info().Str("r_hash", r_hash).Msg("Waiting for Zap!")

	var macaroon string
	if v, err := net.LookupTXT("_macaroon." + domain); err == nil && len(v) > 0 {
		macaroon = v[0]
	}

	var host string
	if v, err := net.LookupTXT("_host." + domain); err == nil && len(v) > 0 {
		host = v[0]
	}

	privateKey := os.Getenv("NOSTR_KEY")
	publicKey, err := nostr.GetPublicKey(privateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get public key")
	}

	r_hash_bytes, err := hex.DecodeString(r_hash)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to decode r_hash")
	}

	uEnc := b64.URLEncoding.EncodeToString(r_hash_bytes)

	formatted := fmt.Sprintf("%s/v2/invoices/subscribe/%s?method=GET", strings.Replace(host, "https", "wss", 1), uEnc)
	authHeader := http.Header{
		"Grpc-Metadata-Macaroon": []string{macaroon},
	}
	conn, _, err := websocket.DefaultDialer.Dial(formatted, authHeader)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to dial")
	}

	log.Info().Msg("Waiting for responses on websocket")
	for {
		var response LNDResponse
		err := conn.ReadJSON(&response)
		if err != nil {
			log.Info().Err(err).Msg("Failed to read JSON")
			break
		}
		fmt.Println(response)
		if response.Result.State == "SETTLED" {
			zapNote := makeZapNote(privateKey, publicKey, response.Result, zapReq)
			fmt.Println(zapNote)

			relays := zapReq.Tags.GetAll([]string{"relays"})[0]

			for _, url := range relays[1:] {
				log.Info().Str("relay", url).Msg("Connecting to relay")
				relay, err := nostr.RelayConnect(context.Background(), url)
				if err != nil {
					log.Info().Err(err).Msg("Failed to connect to relay")
					continue
				}
				if _, err := relay.Publish(context.Background(), zapNote); err != nil {
					log.Info().Err(err).Msg("Failed to publish event to relay")
					continue
				}
			}
		}
	}
}

func makeZapNote(privateKey, publicKey string, invoice Invoice, zapReq nostr.Event) nostr.Event {
	event := nostr.Event{
		PubKey:    publicKey,
		CreatedAt: nostr.Timestamp(invoice.SettleDate),
		Kind:      nostr.KindZap,
		Tags: nostr.Tags{
			*zapReq.Tags.GetFirst([]string{"p"}),
			*zapReq.Tags.GetFirst([]string{"e"}),
			nostr.Tag{"bolt11", invoice.PaymentRequest},
			nostr.Tag{"description", invoice.Memo},
			nostr.Tag{"preimage", invoice.PreImage},
		},
	}

	event.Sign(privateKey)
	return event
}
