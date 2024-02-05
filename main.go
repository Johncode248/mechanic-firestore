package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	authClient, client := initializingFirebase()

	e := echo.New()

	templates := template.Must(template.ParseGlob("templates/*.html"))
	e.Renderer = &TemplateRenderer{
		templates: templates,
	}
	e.GET("/", homeHandler)
	e.GET("/login", loginHandler(authClient, client))
	e.POST("/login", loginHandler(authClient, client))
	e.GET("/rejestracja", rejestrHandler)
	e.POST("/potwierdzenie", potwierdzenieHandler(authClient, client))
	e.GET("/profil", profilHandler(authClient, client))
	e.GET("/mojepojazdy", myVehicleHandler(authClient, client))
	e.GET("/createVehicle", createVehicleHandler)
	e.POST("/addVehicle", addVehicleHandler(authClient, client))
	e.GET("/update/:idvehicle/:brandvehicle/:number/:year", updateVehicleHandler(authClient, client))
	e.POST("/updatingVehicle/:idvehicle", updatingVehicle(authClient, client))
	e.GET("/deleteCar/:idvehicle", delatingVehicle(authClient, client))
	e.Logger.Fatal(e.Start(":1323"))
}
