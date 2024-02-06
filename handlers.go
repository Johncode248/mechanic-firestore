package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func homeHandler(c echo.Context) error {
	token, err := verifyToken(c)
	if err != nil {
		cookie := new(http.Cookie)
		cookie.Name = "token"
		cookie.Value = ""
		cookie.Expires = time.Unix(0, 0)
		c.SetCookie(cookie)
		return c.Render(http.StatusOK, "home.html", nil)
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["id"].(string)

	return c.Render(http.StatusOK, "home.html", userID)
}

func loginHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method == "POST" {
			email := c.FormValue("emailName")
			password := c.FormValue("passwordName")

			// Sign in to acount Firebase Authentication
			params := (&auth.UserToCreate{}).Email(email).Password(password)
			fmt.Println(params)

			user, err := authClient.GetUserByEmail(context.Background(), email)
			if err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("Error with sign in: %v", err))
			}
			if user == nil {
				return c.String(http.StatusUnauthorized, "Incorrect email")

			} else {
				collRef := client.Collection("users")
				docRef := collRef.Doc(user.UID)
				docSnapShot, err := docRef.Get(context.TODO())
				if err != nil {
					return c.Redirect(http.StatusFound, "/login")
				}
				var passwordDb interface{}
				data := docSnapShot.Data()
				for id, val := range data {

					if id == user.UID {
						if m, ok := val.(map[string]interface{}); ok {
							firstname, found := m["password"].(string)
							if !found {
								fmt.Println("Password not found or not a string")
								return c.String(http.StatusUnauthorized, "Incorrect email")
							}
							passwordDb = firstname
						}
					}
				}
				if CheckPasswordHash(password, passwordDb.(string)) {
					err = CreateTokenWithCookie(c, user)
					if err != nil {
						return c.Redirect(http.StatusFound, "/login")
					}
					return c.Redirect(http.StatusFound, "/")

				} else {
					return c.Redirect(http.StatusFound, "/login")
				}
			}

		}
		return c.Render(http.StatusOK, "login.html", nil)
	}
}

func registrHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "rejestracja.html", nil)
}

func SignUpHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		email := c.FormValue("emailName")
		password := c.FormValue("passwordName")

		params := (&auth.UserToCreate{}).
			Email(email).
			Password(password)

		u, err := authClient.CreateUser(context.Background(), params)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			return c.Redirect(http.StatusFound, "/rejestracja")
		}

		CreateTokenWithCookie(c, u)
		password, _ = HashPassword(password)

		user := map[string]User{
			u.UID: {
				Password:      password,
				DateofAccount: time.Now(),
				NumberofCars:  0,
			},
		}
		collRef := client.Collection("users")
		docRef := collRef.Doc(u.UID)

		CreateField(docRef, user)

		return c.Redirect(http.StatusFound, "/")
	}
}

func profilHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {

		token, err := verifyToken(c)
		if err != nil {
			//return c.Render(http.StatusOK, "home.html", nil)
			return c.Redirect(http.StatusFound, "/login")
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["id"].(string)

		collRef := client.Collection("users")
		docRef := collRef.Doc(userID)
		docSnapShot, err := docRef.Get(context.TODO())
		if err != nil {
			log.Fatal("error getting document snapshot: %v", err)
		}
		data := docSnapShot.Data()
		for id, val := range data {
			fmt.Printf("%v : %v", id, val)
			fmt.Println()
		}

		submapa := data[userID]
		var userSend UserProfil
		if submapa, ok := submapa.(map[string]interface{}); ok {
			userSend = UserProfil{
				ID:           userID,
				Password:     submapa["password"],
				CreatedAt:    submapa["date"],
				NumberofCars: submapa["numberofCars"],
			}
		}
		return c.Render(http.StatusOK, "profil.html", userSend)
	}
}

func myVehicleHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := verifyToken(c)
		if err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["id"].(string)

		collRef := client.Collection("users")
		docRef := collRef.Doc(userID)

		docSnapShot, err := docRef.Get(context.TODO())
		if err != nil {
			// Log the error and return an error response to the client
			log.Printf("Error getting document snapshot: %v", err)
			return c.Redirect(http.StatusFound, "/")
		}

		data := docSnapShot.Data()

		submapa := data[userID]
		var RealList []string
		if submapa, ok := submapa.(map[string]interface{}); ok {
			carlist := submapa["carlist"]
			if carlistSlice, ok := carlist.([]interface{}); ok {
				for _, v := range carlistSlice {
					if str, ok := v.(string); ok {
						RealList = append(RealList, str)
					}
				}
			}
		}

		var Lista_produktow []Car
		var e_car Car
		for _, v := range RealList {
			sub := data[v]
			if sub, ok := sub.(map[string]interface{}); ok {
				e_car = Car{
					Brand:  sub["brand"].(string),
					Id:     sub["id"].(string),
					Number: sub["number"].(string),
					Year:   sub["year"].(string),
				}

				Lista_produktow = append(Lista_produktow, e_car)
			}
		}
		return c.Render(http.StatusOK, "myVehicle.html", Lista_produktow)
	}
}

func createVehicleHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "createVehicle.html", nil)
}

func addVehicleHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		marka := c.FormValue("markaName")
		nrRejestracji := c.FormValue("nrRejestracjiName")
		rocznik := c.FormValue("rocznikName")

		if marka == "" || nrRejestracji == "" || rocznik == "" {
			return c.Redirect(http.StatusFound, "/createVehicle")
		}

		token, err := verifyToken(c)
		if err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["id"].(string)

		collRef := client.Collection("users")
		docRef := collRef.Doc(userID)

		docSnapShot, err := docRef.Get(context.TODO())
		if err != nil {
			log.Fatal("error getting document snapshot: %v", err)
		}
		data := docSnapShot.Data()
		var numberofCars interface{}

		submapa := data[userID]
		if submapa, ok := submapa.(map[string]interface{}); ok {
			numberofCars = submapa["numberofCars"]
		}
		newId := uuid.New().String()

		numberCar := numberofCars.(int64) + 1

		carToAdd := map[string]Car{
			newId: {
				Id:     newId,
				Brand:  marka,
				Number: nrRejestracji,
				Year:   rocznik,
			},
		}

		UpdatePath(docRef, userID, "numberofCars", numberCar)
		UpdatePath(docRef, userID, "carlist", firestore.ArrayUnion(newId))

		CreateField(docRef, carToAdd) // Adding vehicle to Firestore
		return c.Redirect(http.StatusFound, "/mojepojazdy")
	}
}

func updateVehicleHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		myCar := Car{
			Id:     c.Param("idvehicle"),
			Brand:  c.Param("brandvehicle"),
			Number: c.Param("number"),
			Year:   c.Param("year"),
		}
		return c.Render(http.StatusOK, "updateVehicle.html", myCar)
	}
}

func updatingVehicle(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {

		token, err := verifyToken(c)
		if err != nil {
			return c.Redirect(302, "/login")
		}

		idVehicle := c.Param("idvehicle")
		marka := c.FormValue("markaName")
		nrRejestracji := c.FormValue("nrRejestracjiName")
		rocznik := c.FormValue("rocznikName")

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["id"].(string)

		collRef := client.Collection("users")
		docRef := collRef.Doc(userID)
		UpdatePath(docRef, idVehicle, "brand", marka)
		UpdatePath(docRef, idVehicle, "number", nrRejestracji)
		UpdatePath(docRef, idVehicle, "year", rocznik)

		return c.Redirect(http.StatusFound, "/mojepojazdy")
	}
}

func delatingVehicle(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := verifyToken(c)
		if err != nil {
			return c.Redirect(302, "/login")
		}
		idVehicle := c.Param("idvehicle")

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["id"].(string)

		collRef := client.Collection("users")
		docRef := collRef.Doc(userID)

		var wg sync.WaitGroup
		wg.Add(4)
		go func() {
			defer wg.Done()
			UpdatePath(docRef, idVehicle, "brand", firestore.Delete)

		}()
		go func() {
			defer wg.Done()
			UpdatePath(docRef, idVehicle, "number", firestore.Delete)

		}()
		go func() {
			defer wg.Done()
			UpdatePath(docRef, idVehicle, "year", firestore.Delete)

		}()
		go func() {
			defer wg.Done()
			UpdatePath(docRef, idVehicle, "id", firestore.Delete)

		}()

		wg.Wait()

		docSnapShot, err := docRef.Get(context.TODO())
		if err != nil {
			log.Printf("Error getting document snapshot: %v", err)
			return c.Redirect(http.StatusFound, "/mojepojazdy")
		}

		data := docSnapShot.Data()

		var numberofCars interface{}
		submapa := data[userID]
		if submapa, ok := submapa.(map[string]interface{}); ok {
			numberofCars = submapa["numberofCars"]
		}

		numberCar := numberofCars.(int64) - 1

		wg.Add(2)
		go func() {
			defer wg.Done()
			UpdatePath(docRef, userID, "numberofCars", numberCar)
		}()

		go func() {
			defer wg.Done()
			UpdatePath(docRef, userID, "carlist", firestore.ArrayRemove(idVehicle))

		}()

		wg.Wait()

		return c.Redirect(http.StatusFound, "/mojepojazdy")
	}
}
