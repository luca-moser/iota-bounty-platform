module github.com/luca-moser/iota-bounty-platform

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/apsdehal/go-logger v0.0.0-20190506062552-f85330a4b532
	github.com/c2h5oh/fresh v0.0.0-20160725163426-cf6b27f7aa64 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dpapathanasiou/go-recaptcha v0.0.0-20190121160230-be5090b17804
	github.com/facebookgo/inject v0.0.0-20180706035515-f23751cae28b
	github.com/facebookgo/structtag v0.0.0-20150214074306-217e25fb9691 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/gorilla/websocket v1.4.1
	github.com/howeyc/fsnotify v0.9.0 // indirect
	github.com/iotaledger/iota.go v1.0.0-beta.8.0.20190919192838-c5a0c534acb4
	github.com/labstack/echo v3.3.10+incompatible
	github.com/luca-moser/confbox v0.0.0-20190408145234-c3ced1d827ec
	github.com/mattn/go-colorable v0.1.2
	github.com/pkg/errors v0.8.1
	go.mongodb.org/mongo-driver v1.1.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gopkg.in/go-playground/webhooks.v5 v5.0.0-00010101000000-000000000000
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/inconshreveable/log15.v2 v2.0.0-20180818164646-67afb5ed74ec
)

replace gopkg.in/go-playground/webhooks.v5 => github.com/luca-moser/webhooks v0.1.1
