package main

import (
	"html/template"
	"time"
)

type TemplateRenderer struct {
	templates *template.Template
}
type TemplateData struct {
	Message string
}

type User struct {
	Password      string    `firestore:"password"`
	DateofAccount time.Time `firestore:"date"`
	NumberofCars  int       `firestore:"numberofCars"`
	CarList       []string  `firestore:"carlist"`
}

type UserProfil struct {
	ID           string
	Password     interface{}
	CreatedAt    interface{}
	NumberofCars interface{}
}

type Car struct {
	Id     string `firestore:"id"`
	Brand  string `firestore:"brand"`
	Number string `firestore:"number"`
	Year   string `firestore:"year"`
}
