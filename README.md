goa-raygun
==========

A goa middleware to recover panics and send them to RayGun

The minimal setup to have it working is to put the Notify middleware in your middleware chain:

```go
service.Use(middleware.RequestID())
service.Use(middleware.ErrorHandler(service, true))

// this is the middleware
service.Use(goaraygun.Notify("MYSECRETRAYGUNKEY", nil))

service.Use(middleware.Recover())
```

And that's it. Panics and crashes will be sent to Raygun using the key you specified.

# Cleaner errors
The goa `recover` middleware creates an error with all the stacktrace in the error message, while raygun wants it outside.
You can simply use the `goaraygun.Recover` middleware:

```go
service.Use(middleware.RequestID())
service.Use(middleware.ErrorHandler(service, true))
service.Use(goaraygun.Notify("MYSECRETRAYGUNKEY", nil))

// It creates errors more raygun-friendly, but still understandable by ErrorHandler
service.Use(goaraygun.Recover())
```

# Debugging
If you don't want to send errors while you are debugging you can use the `Silent Option`. It will print the stacktrace in the stdout instead of sending it to the server

```
service.Use(goaraygun.Notify("MYSECRETRAYGUNKEY", &goaraygun.Opts{Silent: true}))
```

# User info
Every app has its way to retrieve the user info, so if you want that info on raygun you'll have to work a bit for it:

```
func GetUser(ctx context.Context, req *http.Request) string {
	...
}
service.Use(goaraygun.Notify("MYSECRETRAYGUNKEY", &goaraygun.Opts{GetUser: GetUser}))
```