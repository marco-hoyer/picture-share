package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"strings"
	"encoding/json"
	"archive/zip"
	"path/filepath"
)

var config Config
var working_dir string
const version = "0.9"

type Album struct {
	File string `json:"file"`
	Name string `json:"name"`
	Year string `json:"year"`
	Users []string `json:"users"`

}

type Metadata struct {
	Albums []Album `json:"albums"`
	URL string `json:"url"`
}

type Config struct {
	MetadataUrl string `json:"metadata_url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func downloadFromUrl(url string) string {
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	path := working_dir + "/" + fileName

	fmt.Println("Download von: ", url)
	output, err := os.Create(path)
	if err != nil {
		fmt.Println("Konnte Datei nicht erstellen", fileName, "-", err)
		return ""
	}
	defer output.Close()

	response := get(url)

	if response != nil {
		n, err := io.Copy(output, response.Body)
		if err != nil {
			fmt.Println("Fehler beim herunterladen von", url, "-", err)
			return ""
		}
		fmt.Println(n, "bytes heruntergeladen.")
		defer response.Body.Close()
		return path
	}
	return ""
}

func albumExists(album Album) bool {
	path := working_dir + "/" + album.Year + "/" + album.Name
	_, err := os.Stat(path)
	if err != nil {
		return false
	} else {
		return true
	}
}

func removeFile(path string) {
	os.Remove(path)
}

func get(url string) *http.Response {
	transport := &http.Transport {}

	client := &http.Client {}
	client.Transport = transport

	request, err := http.NewRequest("GET", url, nil)
        if err != nil {
                fmt.Println("Konnte", url, "nicht herunterladen. Fehler:", err)
                return nil
        }

	request.Header.Set("User-Agent", "gallery-downloader")
	request.SetBasicAuth(config.Username, config.Password)

	response, err := client.Do(request)
        if err != nil {
                fmt.Println("Konnte", url, "nicht herunterladen. Fehler:", err)
                return nil
        }
	if response.StatusCode == 401 {
		fmt.Println("Konnte", url, "nicht herunterladen. Server verweigert den Zugriff!")
		return nil
	}
	if response.StatusCode == 200 {
		return response
	} else {
                fmt.Println("Konnte", url, "nicht herunterladen. Server meldet Status:", response.StatusCode)
		return nil
	}
}

func getMetadata(url string) Metadata {
	var metadata Metadata
	response := get(url)

	if response != nil {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err == nil {
			err := json.Unmarshal(contents, &metadata)
			if err != nil {
                		fmt.Println("Konnte die Metadaten nicht entziffern: ", err)
			}
		}
	}
	return metadata
}

func unzip(archive string, destination string) bool {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		fmt.Println("Kann Datei", archive, "nicht öffnen:", err)
		return false
	}
	defer reader.Close()

	for _, f := range reader.Reader.File {
		zipped, err := f.Open()
		if err != nil {
			fmt.Println("Kann Datei", f, "nicht öffnen:", err)
			return false
		}
		defer zipped.Close()

		path := filepath.Join(working_dir, "/", destination, "/", f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, f.Mode())
			if err != nil {
				fmt.Println("Kann Datei", path, "nicht schreiben:", err)
				return false
			}
			defer writer.Close()

			_, err = io.Copy(writer, zipped)
			if err != nil {
				fmt.Println("Kann Datei", path, "nicht schreiben:", err)
				return false
			}
		}
	}
	return true
}

func isMyAlbum(album Album) bool {
	for _, user := range album.Users {
		if user == config.Username {
			return true
		} else if user == "all" {
			return true
		} else {
			return false
		}
	}
	return false
}

func readConfig() {
	path := working_dir + "/config.json"
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Konnte meine Konfiguration nicht lesen, so gehts nicht!", err)
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Konnte meine Konfiguration nicht lesen, so gehts nicht!", err)
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}
}

func getFilenameFromPath(path string) string {
	return filepath.Base(path)
}

func setDirectory() {
	binary_path, err := filepath.Abs(os.Args[0])
	directory := filepath.Dir(binary_path)
	if err != nil {
		fmt.Println("Konnte mein aktuelles Verzeichnis nicht finden:", err)
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}
	working_dir = directory
	fmt.Println("Mein Verzeichnis:", working_dir)
}

func printHeader() {
	fmt.Println("#####################################")
	fmt.Println("# Family Album Downloader           #")
	fmt.Println(fmt.Sprintf("# Version: %s                      #", version))
	fmt.Println("#####################################")
	fmt.Println("")
}

func main() {
	printHeader()
	setDirectory()
	readConfig()

	fmt.Println("### Suche nach neuen Alben ###")
	metadata := getMetadata(config.MetadataUrl)
	albums := metadata.Albums
	for i := 0; i < len(albums); i++ {
		album := albums[i]
		if isMyAlbum(album) == true {
			if (albumExists(album) == false) {
			
				fmt.Println("Neues Album gefunden:", album.Year, "/", album.Name)
				url := metadata.URL + "/" + album.File
				archive := downloadFromUrl(url)
				if archive == "" {
					break
				}
				fmt.Println("Entpacke Datei: ", getFilenameFromPath(archive)) 
				unzip(archive, album.Year)
				removeFile(archive)
			} else {
				fmt.Println("Album bereits geladen:", album.Year, "/", album.Name)
			}
		}
	}
	fmt.Println("### Fertig ###")
	time.Sleep(10 * time.Second)
}
