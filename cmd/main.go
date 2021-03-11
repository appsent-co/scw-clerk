package main

import (
	"github.com/scaleway/scaleway-sdk-go/scw"
	"log"
	"os"
	"scw-clerk/controllers"
)

func main() {
	log.Println("The clerk is heading to work")

	scwKey := os.Getenv("SCW_KEY")
	if scwKey == "" {
		log.Fatalln("The clerk is lost. You need to specify your Scaleway key in the SCW_KEY env variable")
	}

	scwSecret := os.Getenv("SCW_SECRET")
	if scwKey == "" {
		log.Fatalln("The clerk is lost. You need to specify your Scaleway Secret in the SCW_SECRET env variable")
	}

	scwOrgID := os.Getenv("SCW_ORG_ID")
	if scwKey == "" {
		log.Fatalln("The clerk is lost. You need to specify your Scaleway Organization ID in the SCW_ORG_ID env variable")
	}

	client, err := scw.NewClient(
		// Get your credentials at https://console.scaleway.com/project/credentials
		scw.WithDefaultOrganizationID(scwOrgID),
		scw.WithAuth(scwKey, scwSecret),
	)
	if err != nil {
		panic(err)
	}

	databaseIds := os.Getenv("DATABASE_IDS")
	if databaseIds == "" {
		log.Fatalln("The clerk has nothing to do. You need to specify your database IDs in the DATABASE_IDS env variable")
	}

	backupsController := controllers.NewDatabaseBackupsController(client, databaseIds)
	backupsController.Run()
}
