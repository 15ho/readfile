package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

func netIPv4() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	var ipv4 []string
	for _, addr := range addrs {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ipv4 = append(ipv4, ipNet.IP.String())
		}
	}
	return ipv4
}

func main() {
	app := &cli.App{
		Name:  "readfile",
		Usage: "reading file over the web",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "port",
				Usage: "web server port",
				Value: 8080,
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

			if !cliCtx.Bool("debug") {
				gin.SetMode(gin.ReleaseMode)
			}
			e := gin.Default()
			port := cliCtx.Uint("port")
			srv := &http.Server{
				Addr:    fmt.Sprintf(":%d", port),
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
				ctx.File(filepath)
			})

			log.Printf("File: %s\n", filepath)
			log.Printf("FileURI: %s\n", fileURI)
			for idx, host := range netIPv4() {
				log.Printf("FileURL(%d): http://%s:%d%s \n", idx, host, port, fileURI)
			}
			log.Printf("Listening and serving HTTP on %s\n", srv.Addr)

			return srv.ListenAndServe()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
