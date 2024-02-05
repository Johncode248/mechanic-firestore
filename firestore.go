package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

func initializingFirebase() (*auth.Client, *firestore.Client) {
	opt := option.WithCredentialsFile("mechanic-bc51e-firebase-adminsdk-o14r3-658cdace63.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic(err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		panic(err)
	}

	client, err := app.Firestore(context.TODO())
	if err != nil {
		panic(err)
	}

	return authClient, client
}

var err error

func CreateField(docRef *firestore.DocumentRef, data interface{}) {
	_, err = docRef.Set(context.TODO(), data, firestore.MergeAll)
	if err != nil {
		fmt.Printf("error creating document: %v", err)
	}
}

func UpdatePath(docRef *firestore.DocumentRef, docPath string, path string, value interface{}) {

	_, err = docRef.Update(context.TODO(), []firestore.Update{

		{
			FieldPath: []string{docPath, path},
			Value:     value,
		},
	})
	if err != nil {
		fmt.Printf("error updating document:", err)
	}
	fmt.Println("update successful")
}
