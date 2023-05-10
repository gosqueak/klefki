package main

import (
	"crypto/rsa"

	kit "github.com/gosqueak/apikit"
	"github.com/gosqueak/jwt"
	"github.com/gosqueak/jwt/rs256"
	"github.com/gosqueak/klefki/api"
	"github.com/gosqueak/klefki/database"
)

const (
	Addr            = "0.0.0.0:8083"
	AuthServerUrl   = "http://0.0.0.0:8081"
	JwtKeyPublicUrl = AuthServerUrl + "/jwtkeypub"
	AudIdentifier   = "ECDHSERVICE"
)

func main() {
	db := database.Load("data.sqlite")

	pKey, err := kit.Retry[*rsa.PublicKey](3, rs256.FetchRsaPublicKey, []any{JwtKeyPublicUrl})
	if err != nil {
		panic("could not fetch RSA key")
	}

	aud := jwt.NewAudience(pKey, AudIdentifier)
	serv := api.NewServer(Addr, db, aud)

	serv.Run()
}
