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

var version = "v1.1.0"

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

// Manager exposes a Notify middleware and and an Error function. Use the Notify middleware in the goa middleware chain, and the Error method to send an error to
type Manager struct {
	Opts Opts
	Key  string
}

// New returns a new Manager
func New(key string, opts *Opts) *Manager {
	if opts == nil {
		opts = &Opts{Silent: false}
	}
	return &Manager{*opts, key}
}

// Error sends a custom error to raygun. You can provide a request, tags and custom data if you want
func (m *Manager) Error(ctx context.Context, err error, req *http.Request, tags []string, data interface{}) {
	post := raygun.NewPost()
	post.Details.Version = m.Opts.Version
	post.Details.Error = raygun.FromErr(err)
	if req != nil {
		post.Details.Request = raygun.FromReq(req)
		if m.Opts.GetUser != nil {
			post.Details.User = raygun.User{Identifier: m.Opts.GetUser(ctx, req)}
		}
	}
	post.Details.UserCustomData = data
	post.Details.Tags = tags
	post.Details.Client = raygun.Client{
		Name:      "goa-raygun",
		Version:   version,
		ClientURL: "https://github.com/codeclysm/goa-raygun",
	}

	if !m.Opts.Silent {
		if e := raygun.Submit(post, m.Key, nil); e != nil {
			panic(e)
		}
	} else {
		fmt.Printf("%+v\n", err)
	}
}

// Middleware is a middleware that sends critical errors to Raygun. It should sit between ErrorHandler and Recover in the middleware chain. Key is the raygun api key, opts can be nil or can be an Opts struct
func (m *Manager) Middleware() goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			err := h(ctx, rw, req)
			if err != nil {
				if !skip(ctx, m.Opts, err) {
					m.Error(ctx, err, req, goa.ContextRequest(ctx).Payload)
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
