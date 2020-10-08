package reader

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"log"
	"strings"
)

//Note: always defer client.Close() after calling InitApp
func InitApp(ctx context.Context) *firestore.Client {
	conf := &firebase.Config{ProjectID: "queensland-ecourt-indexer"}
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

//TODO: could make an App interface?
func WriteDoc(ctx context.Context, client *firestore.Client, doc Document) {
	docId := strings.ReplaceAll(doc.EDocNum, "/", "")
	_, err := client.Collection("documents").Doc(docId).Set(ctx, doc)
	if err != nil {
		log.Printf("error writing document to firestore: %s", err)
	}
}

func WriteBulkDocs(ctx context.Context, client *firestore.Client, docs []Document) {
	batch := client.Batch()

	for _, doc := range(docs) {
		docId := strings.ReplaceAll(doc.EDocNum, "/", "")
		docRef := client.Collection("documents").Doc(docId)
		batch.Set(docRef, doc)
	}

	_, err := batch.Commit(ctx)
	if err != nil {
		log.Printf("error committing batch document writes: %s", err)
	}
	log.Printf("%d docs written to firestore", len(docs))
}