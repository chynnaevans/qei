package reader

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"log"
)

//Note: always defer client.Close() after calling InitApp
func InitApp() *firestore.Client {
	conf := &firebase.Config{ProjectID: "queensland-ecourt-indexer"}
	ctx := context.Background()
	sa := option.WithCredentialsFile("../../Downloads/queensland-ecourt-indexer-firebase-adminsdk-36j5c-17991ecb16.json")
	app, err := firebase.NewApp(ctx, conf, sa)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	return client
}