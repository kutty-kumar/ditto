package main

import (
	"context"
	"ditto/pkg/domain"
	"ditto/pkg/repository"
	"ditto/pkg/svc"
	"github.com/dgrijalva/jwt-go"
	"log"
	"os"
	"time"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcLogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
	"github.com/kutty-kumar/charminder/pkg"
	"github.com/kutty-kumar/ho_oh/ditto_v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gLogger "gorm.io/gorm/logger"
)

var (
	reg                     = prometheus.NewRegistry()
	grpcMetrics             = grpcPrometheus.NewServerMetrics()
	createUserSuccessMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "user_service_create_user_success_count",
		Help: "total number of successful invocations of create user method in user service",
	}, []string{"create_user_success_count"})
	createUserFailureMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "user_service_create_user_failure_count",
		Help: "total number of failure invocations of create user method in user service",
	}, []string{"create_user_failure_count"})
)

type Claims struct {
	UserName string `json:"user_name"`
	UserId   string `json:"user_id"`
	jwt.StandardClaims
}

func init() {
	reg.MustRegister(grpcMetrics, createUserSuccessMetric, createUserFailureMetric)
	createUserSuccessMetric.WithLabelValues("user_service")
	createUserFailureMetric.WithLabelValues("user_service")
}

func NewGRPCServer(logger *logrus.Logger) (*grpc.Server, error) {
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				Time:    time.Duration(viper.GetInt("heart_beat_config.keep_alive_time")) * time.Second,
				Timeout: time.Duration(viper.GetInt("heart_beat_config.keep_alive_timeout")) * time.Second,
			},
		),
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryServer(
				// logging middleware
				grpcLogrus.UnaryServerInterceptor(logrus.NewEntry(logger)),

				// Request-Id interceptor
				requestid.UnaryServerInterceptor(),

				// Metrics middleware
				grpcPrometheus.UnaryServerInterceptor,

				// validation middleware
				grpcValidator.UnaryServerInterceptor(),

				// collection operators middleware
				gateway.UnaryServerInterceptor(),

				AuthUnaryServerInterceptor(),
			),
		),
	)

	dbLogger := gLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		gLogger.Config{
			SlowThreshold:             time.Second,  // Slow SQL threshold
			LogLevel:                  gLogger.Info, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,        // Disable color
		},
	)
	// create new mysql database connection
	db, err := gorm.Open(mysql.Open(viper.GetString("database_config.dsn")), &gorm.Config{Logger: dbLogger})
	if err != nil {
		return nil, err
	}

	//dropTables(db)
	//createTables(db)
	baseDao := pkg.NewBaseGORMDao(pkg.WithDb(db),
		pkg.WithLogger(logger),
		pkg.WithCreator(func() pkg.Base {
			return &domain.Printer{}
		}),
		pkg.WithExternalIdSetter(func(externalId string, base pkg.Base) pkg.Base {
			base.SetExternalId(externalId)
			return base
		}))
	printerDao := repository.NewPrinterGORMRepository(baseDao)
	baseSvc := pkg.NewBaseSvc(baseDao)
	printerSvc := svc.NewPrinterSvc(&baseSvc, printerDao)
	ditto_v1.RegisterPrinterServiceServer(grpcServer, printerSvc)
	grpcMetrics.InitializeMetrics(grpcServer)
	return grpcServer, nil
}

func createTables(db *gorm.DB) {
	err := db.AutoMigrate(domain.Printer{})
	if err != nil {
		log.Fatalf("An error %v occurred while automigrating", err)
	}
}

func AuthUnaryServerInterceptor() grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		headers, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "headers absent")
		}
		if headers.Len() != 0 && headers.Get("Authorization") != nil {
			claims := &Claims{}
			tkn, err := jwt.ParseWithClaims(headers.Get("Authorization")[0], claims, func(token *jwt.Token) (i interface{}, e error) {
				return viper.GetString("server_config.jwt_key"), nil
			})
			if err != nil {
				return nil, status.Errorf(codes.PermissionDenied, "invalid signature")
			}
			if !tkn.Valid {
				return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
			}
			metadata.AppendToOutgoingContext(ctx, "user_id", claims.UserId)
			return handler(ctx, req)
		}

		if headers.Len() != 0 && headers.Get("Authorization") == nil {
			return nil, status.Errorf(codes.Unauthenticated, "auth failure")
		}

		return handler(ctx, req)
	}
}
