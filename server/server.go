package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func IniciaServer() {
	http.HandleFunc("/cotacao", BuscarCotacao)
	http.ListenAndServe(":8080", nil)
}

type Cotacao struct {
	Bid float64 `json:"bid,string"`
}

type MoedaBR struct {
	Cotacao Cotacao `json:"USDBRL"`
}

func BuscarCotacao(w http.ResponseWriter, r *http.Request) {
	//Criação do time de 200 milisec
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	criaRegraContextoPagina(ctx, w)

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}

	cliente := &http.Client{}
	body, err := cliente.Do(req)

	if err != nil {
		panic(err)
	}

	defer body.Body.Close()

	m := convertJsonRequisicao(body)

	t, err := json.Marshal(m.Cotacao)
	if err != nil {
		panic(err)
	}

	w.Write(t)

	salvaBancoDeDados(ctx, m)
}

func convertJsonRequisicao(r *http.Response) *MoedaBR {
	var textoJson MoedaBR

	read, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(read, &textoJson)

	return &textoJson
}

func salvaBancoDeDados(ctx context.Context, m *MoedaBR) {
	con := GetConnection()

	newctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	criaRegraContextoBanco(newctx)
	stm, err := con.PrepareContext(newctx, "INSERT INTO cotacao (bid) values (?)")

	if err != nil {
		panic(err)
	}

	defer stm.Close()

	_, err = stm.Exec(m.Cotacao.Bid)

	if err != nil {
		panic(err)
	}
}

func criaRegraContextoPagina(ctx context.Context, w http.ResponseWriter) {
	select {
	case <-ctx.Done():
		log.Println("tempo limite da requisição atingida, 200ms")
		http.Error(w, "Tempo limite da requisição atingido", http.StatusRequestTimeout)
		return
	default:
		log.Println("Retorno da API feita com sucesso")
	}

}

func criaRegraContextoBanco(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Println("tempo limite da requisição atingida 10ms, não foi possivel salvar os dados no banco")
		return
	default:
		log.Println("Request salvo no banco")
	}
}

func GetConnection() *sql.DB {
	db, err := sql.Open("sqlite3", "./goexpert.db?cache=shared")

	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacao(bid NUMERIC(10,0), create_date DATE DEFAULT CURRENT_DATE);")

	if err != nil {
		panic("Erro ao criar a tabela padrão")
	}

	return db
}
