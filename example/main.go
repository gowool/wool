package main

import (
	"context"
	"github.com/gowool/wool"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.New(slog.HandlerOptions{Level: slog.LevelDebug}.NewJSONHandler(os.Stdout))

	w := wool.New(logger.WithGroup("wool"))
	w.MountHealth()
	w.Group("/api/v1", func(v1 *wool.Wool) {
		v1.Group("/foo", func(foo *wool.Wool) {
			foo.GET("", func(c wool.Ctx) error {
				return c.JSON(http.StatusOK, wool.Map{
					"handler": "foo",
					"action":  "list",
				})
			})
			foo.GET("/no", func(c wool.Ctx) error {
				return c.NoContent()
			})
			foo.GET("/:id", func(c wool.Ctx) error {
				return c.JSON(http.StatusOK, wool.Map{
					"handler": "foo",
					"action":  "take",
					"id":      c.Req().PathParamID(),
				})
			})
		})
	})

	srv := wool.NewServer(&wool.ServerConfig{
		Address: ":8080",
	}, logger.WithGroup("server"))

	if err := srv.StartC(context.Background(), w); err != nil {
		srv.Log.Error("server error", err)
		os.Exit(1)
	}
}
