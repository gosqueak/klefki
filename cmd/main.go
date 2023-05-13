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
	tm := team.Download(os.Getenv("TEAMFILE_JSON_URL"))
	klefki := tm["klefki"]
	steelix := tm["steelix"]

	db := database.Load("data.sqlite")

	pKey, err := kit.Retry[*rsa.PublicKey](3, rs256.FetchRsaPublicKey, steelix.Url + "/jwtkeypub")
	if err != nil {
		panic("could not fetch RSA key")
	}

	aud := jwt.NewAudience(pKey, klefki.JWTInfo.AudienceName)
	serv := api.NewServer(klefki.Url, db, aud)

	serv.Run()
}
