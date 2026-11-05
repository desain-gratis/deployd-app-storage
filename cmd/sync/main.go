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

var user = []*entity.UserProfile{
	{Ns: "example.org", Name: "Keenan", Id: "keenan", NickName: "Kobe", Hobby: []string{"computer", "music", "manga"}, PublishedAt: time.Now().Truncate(1 * time.Second)},
}

var product = []*entity.Product{
	{Ns: "example.org", Id: "tv", Name: "Television", Description: "A Television to watch Forum Keadilan", Price: "4000000", Currency: "IDR"},
	{Ns: "example.org", Id: "fridge", Name: "Kulkas", Description: "Cooling food", Price: "12000000", Currency: "IDR"},
}

func main() {
	uprofile, err := url.Parse(os.Getenv("API_ENDPOINT") + "/user-profile")
	if err != nil {
		log.Fatal().Msgf("err: %v", err)
	}

	uproduct, err := url.Parse(os.Getenv("API_ENDPOINT") + "/product")
	if err != nil {
		log.Fatal().Msgf("err: %v", err)
	}

	userSync := contentsync.Builder[*entity.UserProfile](uprofile).
		WithData(user)

	productSync := contentsync.Builder[*entity.Product](uproduct).
		WithData(product)

	err = userSync.Build().Execute(context.Background())
	if err != nil {
		log.Panic().Msgf("failed to execute: %v", err)
	}

	err = productSync.Build().Execute(context.Background())
	if err != nil {
		log.Panic().Msgf("failed to execute: %v", err)
	}
}
