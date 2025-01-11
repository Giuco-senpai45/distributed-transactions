package app

import (
	"context"
	"dt/controllers"
	"dt/db"
	routes "dt/http"
	"dt/middleware"
	"dt/services"
	"dt/utils"
	"dt/utils/log"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"
)

type App struct {
	router     *http.ServeMux
	migrations []fs.FS
}

func New(migrations []fs.FS) *App {
	router := http.NewServeMux()

	app := &App{
		router:     router,
		migrations: migrations,
	}

	return app
}

func (a *App) Start(ctx context.Context) error {
	mvccDbName := utils.GetEnvOrDefault("MVCC_DB_NAME", "mvcc-db")
	appDbName := utils.GetEnvOrDefault("APP_DB_NAME", "app-db")

	mvccDbConfig := db.LoadConfigFromEnv(a.migrations[0], "db/migrations/mvcc")
	appDbConfig := db.LoadConfigFromEnv(a.migrations[1], "db/migrations/app")
	mvccDbConfig.DBName = mvccDbName
	appDbConfig.DBName = appDbName

	mvccDbAdapter, err := db.NewAdapter(mvccDbConfig)
	if err != nil {
		log.Error("Failed to connect to mvcc database ", err)
	}
	defer db.CloseConnection(mvccDbAdapter)

	appDbAdapter, err := db.NewAdapter(appDbConfig)
	if err != nil {
		log.Error("Failed to connect to app database ", err)
	}
	defer db.CloseConnection(appDbAdapter)

	// service, controllers
	ms := services.NewMVCCService(mvccDbAdapter, appDbAdapter)
	us := services.NewUserService(ms)
	acs := services.NewAccountService(ms)
	as := services.NewAuditService(ms)

	uc := controllers.NewUserController(us)
	acc := controllers.NewAccountController(acs)
	ac := controllers.NewAuditController(as)

	router := http.NewServeMux()

	routes.RegisterRoutes(router, uc, acc, ac)
	routerHandler := middleware.CorsMiddleware(middleware.LoggingMiddleware(router))

	appPort := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	if appPort == "" {
		appPort = ":8080"
	}

	server := http.Server{
		Addr:    appPort,
		Handler: routerHandler,
	}

	done := make(chan struct{})
	go func() {
		log.Info("Server started on port %s", appPort)
		err := server.ListenAndServe()
		if err != nil {
			log.Error("Server stopped", err)
		}
		close(done)
	}()

	select {
	case <-done:
		break
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		server.Shutdown(ctx)
		cancel()
	}

	return nil
}
