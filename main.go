package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	redisStore "github.com/go-oauth2/redis/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"

	"github.com/byebyebymyai/oauth2-api/ent"
	"github.com/byebyebymyai/oauth2-api/ent/migrate"
)

var jose string
var rbac string

var dsn string

var redisOptions *redis.Options

var redisClient *redis.Client

var logger *slog.Logger
var errorLogger *slog.Logger

// oauth2
var srv *server.Server

func main() {
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	errorLogHandler := slog.NewJSONHandler(os.Stdout, nil)

	logger = slog.New(logHandler)
	errorLogger = slog.New(errorLogHandler)

	client, err := ent.Open("mysql", dsn, ent.Debug(), ent.Log(func(a ...any) {
		logger.Debug("[ent]", "msg", a)
	}))
	if err != nil {
		panic(err.Error())
	}

	defer client.Close()
	if err := client.Schema.Create(context.Background(), migrate.WithGlobalUniqueID(true)); err != nil {
		logger.Error("[main]", "msg", "failed creating schema resources", "err", err)
	}

	if os.Getenv("REDIS_ENABLED") == "true" {
		redisClient = redis.NewClient(redisOptions)
		_, err = redisClient.Ping(context.Background()).Result()
		if err != nil {
			panic(err)
		}
	}

	initOAuth2(context.Background(), client)

	mux := http.NewServeMux()

	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			errorLogger.Error("[authorizeHandle]", "error", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		err := srv.HandleTokenRequest(w, r)
		if err != nil {
			errorLogger.Error("[tokenHandle]", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	logger.Info("[main]", "message", "starting server", "port", 8080)
	http.ListenAndServe(":8080", mux)

}

func init() {

	jose = os.Getenv("JOSE_URL")
	rbac = os.Getenv("RBAC_URL")

	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	hostname := os.Getenv("DB_HOST")
	database := os.Getenv("DB_NAME")

	dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", username, password, hostname, database)

	redis_address := os.Getenv("REDIS_ADDRESS")
	redis_password := os.Getenv("REDIS_PASSWORD")

	redisOptions = &redis.Options{
		Addr:     redis_address,
		Password: redis_password,
	}

}

func initOAuth2(ctx context.Context, client *ent.Client) {
	// token store
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.MapAccessGenerate(makeProxyTokenService(ctx, jose)(&defaultTokenService{}))

	if os.Getenv("REDIS_ENABLED") == "true" {
		// token redis store
		manager.MapTokenStorage(redisStore.NewRedisStore(redisOptions))
	} else {
		// token memory store
		manager.MustTokenStorage(store.NewMemoryTokenStore())
	}

	// init client store
	manager.MapClientStorage(&ClientStorage{
		ctx:    ctx,
		client: client,
	})

	// oauth2 server setting
	srv = server.NewServer(server.NewConfig(), manager)
	// check the client allows to use this authorization grant type
	srv.SetAllowGetAccessRequest(true)
	// get client info from request
	srv.SetClientInfoHandler(server.ClientFormHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		errorLogger.Error("[internalError]", "error", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		errorLogger.Error("[responseError]", "error", re.Error.Error(), "errorCode", re.ErrorCode, "description", re.Description, "uri", re.URI, "statusCode", re.StatusCode, "header", re.Header)
	})

	srv.SetPasswordAuthorizationHandler(func(ctx context.Context, clientID, username, password string) (userID string, err error) {
		users, err := makeProxyUserService(ctx, rbac)(&defaultUserService{}).All(User{
			Username: username,
		})
		if err != nil {
			return "", err
		}
		if len(users) == 0 {
			return "", errors.New("User not found")
		}
		err = bcrypt.CompareHashAndPassword(users[0].Password, []byte(password))
		if err != nil {
			return "", errors.New("Invalid password")
		}
		return users[0].ID.String(), nil
	})

	// get user id from request authorization
	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			return "", errors.New("Authorization header is missing")
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return "", errors.New("Invalid authorization header format")
		}

		token, _, err := new(jwt.Parser).ParseUnverified(tokenParts[1], jwt.MapClaims{})
		if err != nil {
			return "", errors.New("Invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return "", errors.New("Invalid token claims")
		}

		userID, ok = claims["sub"].(string)
		if !ok {
			return "", errors.New("Sub claim is missing in token")
		}
		if userID == "" {
			return "", errors.New("UserID is empty")
		}
		return userID, nil
	})
}
