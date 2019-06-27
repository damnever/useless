package runtime

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const defaultTimeout = 60 * time.Second

type Function func(ctx context.Context, input string) (output string, err error)

type Supervisor struct {
	name     string
	function Function
	e        *echo.Echo
}

func NewSupervisor(name string, function Function) *Supervisor {
	s := &Supervisor{
		name:     name,
		function: function,
		e:        echo.New(),
	}
	s.e.HideBanner = true
	// e.Use(middleware.Logger())
	s.e.Use(middleware.Recover())
	s.e.POST("/", s.handle)
	s.e.GET("/meta", s.meta)
	return s
}

func (s Supervisor) Run(laddr string) error {
	errc := make(chan error, 1)
	go func() {
		errc <- s.e.Start(laddr)
	}()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigc:
		return nil
	case err := <-errc:
		return err
	}
}

func (s Supervisor) Close() error {
	return s.e.Close()
}

func (s Supervisor) meta(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{
		"name": s.name,
	})
}

func (s Supervisor) handle(c echo.Context) error {
	var (
		req struct {
			Meta  string `json:"meta"`
			Input string `json:"input"`
		}
		output string
		err    error
	)
	if err = c.Bind(&req); err == nil {
		ctx, cancel := context.WithTimeout(context.TODO(), defaultTimeout)
		defer cancel()
		output, err = s.function(ctx, req.Input)
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusBadRequest, echo.Map{
		"output": output,
	})
}
