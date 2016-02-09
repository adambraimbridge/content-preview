package main

import (
	"os"
	"io"
	"log"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"

)

func main() {
	app := cli.App("content-preview", "A RESTful API for retrieving and transforming content preview data")
	appPort := app.StringOpt("app-port", "8084", "Default port for app")
	mapiAuth := app.StringOpt("mapi-auth", "default", "Basic auth for MAPI")
	mapiUri := app.StringOpt("mapi-uri", "http://methode-api-uk-p.svc.ft.com/eom-file/", "Host and path for MAPI")
	matHostHeader := app.StringOpt("mat-host-header", "methode-article-transformer", "Hostheader for MAT")
	matUri := app.StringOpt("mat-uri", "http://ftapp05951-lvpr-uk-int:8080/content-transform/", "Host and path for MAT")

	app.Action = func() {
		r := mux.NewRouter()
		handler := Handlers{*mapiAuth, *mapiUri, *matUri, *matHostHeader}
		r.HandleFunc("/content-preview/{uuid}", handler.contentPreviewHandler)
		r.HandleFunc("/build-info", handler.buildInfoHandler)
		r.HandleFunc("/ping", pingHandler)
		http.Handle("/", r)

		log.Fatal(http.ListenAndServe(":" + *appPort, nil))

	}
	app.Run(os.Args)

}

type Handlers struct {
	mapiAuth string
	mapiUri string
	matUri string
	matHostHeader string
}

func pingHandler(w http.ResponseWriter, r *http.Request){

}

func (h Handlers) buildInfoHandler(w http.ResponseWriter, r *http.Request){

}

func (h Handlers) contentPreviewHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request");

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		log.Fatal("missing UUID")
	}

	methode := h.mapiUri + uuid

	log.Printf("sending to MAPI at " + methode);

	client := &http.Client{}
	mapReq, err := http.NewRequest("GET", methode, nil)
	mapReq.Header.Set("Authorization", "Basic " + h.mapiAuth)
	mapiResp, err := client.Do(mapReq)


	if err !=nil {
		log.Fatal(err)
	}

	fmt.Printf("mapi the status code %v", mapiResp.StatusCode)
	if mapiResp.StatusCode !=200 {
		if mapiResp.StatusCode == 404 {
			w.WriteHeader(http.StatusNotFound);
			return
		}
		//TODO break this down
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	// order of writing a response
	//header
	//responseCode
	//body

	matUrl := h.matUri + uuid
	log.Printf("sending to MAT at " + matUrl);
	client2 := &http.Client{}
	matReq, err := http.NewRequest("POST", matUrl, mapiResp.Body)
	matReq.Header.Add("Host", h.matHostHeader)
	matReq.Header.Add("Content-Type", "application/json")
	matResp, err := client2.Do(matReq)

	if matResp.StatusCode !=200 {
		//TODO break this down
		fmt.Printf("---the status code %v", matResp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("mat the status code %v", matResp.StatusCode)
	io.Copy(w, matResp.Body)
}

