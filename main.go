package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	requestTimeout      = time.Second * 1
	requestURLBrasilAPI = "https://brasilapi.com.br/api/cep/v1/%s"
	requestURLViaCEP    = "http://viacep.com.br/ws/%s/json"
	sourceNameBrasilAPI = "Brasil API"
	sourceNameViaCEP    = "Via CEP"
)

type (
	ZipCodeData struct {
		Source       string
		CEP          string
		State        string
		City         string
		Neighborhood string
		Street       string
	}
	BrasilAPIDataResponse struct {
		CEP          string `json:"cep"`
		City         string `json:"city"`
		State        string `json:"state"`
		Neighborhood string `json:"neighborhood"`
		Street       string `json:"street"`
	}
	ViaCEPDataResponse struct {
		CEP        string `json:"cep"`
		Localidade string `json:"localidade"`
		UF         string `json:"uf"`
		Bairro     string `json:"bairro"`
		Logradouro string `json:"logradouro"`
	}
)

func (m *ZipCodeData) String() string {
	return fmt.Sprintf("CEP: %s, City: %s, State: %s, Neighborhood: %s, Street: %s", m.CEP, m.City, m.State, m.Neighborhood, m.Street)
}

func (r *BrasilAPIDataResponse) mapToZipCodeData() *ZipCodeData {
	return &ZipCodeData{
		Source:       sourceNameBrasilAPI,
		CEP:          r.CEP,
		City:         r.City,
		State:        r.State,
		Neighborhood: r.Neighborhood,
		Street:       r.Street,
	}
}

func (r *ViaCEPDataResponse) mapToZipCodeData() *ZipCodeData {
	return &ZipCodeData{
		Source:       sourceNameViaCEP,
		CEP:          r.CEP,
		City:         r.Localidade,
		State:        r.UF,
		Neighborhood: r.Bairro,
		Street:       r.Logradouro,
	}
}

func main() {
	zipCode := "01153000"
	data, err := SearchZipCode(zipCode)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Printf("Data received from '%s': %s", data.Source, data.String())
}

func SearchZipCode(zipCode string) (*ZipCodeData, error) {
	ch := make(chan *ZipCodeData)

	go getZipCodeFromBrasilAPI(zipCode, ch)
	go getZipCodeFromViaCEP(zipCode, ch)

	select {
	case msg := <-ch:
		return msg, nil
	case <-time.After(requestTimeout):
		return nil, fmt.Errorf("timeout")
	}
}

func getZipCodeFromBrasilAPI(cep string, ch chan<- *ZipCodeData) {
	req, err := http.NewRequest("GET", fmt.Sprintf(requestURLBrasilAPI, cep), nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var r BrasilAPIDataResponse
	if err = json.Unmarshal(body, &r); err != nil {
		return
	}
	ch <- r.mapToZipCodeData()
}

func getZipCodeFromViaCEP(cep string, msg chan<- *ZipCodeData) {
	req, err := http.NewRequest("GET", fmt.Sprintf(requestURLViaCEP, cep), nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var r ViaCEPDataResponse
	if err = json.Unmarshal(body, &r); err != nil {
		return
	}
	msg <- r.mapToZipCodeData()
}
