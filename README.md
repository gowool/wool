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
    
    "github.com/gowool/middleware/logger"
    "github.com/gowool/middleware/proxy"
    "github.com/gowool/wool"
    "go.uber.org/zap"
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
    log, _ := zap.NewDevelopmentConfig().Build()
    w := wool.New(wool.WithLog(log))
    w.Use(
        proxy.Middleware(),
        logger.Middleware(logger.Config{
            ExcludeRegexEndpoint: "^/favicon.ico",
        }),
    )
    w.MountHealth()
    
    crudHandlers := new(crud)
    
    w.Group("/api/v1", func(api *wool.Wool) {
        api.Group("/boards", func(b *wool.Wool) {
            b.Get("/panic", crudHandlers.Panic)
        })
        api.CRUD("/boards", crudHandlers)
    })
    
    srv := wool.NewServer(&wool.ServerConfig{
        Address: ":8080",
    })
	srv.Log = log
    
    if err := srv.StartC(context.Background(), w); err != nil {
        log.Fatal("server error", zap.Error(err))
    }
}
```

## License

Distributed under MIT License, please see license file within the code for more details.
