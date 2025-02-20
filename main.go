package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Cotacao struct {
	Bid float64 `json:"bid,string"`
}

func main() {
	c := http.Client{}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	criaContexto(ctx)
	res, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)

	if err != nil {
		panic(err)
	}

	body, err := c.Do(res)

	if err != nil {
		panic(err)
	}

	defer body.Body.Close()

	r, err := io.ReadAll(body.Body)

	if err != nil {
		panic(err)
	}

	var cotacao Cotacao
	err = json.Unmarshal(r, &cotacao)

	if err != nil {
		panic(err)
	}
	stringValue := strconv.FormatFloat(cotacao.Bid, 'f', -1, 64)
	escreveTexto(stringValue)
}

func escreveTexto(v string) {
	file, err := os.Create("cotacao.txt")

	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprintf(file, "Dólar : %s", v)
}

func criaContexto(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Println("Tempo limite para a requisição excedido, 300ms")
		return
	default:
		fmt.Println("Resposta da API recebida e salva no bloco de notas: cotacao.txt")
	}
}
