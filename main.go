package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type AddressInfo struct {
	CEP          string `json:"cep"`
	Logradouro   string `json:"logradouro"`
	Bairro       string `json:"bairro"`
	Localidade   string `json:"localidade"`
	UF           string `json:"uf"`
	APIProvider  string `json:"api_provider"`
	ResponseTime string `json:"response_time"`
}

func main() {
	cep := "88905440"

	// Canal para receber os resultados das requisições
	resultChan := make(chan *AddressInfo, 2)

	// WaitGroup para aguardar o término de ambas as goroutines
	wg := sync.WaitGroup{}

	wg.Add(1)
	// Inicia a primeira goroutine para chamar a primeira API do BrasilAPI
	go callAPI("https://brasilapi.com.br/api/cep/v1/"+cep, "BrasilAPI", &wg, resultChan)

	wg.Add(1)
	// Inicia a segunda goroutine para chamar a segunda API do VIACEP
	go callAPI("https://viacep.com.br/ws/"+cep+"/json/", "ViaCEP", &wg, resultChan)

	// Aguarda o término de ambas as goroutines
	wg.Wait()

	// Fecha o canal para garantir que a função main não fique bloqueada
	close(resultChan)

	// Processa os resultados
	var fastestAddress *AddressInfo
	var fastestTime time.Duration

	for result := range resultChan {
		fmt.Printf("API Provider: %s\n", result.APIProvider)
		fmt.Printf("CEP: %s\n", result.CEP)
		fmt.Printf("Logradouro: %s\n", result.Logradouro)
		fmt.Printf("Bairro: %s\n", result.Bairro)
		fmt.Printf("Localidade: %s\n", result.Localidade)
		fmt.Printf("UF: %s\n", result.UF)
		fmt.Printf("Tempo de Resposta: %s\n", result.ResponseTime)
		fmt.Println("")

		// Verifica se é a resposta mais rápida
		if fastestAddress == nil || time.Since(time.Now().Add(-time.Second)) < fastestTime {
			fastestAddress = result
			fastestTime = time.Since(time.Now())
		}
	}

	// Exibe a resposta mais rápida
	fmt.Printf("API mais rápido nesse resultado foi: %s\n", fastestAddress.APIProvider)
}

func callAPI(apiURL, apiProvider string, wg *sync.WaitGroup, resultChan chan *AddressInfo) {
	defer wg.Done()
	startTime := time.Now()
	addressInfo, err := getAddressInfo(apiURL, apiProvider)
	if err == nil {
		addressInfo.ResponseTime = time.Since(startTime).String()
		resultChan <- addressInfo
	}
}

func getAddressInfo(apiURL, apiProvider string) (*AddressInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var addressInfo AddressInfo
	err = json.Unmarshal(body, &addressInfo)
	if err != nil {
		return nil, err
	}

	addressInfo.APIProvider = apiProvider
	return &addressInfo, nil
}
