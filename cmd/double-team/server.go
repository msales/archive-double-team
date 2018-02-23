package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/msales/double-team"
	"github.com/msales/double-team/server"
	"github.com/msales/double-team/server/middleware"
	"gopkg.in/urfave/cli.v1"
)

func runServer(c *cli.Context) {
	ctx, err := newContext(c)
	if err != nil {
		log.Fatal(err.Error())
	}

	app, err := newApplication(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	port := c.String(FlagPort)
	s := newServer(ctx, app)
	ctx.logger.Info(fmt.Sprintf("Starting server on port %s", port))
	log.Fatal(http.ListenAndServe(":"+port, s))
}

func newServer(ctx *Context, app *double_team.Application) http.Handler {
	s := server.New(app)

	h := middleware.Common(s)
	return middleware.WithContext(ctx, h)
}
