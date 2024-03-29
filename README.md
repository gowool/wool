# Wool Web Framework

![License](https://img.shields.io/dub/l/vibe-d.svg)

Wool is a web framework written in Go (Golang).

## Installation

To install Wool package, you need to install Go and set your Go workspace first.

1. You first need [Go](https://go.dev/) installed, then you can use the below Go command to install Wool.

```sh
go get github.com/gowool/wool
```

2. Import it in your code:

```go
import "github.com/gowool/wool"
```

### Running Wool

```go
package main

import (
    "context"
    "net/http"
	"os"
    
    "github.com/gowool/middleware/proxy"
    "github.com/gowool/wool"
	"golang.org/x/exp/slog"
)

type crud struct {
}

func (*crud) List(c wool.Ctx) error {
    return c.JSON(http.StatusOK, "list")
}

func (*crud) Take(c wool.Ctx) error {
    return c.JSON(http.StatusOK, "take: " + c.Req().PathParamID())
}

func (*crud) Panic(c wool.Ctx) error {
    panic("panic message")
}

func main() {
	l := slog.New(slog.HandlerOptions{Level: slog.LevelDebug}.NewJSONHandler(os.Stdout))

	w := wool.New(l.WithGroup("wool"))
    w.Use(proxy.Middleware())
    w.MountHealth()
    
    crudHandlers := new(crud)
    
    w.Group("/api/v1", func(api *wool.Wool) {
        api.Group("/boards", func(b *wool.Wool) {
            b.GET("/panic", crudHandlers.Panic)
        })
        api.CRUD("/boards", crudHandlers)
    })
    
    srv := wool.NewServer(&wool.ServerConfig{
        Address: ":8080",
    }, l.WithGroup("server"))
    
    if err := srv.StartC(context.Background(), w); err != nil {
		srv.Log.Error("server error", "err", err)
		os.Exit(1)
    }
}
```

## License

Distributed under MIT License, please see license file within the code for more details.
