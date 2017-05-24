// Package goaraygun provides a couple of middlewares for https://goa.design/ to send failures to https://raygun.com/
package goaraygun

import (
	"context"
	"fmt"
	"net/http"

	"errors"

	"github.com/codeclysm/raygun"
	"github.com/goadesign/goa"
)

var version = "v1.0.0"

// Recover is a middleware that recovers panics and maps them to errors. Use this instead of the goa one to have cleaner errors without the stacktrace in their main message.
func Recover() goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
			defer func() {
				if e := recover(); e != nil {
					var ok bool
					err, ok = e.(error)
					if !ok {
						err = errors.New(e.(string))
					}

					rayErr := raygun.FromErr(err)
					rayErr.StackTrace = rayErr.StackTrace[2:]
					err = rayErr

				}
			}()
			return h(ctx, rw, req)
		}
	}
}

// Opts contain the configuration of the Notify middleware
type Opts struct {
	// Version is an optional value to provide the version of the app you are monitoring
	Version string
	// Silent is an optional value to avoid sending errors but just to print them in the stdout. Useful for debugging
	Silent bool
	// Skip is an optional function to decide if the error is worth sending to raygun or not.
	// If it's not defined, only panics and status codes of 500 are sent.
	Skip func(ctx context.Context, err error) bool
	// GetUser is an optional function to retrieve the username from the context and/or the request
	GetUser func(ctx context.Context, req *http.Request) string
}

// Notify is a middleware that sends critical errors to Raygun. It should sit between ErrorHandler and Recover in the middleware chain. Key is the raygun api key, opts can be nil or can be an Opts struct
func Notify(key string, opts *Opts) goa.Middleware {
	if opts == nil {
		opts = &Opts{Silent: false}
	}

	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			err := h(ctx, rw, req)
			if err != nil {
				if !skip(ctx, *opts, err) {
					post := raygun.NewPost()
					post.Details.Version = opts.Version
					post.Details.Error = raygun.FromErr(err)
					post.Details.Request = raygun.FromReq(req)
					post.Details.UserCustomData = goa.ContextRequest(ctx).Payload
					post.Details.Client = raygun.Client{
						Name:      "goa-raygun",
						Version:   version,
						ClientURL: "https://github.com/codeclysm/goa-raygun",
					}

					if opts.GetUser != nil {
						post.Details.User = raygun.User{Identifier: opts.GetUser(ctx, req)}
					}

					if !opts.Silent {
						if e := raygun.Submit(post, key, nil); e != nil {
							panic(e)
						}
					} else {
						fmt.Println(post.Details.Error.StackTrace.String())
					}
				}
			}

			return err
		}
	}
}

func skip(ctx context.Context, opts Opts, err error) bool {
	if opts.Skip != nil {
		return opts.Skip(ctx, err)
	}

	if err, ok := err.(goa.ServiceError); ok {
		return err.ResponseStatus() != 500
	}

	return false
}
