package articleApp

import (
	"context"
	"log/slog"
	httparticle "sso/internal/http/acticle"
	articleservice "sso/internal/services/article"
	articleauth "sso/internal/services/article/auth"
	"sso/internal/storage/postgres"
	redishandler "sso/internal/storage/redis"
	"time"

	ssov1 "github.com/los3-smurf/photos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ArticleApp struct {
	log         *slog.Logger
	Srv         *httparticle.ArticleServer
	SsoClient   ssov1.AuthClient
	gRPCConn    *grpc.ClientConn
	RedisClient *redishandler.CacheStorage
	Storage     *postgres.Storage
	AppId       int64
}

func NewArticleApp(log *slog.Logger,
	name string,
	secret string,
	storagePath string,
	redisAddr string,
	redisPassword string,
	redisDBNum int,
	redisTTL time.Duration,
	httpAddr string,
	gRPCAddr string,
	timeout time.Duration,

) *ArticleApp {

	storage, err := postgres.NewStorage(storagePath)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	appId, err := storage.Init(ctx, name, secret)
	if err != nil {
		panic(err)
	}

	conn, err := grpc.NewClient(gRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	redisClient, err := redishandler.NewCacheStorage(redisAddr, redisPassword, redisDBNum, redisTTL)
	if err != nil {
		panic(err)
	}

	ssoClient := ssov1.NewAuthClient(conn)

	authArticle := articleauth.NewArticleAuth(ssoClient, int32(appId))
	handler := articleservice.NewArticleHandler(name, appId, log, storage, storage, storage, redisClient, authArticle)
	srv := httparticle.NewArticleServer(log, httpAddr, timeout, handler, secret)

	return &ArticleApp{
		log:         log,
		Srv:         srv,
		SsoClient:   ssoClient,
		gRPCConn:    conn,
		RedisClient: redisClient,
		Storage:     storage,
		AppId:       appId,
	}
}

func (a *ArticleApp) Stop(ctx context.Context) {
	const op = "app.article.stop"

	a.log.With(slog.String("op", op)).Info("stopping notification app")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.Srv.Stop(shutdownCtx); err != nil {
		a.log.Error("failed to stop http server", slog.Any("err", err))
	}

	if a.gRPCConn != nil {
		if err := a.gRPCConn.Close(); err != nil {
			a.log.Error("failed to close grpc connection", slog.Any("err", err))
		}
	}

	a.log.Info("stopping cache")
	a.RedisClient.Close()

	a.log.Info("stopping storage")
	a.Storage.Stop()

}
