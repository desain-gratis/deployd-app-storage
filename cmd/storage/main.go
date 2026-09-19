package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	mycontentapi "github.com/desain-gratis/common/delivery/mycontent-api"
	mycontent_base "github.com/desain-gratis/common/delivery/mycontent-api/mycontent/base"
	content_badgerraft "github.com/desain-gratis/common/delivery/mycontent-api/storage/content/badger-raft"
	runneretcd "github.com/desain-gratis/common/lib/raft/runner-etcd"
	"github.com/desain-gratis/deployd-app-storage/src/entity"
)

const (
	publicBaseURL = "https://deployd-app-storage.desain.gratis"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Logger()
}

const (
	httpPublicAddress = ":9090"
)

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())

	router := httprouter.New()

	enableStorageModule(ctx, router)

	wg := new(sync.WaitGroup)

	go startHttpListener(ctx, wg, router, httpPublicAddress)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	log.Info().Msgf("waiting for sigint")
	<-sigint
	cancel(errors.New("server closed"))
	wg.Wait()
	log.Info().Msgf("bye bye")
}

func enableStorageModule(appCtx context.Context, router *httprouter.Router) {
	db, err := badger.Open(badger.DefaultOptions("./tmp/storage.db"))
	if err != nil {
		log.Fatal().Msgf("UHUY %v", err)
	}

	// A raft application that provides distributed mycontent storage
	// TODO: have better API
	badgerStorageApp := content_badgerraft.New(
		db,
		content_badgerraft.TableConfig{Name: "user_profile", RefSize: 0},
	)

	raftCtx, _, err := runneretcd.RunWithConfig(appCtx, os.Getenv("DEPLOYD_RAFT"), "storage", badgerStorageApp)
	if err != nil {
		log.Fatal().Msgf("err init raft: %v", err)
	}

	userProfileRepo, err := badgerStorageApp.GetKVTable(raftCtx, "user_profile")
	if err != nil {
		log.Fatal().Msgf("%v", err)
	}
	userProfileUsecase := mycontent_base.New[*entity.UserProfile](userProfileRepo)

	userProfileHandler := mycontentapi.New(
		userProfileUsecase,
		publicBaseURL+"/user-profile",
		nil,
	)

	router.GET("/user-profile", userProfileHandler.Get)
	router.POST("/user-profile", userProfileHandler.Post)
	router.DELETE("/user-profile", userProfileHandler.Delete)
}
