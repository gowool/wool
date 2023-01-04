package main

import (
	"context"
	"github.com/gowool/wool"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	log, _ := zap.NewDevelopmentConfig().Build()
	w := wool.New(log)
	w.MountHealth()
	w.Group("/api/v1", func(v1 *wool.Wool) {
		v1.Group("/foo", func(foo *wool.Wool) {
			foo.Get("", func(c wool.Ctx) error {
				return c.JSON(http.StatusOK, wool.Map{
					"handler": "foo",
					"action":  "list",
				})
			})
			foo.Get("/no", func(c wool.Ctx) error {
				return c.NoContent()
			})
			foo.Get("/:id", func(c wool.Ctx) error {
				return c.JSON(http.StatusOK, wool.Map{
					"handler": "foo",
					"action":  "take",
					"id":      c.Req().PathParamID(),
				})
			})
		})
	})

	srv := wool.NewServer(wool.ServerConfig{
		Address: ":8080",
	}, log)

	if err := srv.Start(context.Background(), w); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}
