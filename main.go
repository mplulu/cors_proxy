package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mplulu/log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var port = flag.String("port", "13222", "port")

func main() {
	flag.Parse()
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))
	e.HTTPErrorHandler = customErrorHandler
	e.Use(customRecover())
	e.GET("/", serveCORSProxy)
	e.POST("/", serveCORSProxy)
	err := e.Start(fmt.Sprintf(":%v", *port))
	if err != nil {
		panic(err)
	}
}

func serveCORSProxy(c echo.Context) error {
	method := c.Request().Method
	url := c.QueryParam("url")
	body := c.Request().Body
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return errors.New(fmt.Sprintf("HTTP Request Error %v", err))
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("HTTP Request Read Content Error %v", err))
	}
	readSeeker := bytes.NewReader(content)
	http.ServeContent(c.Response(), c.Request(), "", time.Now(), readSeeker)
	return nil
}

func customErrorHandler(err error, c echo.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": false,
		"err":     err.Error(),
	})
}

func customRecover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					log.LogSeriousWithStack("panic %v", err)
					c.Error(errors.New("err:internal_error"))
				}
			}()
			return next(c)
		}
	}
}
