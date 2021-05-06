package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/proto"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	"github.com/infobloxopen/atlas-app-toolkit/health"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
	"github.com/infobloxopen/atlas-app-toolkit/server"
	"github.com/kutty-kumar/ho_oh/ditto_v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	configProviderKey    = "CONFIG_PROVIDER"
	configSvcEndpointKey = "CONFIG_ENDPOINT"
	configPathKey        = "CONFIG_PATH"
)

func main() {
	doneC := make(chan error)
	logger := NewLogger()

	if viper.GetBool("server_config.internal_enable") {
		go func() { doneC <- ServeInternal(logger) }()
	}

	go func() { doneC <- ServeExternal(logger) }()

	if err := <-doneC; err != nil {
		logger.Fatal(err)
	}
}

func NewLogger() *logrus.Logger {
	logger := logrus.StandardLogger()
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Set the log level on the default logger based on command line flag
	logLevels := map[string]logrus.Level{
		"debug":   logrus.DebugLevel,
		"info":    logrus.InfoLevel,
		"warning": logrus.WarnLevel,
		"error":   logrus.ErrorLevel,
		"fatal":   logrus.FatalLevel,
		"panic":   logrus.PanicLevel,
	}
	if level, ok := logLevels[viper.GetString("logging_config.log_level")]; !ok {
		logger.Errorf("Invalid %q provided for log level", viper.GetString("logging_config.log_level"))
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(level)
	}

	return logger
}

// ServeInternal builds and runs the server that listens on InternalAddress
func ServeInternal(logger *logrus.Logger) error {
	healthChecker := health.NewChecksHandler(
		viper.GetString("server_config.internal_health"),
		viper.GetString("server_config.internal_readiness"),
	)
	healthChecker.AddReadiness("DB ready check", dbReady)
	healthChecker.AddLiveness("ping", health.HTTPGetCheck(
		fmt.Sprint("http://", viper.GetString("server_config.internal_address"), ":", viper.GetString("server_config.internal_port"), "/ping"), time.Minute),
	)

	s, err := server.NewServer(
		// register our health checks
		server.WithHealthChecks(healthChecker),
		// this endpoint will be used for our health checks
		server.WithHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("pong"))
		})),
		// register metrics
		server.WithHandler("/metrics", promhttp.Handler()),
	)
	if err != nil {
		return err
	}
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", viper.GetString("server_config.internal_address"), viper.GetString("server_config.internal_port")))
	log.Printf("%s:%s", viper.GetString("server_config.internal_address"), viper.GetString("server_config.internal_port"))
	if err != nil {
		log.Fatalf("%v", err)
		return err
	}

	logger.Debugf("serving internal http at %q", fmt.Sprintf("%s:%s", viper.GetString("server_config.internal_address"), viper.GetString("server_config.internal_port")))
	return s.Serve(nil, l)
}

// ServeExternal builds and runs the server that listens on ServerAddress and GatewayAddress
func ServeExternal(logger *logrus.Logger) error {

	if viper.GetString("database_config.dsn") == "" {
		setDBConnection()
	}
	grpcServer, err := NewGRPCServer(logger)
	if err != nil {
		logger.Fatalln(err)
	}
	grpcPrometheus.Register(grpcServer)

	s, err := server.NewServer(
		server.WithGrpcServer(grpcServer),
		server.WithGateway(
			gateway.WithGatewayOptions(
				runtime.WithForwardResponseOption(forwardResponseOption),
				runtime.WithIncomingHeaderMatcher(gateway.ExtendedDefaultHeaderMatcher(
					requestid.DefaultRequestIDKey)),
			),
			gateway.WithServerAddress(fmt.Sprintf("%s:%s", viper.GetString("server_config.address"), viper.GetString("server_config.port"))),
			gateway.WithEndpointRegistration(viper.GetString("server_config.gateway_url"), ditto_v1.RegisterPrinterServiceHandlerFromEndpoint),
		),
	)
	if err != nil {
		logger.Fatalln(err)
	}

	grpcL, err := net.Listen("tcp", fmt.Sprintf("%s:%s", viper.GetString("server_config.address"), viper.GetString("server_config.port")))
	if err != nil {
		logger.Fatalln(err)
	}

	httpL, err := net.Listen("tcp", fmt.Sprintf("%s:%s", viper.GetString("server_config.gateway_address"), viper.GetString("server_config.gateway_port")))
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Printf("serving gRPC at %s:%s", viper.GetString("server_config.address"), viper.GetString("server_config.port"))
	logger.Printf("serving http at %s:%s", viper.GetString("server_config.gateway_address"), viper.GetString("server_config.gateway_port"))

	return s.Serve(grpcL, httpL)
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := viper.AddRemoteProvider(viper.GetString(configProviderKey), viper.GetString(configSvcEndpointKey), viper.GetString(configPathKey))
	if err != nil {
		log.Fatalf("An error %v occurred while fetching config form %v", err, viper.GetString(configProviderKey))
	}
	viper.SetConfigType("json") // Need to explicitly set this to json
	err = viper.ReadRemoteConfig()
	if err != nil {
		log.Fatalf("An error %v occurred while reading config", err)
	}
	resource.RegisterApplication(viper.GetString("app.id"))
	resource.SetPlural()
}

func forwardResponseOption(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	return nil
}

func dbReady() error {
	if viper.GetString("database_config.dsn") == "" {
		setDBConnection()
	}
	db, err := sql.Open(viper.GetString("database_config.type"), viper.GetString("database_config.dsn"))
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}

// setDBConnection sets the db connection string
func setDBConnection() {
	viper.Set("database_config.dsn", fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s dbname=%s",
		viper.GetString("database_config.host_name"), viper.GetString("database_config.port"),
		viper.GetString("database_config.user_name"), viper.GetString("database_config.password"),
		viper.GetString("database_config.ssl"), viper.GetString("database_config.name")))
}