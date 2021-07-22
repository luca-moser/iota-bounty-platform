module github.com/luca-moser/iota-bounty-platform

go 1.12

require (
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/inject v0.0.0-20180706035515-f23751cae28b
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/structtag v0.0.0-20150214074306-217e25fb9691 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/iotaledger/iota.go v1.0.0-beta.8.0.20190919192838-c5a0c534acb4
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.2.8 // indirect
	github.com/mattn/go-colorable v0.1.2
	github.com/pkg/errors v0.9.1
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v0.0.0-20170224212429-dcecefd839c4 // indirect
	go.mongodb.org/mongo-driver v1.5.1
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gopkg.in/go-playground/webhooks.v5 v5.0.0-00010101000000-000000000000
	gopkg.in/inconshreveable/log15.v2 v2.0.0-20180818164646-67afb5ed74ec
)

replace gopkg.in/go-playground/webhooks.v5 => github.com/luca-moser/webhooks v0.1.1
