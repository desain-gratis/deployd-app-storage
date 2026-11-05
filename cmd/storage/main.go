package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"os/signal"
	"sync"

	_ "github.com/chdb-io/chdb-go/v2/chdb/driver"
	"github.com/dgraph-io/badger/v4"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	mycontentapi "github.com/desain-gratis/common/delivery/mycontent-api"
	mycontent_base "github.com/desain-gratis/common/delivery/mycontent-api/mycontent/base"
	content_badgerraft "github.com/desain-gratis/common/delivery/mycontent-api/storage/content/badger-raft"
	chdbraft "github.com/desain-gratis/common/delivery/mycontent-api/storage/content/raftstorage/chdb-raft"
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

	enableBadgerStorageModule(ctx, router)
	enableChDBStorageModule(ctx, router)

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

func enableBadgerStorageModule(appCtx context.Context, router *httprouter.Router) {
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

	raftInst, _, err := runneretcd.RunWithConfig(appCtx, os.Getenv("DEPLOYD_RAFT"), "storage", badgerStorageApp)
	if err != nil {
		log.Fatal().Msgf("err init raft: %v", err)
	}

	raftCtx := runneretcd.WithRaftContext(appCtx, raftInst)

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

	// start serving mutation to the underlying storage
	err = raftInst.Run(appCtx, nil)
	if err != nil {
		log.Fatal().Msgf("err raft run: %v", err)
	}
}

func enableChDBStorageModule(appCtx context.Context, router *httprouter.Router) {
	chDB, err := sql.Open("chdb", "session=./tmp/storage.chdb")
	if err != nil {
		log.Fatal().Msgf("UHUY %v", err)
	}

	// init the state machine
	app, err := chdbraft.New(appCtx, chDB,
		chdbraft.TableConfig{Name: "product", RefSize: 0, TableType: chdbraft.TableTypeKV},
		chdbraft.TableConfig{Name: "product_review", RefSize: 1, TableType: chdbraft.TableTypeAutoIncrement},
	)
	if err != nil {
		log.Fatal().Msgf("UHUY %v", err)
	}

	// connect & run the raft app to the raft state machine
	raftInst, _, err := runneretcd.RunWithConfig(appCtx, os.Getenv("DEPLOYD_RAFT"), "storage-chdb", app)
	if err != nil {
		log.Fatal().Msgf("err init raft: %v", err)
	}

	productRepo, ok := app.GetRepository(raftInst, "product")
	if !ok {
		log.Fatal().Msgf("err get repository: %v", err)
	}
	productReviewRepo, ok := app.GetRepository(raftInst, "product_review")
	if !ok {
		log.Fatal().Msgf("err get repository: %v", err)
	}

	productUsecase := mycontent_base.New[*entity.Product](productRepo)
	productReviewUsecase := mycontent_base.New[*entity.ProductReview](productReviewRepo)

	productHandler := mycontentapi.New(
		productUsecase,
		publicBaseURL+"/product",
		nil,
	)
	productReviewHandler := mycontentapi.New(
		productReviewUsecase,
		publicBaseURL+"/product/review",
		[]string{"product_id"},
	)

	router.GET("/product", productHandler.Get)
	router.POST("/product", productHandler.Post)
	router.DELETE("/product", productHandler.Delete)

	router.GET("/product/review", productReviewHandler.Get)
	router.POST("/product/review", productReviewHandler.Post)
	router.DELETE("/product/review", productReviewHandler.Delete)

	// start serving mutation to the underlying storage
	err = raftInst.Run(appCtx, nil)
	if err != nil {
		log.Fatal().Msgf("err raft run: %v", err)
	}
}
