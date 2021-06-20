package handlers
import (
	"context"
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"


)


type Options struct {
	corsOrigin string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origin string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origin
	}
}


func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()


	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}


func DebugMux(build string, log *zap.SugaredLogger, db *sqlx.DB) http.Handler {
	mux := DebugStandardLibraryMux()

	// Register debug check endpoints.
	cg := checkGroup{
		build: build,
		log:   log,
		db:    db,
	}
	mux.HandleFunc("/debug/readiness", cg.readiness)
	mux.HandleFunc("/debug/liveness", cg.liveness)

	return mux
}

// APIMux constructs an http.Handler with all application routes defined.
func APIMux(build string, shutdown chan os.Signal, log *zap.SugaredLogger, metrics *metrics.Metrics, a *auth.Auth, db *sqlx.DB, options ...func(opts *Options)) http.Handler {
	var opts Options
	for _, option := range options {
		option(&opts)
	}


	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(metrics), mid.Panics())


	ug := userGroup{
		store: user.NewStore(log, db),
		auth:  a,
	}
	app.Handle(http.MethodGet, "/v1/users/:page/:rows", ug.query, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/users/token/:kid", ug.token)
	app.Handle(http.MethodGet, "/v1/users/:id", ug.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/users", ug.create, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodPut, "/v1/users/:id", ug.update, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodDelete, "/v1/users/:id", ug.delete, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))


	pg := productGroup{
		store: product.NewStore(log, db),
	}
	app.Handle(http.MethodGet, "/v1/products/:page/:rows", pg.query, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/products/:id", pg.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/products", pg.create, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/products/:id", pg.update, mid.Authenticate(a))
	app.Handle(http.MethodDelete, "/v1/products/:id", pg.delete, mid.Authenticate(a))


	if opts.corsOrigin != "" {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return nil
		}
		app.Handle(http.MethodOptions, "/*", h)
	}

	return app
}