package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ApiResponse struct {
	Response ResponseValue `json:"USDBRL"`
}

type ResponseValue struct {
	ValorCotacao string `json:"bid"`
}

type CotacaoResponse struct {
	Valor float64
}

type Cotacao struct {
	ID    int `gorm:"primaryKey"`
	Moeda string
	Data  time.Time
	Valor float64
}

func main() {
	db, err := OpenDb()
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Cotacao{})

	http.HandleFunc("/cotacao", CotacaoHandler)
	http.ListenAndServe(":8080", nil)
}

func OpenDb() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("goexpert.db"), &gorm.Config{})
	return db, err
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	res, err := BuscaCotacao()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f, err := strconv.ParseFloat(res.Response.ValorCotacao, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = AtualizarDB(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := CotacaoResponse{Valor: f}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func BuscaCotacao() (*ApiResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	resJson, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data ApiResponse
	err = json.Unmarshal(resJson, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func AtualizarDB(valor float64) error {
	db, err := gorm.Open(sqlite.Open("goexpert.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	c := Cotacao{
		Moeda: "USD",
		Data:  time.Now(),
		Valor: valor,
	}
	db.WithContext(ctx).Create(&c)

	return nil
}
