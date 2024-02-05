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

	return c.Render(http.StatusOK, "home.html", userID) // UserId
}

func loginHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method == "POST" {
			email := c.FormValue("emailName")
			password := c.FormValue("passwordName")

			// Logowanie do konta w Firebase Authentication
			params := (&auth.UserToCreate{}).Email(email).Password(password)
			fmt.Println(params)

			user, err := authClient.GetUserByEmail(context.Background(), email)
			if err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("Błąd logowania: %v", err))
			}
			if user == nil {
				c.String(http.StatusUnauthorized, "Bład nieprawidłowy email")

			} else {
				collRef := client.Collection("users")
				docRef := collRef.Doc(user.UID)
				docSnapShot, err := docRef.Get(context.TODO())
				if err != nil {
					log.Fatal("error getting document snapshot: %v", err)
				}
				var passwordDb interface{}
				data := docSnapShot.Data()
				for id, val := range data {
					//fmt.Printf("%v : %v", id, val)
					if id == user.UID {
						//passwordDb = val
						if m, ok := val.(map[string]interface{}); ok {
							// Access the value using the key "firstname"
							firstname, found := m["password"].(string)
							if found {
								fmt.Println("Firstname:", firstname)
							} else {
								fmt.Println("Firstname not found or not a string")
							}
							// Assign the map to passwordDb if needed
							passwordDb = firstname
						}
					}

				}
				if passwordDb == password {
					fmt.Println(passwordDb, "==", password)
					err = CreateTokenWithCookie(c, user)
					if err != nil {
						panic(err)
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

func rejestrHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "rejestracja.html", nil)
}

func potwierdzenieHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		email := c.FormValue("emailName")
		password := c.FormValue("passwordName")

		params := (&auth.UserToCreate{}).
			Email(email).
			Password(password)

		u, err := authClient.CreateUser(context.Background(), params)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Error creating user: "+err.Error())
		}

		CreateTokenWithCookie(c, u)

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

		return c.Render(http.StatusOK, "potwierdzenie.html", nil)
	}
}

func profilHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {

		token, err := verifyToken(c)
		if err != nil {
			return c.Render(http.StatusOK, "home.html", nil)
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

		submapa, _ := data[userID]
		var userSend UserProfil
		if submapa, ok := submapa.(map[string]interface{}); ok {
			wFirst, _ := submapa["password"]
			wLast, _ := submapa["date"]
			numberofCars, _ := submapa["numberofCars"]

			userSend.ID = userID
			userSend.Password = wFirst
			userSend.CreatedAt = wLast
			userSend.NumberofCars = numberofCars
		}

		return c.Render(http.StatusOK, "profil.html", userSend)
	}
}

func myVehicleHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := verifyToken(c)
		if err != nil {
			// Return an error response to the client
			return c.Render(http.StatusBadRequest, "error.html", map[string]string{"message": "Invalid token"})
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["id"].(string)

		collRef := client.Collection("users")
		docRef := collRef.Doc(userID)

		docSnapShot, err := docRef.Get(context.TODO())
		if err != nil {
			// Log the error and return an error response to the client
			log.Printf("Error getting document snapshot: %v", err)
			return c.Render(http.StatusInternalServerError, "error.html", map[string]string{"message": "Internal Server Error"})
		}

		data := docSnapShot.Data()

		submapa, _ := data[userID]
		var RealList []string
		if submapa, ok := submapa.(map[string]interface{}); ok {
			carlist, _ := submapa["carlist"]
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
			sub, _ := data[v]
			if sub, ok := sub.(map[string]interface{}); ok {
				brand := sub["brand"]
				id := sub["id"]
				number := sub["number"]
				year := sub["year"]

				e_car.Brand = brand.(string)
				e_car.Id = id.(string)
				e_car.Number = number.(string)
				e_car.Year = year.(string)

				Lista_produktow = append(Lista_produktow, e_car)
			}
		}
		fmt.Println(Lista_produktow)
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

		token, err := verifyToken(c)
		if err != nil {
			return c.Render(http.StatusOK, "home.html", nil)
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

		submapa, _ := data[userID]
		if submapa, ok := submapa.(map[string]interface{}); ok {
			numberofCars, _ = submapa["numberofCars"]
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

		CreateField(docRef, carToAdd) // Dodawanie pojazdu
		return c.Redirect(http.StatusFound, "/mojepojazdy")

		//return c.Render(http.StatusOK, "myVehicle.html", nil)
	}
}

func updateVehicleHandler(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		idVehicle := c.Param("idvehicle")
		brandvehicle := c.Param("brandvehicle")
		number := c.Param("number")
		year := c.Param("year")

		myCar := Car{
			Id:     idVehicle,
			Brand:  brandvehicle,
			Number: number,
			Year:   year,
		}
		return c.Render(http.StatusOK, "updateVehicle.html", myCar)
	}
}

func updatingVehicle(authClient *auth.Client, client *firestore.Client) echo.HandlerFunc {
	return func(c echo.Context) error {

		idVehicle := c.Param("idvehicle")
		marka := c.FormValue("markaName")
		nrRejestracji := c.FormValue("nrRejestracjiName")
		rocznik := c.FormValue("rocznikName")

		token, err := verifyToken(c)
		if err != nil {
			return c.Redirect(302, "/login")
		}
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

		//UpdatePath(docRef, userID, "numberofCars", liczbaAut)

		docSnapShot, err := docRef.Get(context.TODO())
		if err != nil {
			// Log the error and return an error response to the client
			log.Printf("Error getting document snapshot: %v", err)
			return c.Render(http.StatusInternalServerError, "error.html", map[string]string{"message": "Internal Server Error"})
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
