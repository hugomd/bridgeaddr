module github.com/fiatjaf/lightningaddr

go 1.16

replace github.com/fiatjaf/go-lnurl => /home/fiatjaf/comp/go-lnurl

require (
	github.com/fiatjaf/go-lnurl v1.4.0
	github.com/fiatjaf/lightningd-gjson-rpc v1.4.0
	github.com/fiatjaf/makeinvoice v0.0.0-20210812235429-f0615a4b9b34 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.2
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/rs/zerolog v1.23.0
	github.com/tidwall/gjson v1.8.1
	github.com/tidwall/sjson v1.1.7
)
