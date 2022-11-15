package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"text/template"
	"time"
)

type ApiResponse struct {
	Valor float64
}

func main() {
	data, err := BuscaCotacao()
	if err != nil {
		panic(err)
	}

	err = CriarArquivo(data)
	if err != nil {
		panic(err)
	}

}

func CriarArquivo(data *ApiResponse) error {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	tmp := template.Must(template.New("CotacaoTemplate").Parse("Dólar: {{.Valor}}"))
	err = tmp.Execute(f, data)
	if err != nil {
		return err
	}

	return nil
}

func BuscaCotacao() (*ApiResponse, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
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
