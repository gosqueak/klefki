package main

import (
	"crypto/rsa"
	"os"

	kit "github.com/gosqueak/apikit"
	"github.com/gosqueak/jwt"
	"github.com/gosqueak/jwt/rs256"
	"github.com/gosqueak/klefki/api"
	"github.com/gosqueak/klefki/database"
	"github.com/gosqueak/leader/team"
)

func main() {
	tm, err := team.Download(os.Getenv("TEAMFILE_URL"))

	if err != nil {
		panic(err)
	}

	klefki := tm.Member("klefki")
	steelix := tm.Member("steelix")

	db := database.Load("data.sqlite")

	pKey, err := kit.Retry[*rsa.PublicKey](3, rs256.FetchRsaPublicKey, steelix.Url.String()+"/jwtkeypub")
	if err != nil {
		panic("could not fetch RSA key")
	}

	aud := jwt.NewAudience(pKey, klefki.JWTInfo.AudienceName)
	serv := api.NewServer(klefki.ListenAddress, db, aud)

	serv.Run()
}
