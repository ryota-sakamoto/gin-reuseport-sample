package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sys/unix"
)

func main() {
	r := gin.Default()
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(time.Second * 10)
		c.JSON(200, gin.H{
			"message": "world",
		})
	})

	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(s uintptr) {
				err = unix.SetsockoptInt(int(s), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if err != nil {
					return
				}
			})
			return err
		},
	}
	listen, err := lc.Listen(context.Background(), "tcp4", "localhost:8080")
	if err != nil {
		panic(err)
	}

	server := http.Server{
		Handler: r,
	}

	go func() {
		if err := server.Serve(listen); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	log.Println("start shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}
}
