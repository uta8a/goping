package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type (
	info struct {
		Ip          string `json:"ip"`
		Host        string `json:"host"`
		CountryCode string `json:"country_code"`
		Ua          string `json:"ua"`
		Port        string `json:"port"`
		Lang        string `json:"lang"`
		Encoding    string `json:"encoding"`
		Forwarded   string `json:"forwarded"`
	}
)

func HealthCheck() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "HealthCheck!\n")
	}
}

func GetRealIP() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, c.RealIP())
	}
}

func GetAllInfo() echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := info{
			Ip          : c.RealIP(),
			Host        : req.Host,
			CountryCode : "JP not implemented",
			Ua          : "not implemented",
			Port        : "not implemented",
			Lang        : "not implemented",
			Encoding    : "not implemented",
			Forwarded   : "not implemented",
		}
		return c.JSON(http.StatusOK, res)
	}
}
