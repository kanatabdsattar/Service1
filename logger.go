package mid

import (
	"context"
	"net/http"
	"time"
)

package mid

import (
"context"
"net/http"
"time"


)


func Logger(log *zap.SugaredLogger) web.Middleware {

	m := func(handler web.Handler) web.Handler {


		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {


			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			log.Infow("request started", "traceid", v.TraceID, "method", r.Method, "path", r.URL.Path,
				"remoteaddr", r.RemoteAddr)


			err := handler(ctx, w, r)

			log.Infow("request completed", "traceid", v.TraceID, "method", r.Method, "path", r.URL.Path,
				"remoteaddr", r.RemoteAddr, "statuscode", v.StatusCode, "since", time.Since(v.Now))


			return err
		}

		return h
	}

	return m
}