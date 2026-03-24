package app

import (
	"context"
	"log"

	notifypb "github.com/tasker-iniutin/api-contracts/gen/go/proto/notify/v1alpha"
	authsec "github.com/tasker-iniutin/common/authsecurity"
	"github.com/tasker-iniutin/common/grpcauth"
	"github.com/tasker-iniutin/common/postgres"
	"github.com/tasker-iniutin/common/runtime"
	"google.golang.org/grpc"

	"github.com/tasker-iniutin/notify-service/internal/store/postgre"
	handler "github.com/tasker-iniutin/notify-service/internal/transport/grpc"
)

type App struct {
	cfg Config
}

func New(cfg Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Run(ctx context.Context) error {
	db, err := postgres.Open(context.Background(), a.cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	repo := postgre.NewNotifyPostgreRepo(db)
	h := handler.NewServer(repo)

	pub, err := authsec.LoadRSAPublicKeyFromPEMFile(a.cfg.PublicKeyPath)
	if err != nil {
		return err
	}

	verifier := authsec.NewRS256Verifier(pub, a.cfg.JWTIssuer, a.cfg.JWTAudience)

	whitelist := map[string]struct{}{}

	log.Printf("notify-service gRPC listening on %s", a.cfg.GRPCAddr)
	log.Printf("notify-service auth config: public_key=%s issuer=%s audience=%s",
		a.cfg.PublicKeyPath, a.cfg.JWTIssuer, a.cfg.JWTAudience,
	)
	log.Printf("notify-service database config: dsn=%s", a.cfg.DatabaseURL)

	return runtime.ServeGRPCWithContext(
		ctx,
		a.cfg.GRPCAddr,
		func(server *grpc.Server) {
			notifypb.RegisterNotifyServiceServer(server, h)
		},
		a.cfg.EnableReflection,
		grpc.UnaryInterceptor(grpcauth.UnaryAuthInterceptor(verifier, whitelist)),
	)
}
