package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	app := &cli.App{
		Name:  "readfile",
		Usage: "reading file over the web",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr",
				Usage: "web server listen address",
				Value: "localhost:8080",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
			},
			&cli.StringFlag{
				Name:     "file",
				Required: true,
			},
		},
		Action: func(cliCtx *cli.Context) error {
			filepath := cliCtx.String("file")
			ff, err := os.Stat(filepath)
			if err != nil {
				return fmt.Errorf("filepath: %s, err:%s\n", filepath, err.Error())
			}
			if ff.IsDir() {
				return fmt.Errorf("filepath: %s, err: file is a directory\n", filepath)
			}

			data, err := ioutil.ReadFile(filepath)
			if err != nil {
				return fmt.Errorf("filepath: %s, err:%s\n", filepath, err.Error())
			}
			if !cliCtx.Bool("debug") {
				gin.SetMode(gin.ReleaseMode)
			}
			e := gin.Default()
			srv := &http.Server{
				Addr:    cliCtx.String("addr"),
				Handler: e,
			}
			go func() {
				ch := make(chan os.Signal)
				signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
				<-ch
				cliCtx.Done()
				_ = srv.Shutdown(context.Background())
			}()

			fileURI := "/file/" + url.QueryEscape(ff.Name())
			e.GET(fileURI, func(ctx *gin.Context) {
				ctx.Data(http.StatusOK, "application/octet-stream", data)
			})

			log.Printf("File: %s\n", filepath)
			log.Printf("FileURI: %s\n", fileURI)
			log.Printf("Listening and serving HTTP on %s\n", srv.Addr)

			return srv.ListenAndServe()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
