package goaraygun

import (
	"context"
	"errors"
	"net/http"

	"github.com/goadesign/goa"
	"github.com/gsblue/raygun4go"
)

// Opts contain the configuration of the middleware
type Opts struct {
	Name    string
	Key     string
	Silent  bool
	GetUser func(ctx context.Context) string
}

// Recover is a middleware that recover panics and notifies Raygun. It's intended to replace middleware.Recover
// Note that it doesn't print the full stack trace anymore (since it's sent to raygun)
func Recover(opts Opts) goa.Middleware {
	raygun, err := raygun4go.New(opts.Name, opts.Key)
	if err != nil {
		panic("Unable to create Raygun client:" + err.Error())
	}

	raygun.Silent(opts.Silent)

	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
			defer func() {
				if e := recover(); e != nil {
					var ok bool
					err, ok = e.(error)
					if !ok {
						err = errors.New(e.(string))
					}

					rayErr := raygun.CreateErrorEntry(err)
					rayErr.SetRequest(req)

					// Attempt to retrieve a user
					if opts.GetUser != nil {
						rayErr.SetUser(opts.GetUser(ctx))
					}

					if subErr := raygun.SubmitError(rayErr); subErr != nil {
						panic(subErr.Error())
					}
				}
			}()

			return h(ctx, rw, req)
		}
	}
}
