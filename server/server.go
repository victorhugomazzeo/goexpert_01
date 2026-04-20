package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	db, err := gorm.Open(sqlite.Open("mydb.db"), &gorm.Config{})
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}

	err = db.AutoMigrate(&CotacaoDB{})

	if err != nil {
		fmt.Println("Error migrating database:", err)
		return
	}

	http.HandleFunc("/cotacao", Handler(db))
	http.ListenAndServe(":8080", nil)
}

func Handler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			fmt.Println("Method not allowed:", r.Method)
			return
		}

		cotacao, err := fetchCotacao(r.Context())
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("timeout fetching external API")
				http.Error(w, "timeout fetching data", http.StatusGatewayTimeout)
				return
			}

			http.Error(w, "Error fetching data", http.StatusInternalServerError)
			fmt.Println("Error fetching data:", err)
			return
		}

		err = saveCotacao(r.Context(), cotacao, db)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("timeout saving to database")
				http.Error(w, "timeout saving data", http.StatusInternalServerError)
				return
			}

			http.Error(w, "Error saving data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]string{
			"bid": cotacao.USDBRL.Bid,
		})

	}

}

func fetchCotacao(ctx context.Context) (*CotacaoResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var cotacao CotacaoResponse
	err = json.NewDecoder(res.Body).Decode(&cotacao)
	if err != nil {
		fmt.Println("Error decoding data:", err)
		return nil, err
	}
	return &cotacao, nil

}

func saveCotacao(ctx context.Context, cotacao *CotacaoResponse, db *gorm.DB) error {

	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	result := db.WithContext(ctx).Create(&CotacaoDB{
		Code:       cotacao.USDBRL.Code,
		Bid:        cotacao.USDBRL.Bid,
		CreateDate: cotacao.USDBRL.CreateDate,
	})

	if result.Error != nil {
		return result.Error
	}

	return nil

}
