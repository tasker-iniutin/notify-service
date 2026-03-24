package app

import "github.com/tasker-iniutin/common/configenv"

type Config struct {
	GRPCAddr         string
	PublicKeyPath    string
	JWTIssuer        string
	JWTAudience      string
	EnableReflection bool
	DatabaseURL      string
}

func LoadConfig() Config {
	return Config{
		GRPCAddr:         configenv.String("NOTIFY_GRPC_ADDR", ":50053"),
		PublicKeyPath:    configenv.String("JWT_PUBLIC_KEY_PEM", "../auth-service/keys/public.pem"),
		JWTIssuer:        configenv.String("JWT_ISSUER", "todo-auth"),
		JWTAudience:      configenv.String("JWT_AUDIENCE", "todo-api"),
		EnableReflection: configenv.Bool("ENABLE_GRPC_REFLECTION", false),
		DatabaseURL:      configenv.String("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"),
	}
}
