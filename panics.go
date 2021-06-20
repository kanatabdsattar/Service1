package mid

import (
	"context"
	"errors"
	"net/http"

)


func Panics() web.Middleware {


	m := func(handler web.Handler) web.Handler {


		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {


			defer func() {
				if rec := recover(); rec != nil {
					err = errors.Errorf("PANIC: %v", rec)
					metrics.AddPanics(ctx)
				}
			}()


			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
