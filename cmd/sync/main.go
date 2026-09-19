package main

import (
	"context"
	"time"

	"net/url"
	"os"

	contentsync "github.com/desain-gratis/common/delivery/mycontent-api-client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/desain-gratis/deployd-app-storage/src/entity"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Logger()
}

var data = []*entity.UserProfile{
	{Ns: "example.org", Name: "Keenan", Id: "keenan", NickName: "Kobe", Hobby: []string{"computer", "music", "manga"}, PublishedAt: time.Now().Truncate(1 * time.Second)},
}

func main() {
	u, err := url.Parse(os.Getenv("API_ENDPOINT") + "/user-profile")
	if err != nil {
		log.Fatal().Msgf("err: %v", err)
	}

	buildSync := contentsync.Builder[*entity.UserProfile](u).
		WithData(data)

	err = buildSync.Build().Execute(context.Background())
	if err != nil {
		log.Panic().Msgf("failed to execute: %v", err)
	}
}
