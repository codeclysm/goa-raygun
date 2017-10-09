goa-raygun
==========

A [Goa](https://goa.design) middleware to recover panics and send them to [RayGun](https://raygun.com/). [Godoc](https://godoc.org/github.com/codeclysm/goa-raygun)

The minimal setup to have it working is to put the Notify middleware in your middleware chain:

```go
service.Use(middleware.RequestID())
service.Use(middleware.ErrorHandler(service, true))

// this is the middleware
notify := goaraygun.New("MYSECRETRAYGUNKEY", nil)
service.Use(notify.Middleware())

service.Use(middleware.Recover())
```

And that's it. Panics and crashes will be sent to Raygun using the key you specified

# Cleaner errors
The goa `recover` middleware creates an error with all the stacktrace in the error message, while raygun wants it outside.
You can simply use the `goaraygun.Recover` middleware:

```go
service.Use(middleware.RequestID())
service.Use(middleware.ErrorHandler(service, true))
notify := goaraygun.New("MYSECRETRAYGUNKEY", nil)
service.Use(notify.Middleware())

// It creates errors more raygun-friendly, but still understandable by ErrorHandler
service.Use(goaraygun.Recover())
```

# Send errors directly
You can use the goa-raygun manager to send an error directly to raygun. Useful if you don't want to stop execution but you still want to know if something went wrong:

```go
notify.Error(context.Background(), err, request, data)
```

# Debugging
If you don't want to send errors while you are debugging you can use the `Silent Option`. It will print the error in the stdout instead of sending it to the server

```go
notify := goaraygun.New("MYSECRETRAYGUNKEY", &goaraygun.Opts{Silent: true})
service.Use(notify.Middleware())
```

# User info
Every app has its way to retrieve the user info, so if you want that info on raygun you'll have to work a bit for it:

```go
func GetUser(ctx context.Context, req *http.Request) string {
	...
}
notify := goaraygun.New("MYSECRETRAYGUNKEY", &goaraygun.Opts{GetUser: GetUser})
service.Use(notify.Middleware())
```