package controllers

import (
	"errors"
	"fmt"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const backupFormat = "%s/%s/%s_%s"
const backupPath = "backups"

func NewDatabaseBackupsController(scwClient *scw.Client, DatabaseIDsEnv string) *DatabaseBackupsController {
	return &DatabaseBackupsController{
		databaseIDs: strings.Split(DatabaseIDsEnv, ","),
		scwClient:   scwClient,
	}
}

func (d *DatabaseBackupsController) startInventory() {
	dbAPI := rdb.NewAPI(d.scwClient)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		os.Mkdir(backupPath, 0700)
	}

	for _, dbID := range d.databaseIDs {
		id, region, err := getRegionalizedID(dbID)
		if err != nil {
			log.Printf("could not get id and region from %s: %v\n", dbID, err)
			continue
		}

		backupInstanceIdPath := fmt.Sprintf("%s/%s", backupPath, dbID)
		if _, err := os.Stat(backupInstanceIdPath); os.IsNotExist(err) {
			os.Mkdir(backupInstanceIdPath, 0700)
		}

		var backups []*rdb.DatabaseBackup
		pageSize := uint32(10)
		currentPage := int32(1)
		totalCount := uint32(math.MaxInt32)
		for uint32(currentPage) * pageSize < totalCount {
			dbBackupsReponse, err := dbAPI.ListDatabaseBackups(&rdb.ListDatabaseBackupsRequest{
				Region:     scw.Region(region),
				InstanceID: &id,
				PageSize: &pageSize,
				Page: &currentPage,
			})
			if err != nil {
				log.Printf("could not get rdb instance %s: %v\n", id, err)
				continue
			}

			backups = append(backups, dbBackupsReponse.DatabaseBackups...)

			totalCount = dbBackupsReponse.TotalCount
			currentPage += 1
		}


		for _, backup := range backups {
			fileName := fmt.Sprintf(backupFormat, backupPath, dbID, backup.DatabaseName, backup.CreatedAt.Format(time.RFC3339))
			if !fileExists(fileName) {
				d.downloadFile(dbAPI, backup, fileName)
			}
		}

		d.deleteOldBackups(backups, dbID)
	}
}

func (d *DatabaseBackupsController) deleteOldBackups(backups []*rdb.DatabaseBackup, instanceId string) {
	var files []string

	folderPath := fmt.Sprintf("%s/%s", backupPath, instanceId)
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		files = append(files, info.Name())
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		contained := false
		for _, backup := range backups {
			fileName := fmt.Sprintf("%s_%s", backup.DatabaseName, backup.CreatedAt.Format(time.RFC3339))
			if file == fileName {
				contained = true
				break
			}
		}
		if !contained {
			err := os.Remove(fmt.Sprintf("%s/%s/%s", backupPath, instanceId, file))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (d *DatabaseBackupsController) downloadFile(dbAPI *rdb.API, backup *rdb.DatabaseBackup, fileName string) {
	log.Printf("Clerk is archiving file %s\n", fileName)
	newBackup, _ := dbAPI.ExportDatabaseBackup(&rdb.ExportDatabaseBackupRequest{
		DatabaseBackupID: backup.ID,
		Region:           backup.Region,
	})

	for ok := true; ok; ok = newBackup.DownloadURL == nil {
		newBackup, _ = dbAPI.GetDatabaseBackup(&rdb.GetDatabaseBackupRequest{
			Region:           newBackup.Region,
			DatabaseBackupID: newBackup.ID,
		})
		fmt.Println("Backup not ready... retrying in 10 seconds")
		time.Sleep(10 * time.Second)
	}

	err := downloadFile(*newBackup.DownloadURL, fileName)
	if err != nil {
		fmt.Printf("Unexpected error while archiving file: %s\n", err)
	}
}

func (d *DatabaseBackupsController) Run() {
	for {
		d.startInventory()
		log.Println("The clerk is done archiving. See you in an hour.")
		time.Sleep(1 * time.Hour)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func getRegionalizedID(r string) (string, string, error) {
	split := strings.Split(r, "/")
	switch len(split) {
	case 1:
		return split[0], os.Getenv("SCW_DEFAULT_REGION"), nil
	case 2:
		return split[1], split[0], nil
	default:
		return "", "", fmt.Errorf("couldn't parse ID %s", r)
	}
}
