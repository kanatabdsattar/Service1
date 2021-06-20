package mid
import (
	"context"
	"net/http"
	"runtime"


)


func Metrics(data *metrics.Metrics) web.Middleware {


	m := func(handler web.Handler) web.Handler {


		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {


			ctx = context.WithValue(ctx, metrics.Key, data)


			err := handler(ctx, w, r)




			data.Requests.Add(1)


			if err != nil {
				data.Errors.Add(1)
			}


			if data.Requests.Value()%100 == 0 {
				data.Goroutines.Set(int64(runtime.NumGoroutine()))
			}

			return err
		}

		return h
	}

	return m
}
