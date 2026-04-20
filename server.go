package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type CotacaoResponse struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type CotacaoDB struct {
	ID         uint `gorm:"primaryKey"`
	Code       string
	Bid        string
	CreateDate string
}

const url = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", Handler)

	http.ListenAndServe(":8080", mux)
}

func Handler(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)

	defer cancel()

	cotacao, err := fetchCotacao(ctx, w)
	if err != nil {
		return
	}

	err = saveCotacao(cotacao)

	if err != nil {
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(cotacao.USDBRL.Bid))
}

func fetchCotacao(ctx context.Context, w http.ResponseWriter) (*CotacaoResponse, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return nil, err
	}

	var cotacao CotacaoResponse
	err = json.NewDecoder(res.Body).Decode(&cotacao)
	if err != nil {
		http.Error(w, "Error decoding data", http.StatusInternalServerError)
		return nil, err
	}
	return &cotacao, nil

}

func saveCotacao(cotacao *CotacaoResponse) error {

	db, err := gorm.Open(sqlite.Open("mydb.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&CotacaoDB{})

	db.Create(&CotacaoDB{
		Code:       cotacao.USDBRL.Code,
		Bid:        cotacao.USDBRL.Bid,
		CreateDate: cotacao.USDBRL.CreateDate,
	})

	return nil

}
