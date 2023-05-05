module github.com/fiatjaf/bridgeaddr

go 1.16

require (
	github.com/fiatjaf/go-lnurl v1.12.1
	github.com/fiatjaf/makeinvoice v1.5.4
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.3
	github.com/nbd-wtf/go-nostr v0.17.3
	github.com/nbd-wtf/ln-decodepay v1.5.1
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/rs/zerolog v1.23.0
	github.com/tidwall/sjson v1.1.7
)

replace github.com/fiatjaf/go-lnurl => github.com/hugomd/go-lnurl v0.0.0-20230505085824-f1adc5bfdb48
