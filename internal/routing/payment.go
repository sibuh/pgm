package routing

import "github.com/labstack/echo/v4"

type Route struct {
	Method      string
	Pattern     string
	Handler     echo.HandlerFunc
	Middlewares []echo.MiddlewareFunc
}

func PaymentRoutes(handler echo.HandlerFunc) []Route {
	return []Route{
		{
			Method:  "POST",
			Pattern: "/payments",
			Handler: handler,
		},
		{
			Method:  "GET",
			Pattern: "/payments/:id",
			Handler: handler,
		},
	}
}
