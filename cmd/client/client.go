package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {



	svc := NewService("http://api.qrserver.com/v1/create-qr-code/")
	data := "Hello,Frank"
	qr, err := svc.Encode(data, 100)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	err = byteConvertToPNG(qr)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	//if err := Extract(); err != nil {
	//	os.Exit(1)
	//}
}

type XmlData struct {
	ID       string  `xml:"ID,attr"`
	NumCode  int64   `xml:"NumCode"`
	CharCode string  `xml:"CharCode"`
	Nominal  int64   `xml:"Nominal"`
	Name     string  `xml:"Name"`
	Value    float64 `xml:"Value"`
}

type Curriencies struct {
	XMLName   xml.Name  `xml:"ValCurs"`
	Date      string    `xml:"Date,attr"`
	Name      string    `xml:"name,attr"`
	ValuteIds []XmlData `xml:"Valute"`
}

type JsonData struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Value float64 `json:"value"`
}

func Extract() error { // Скачивает файл, его переделываем в json и сохраняем как currencies.json
	reqUrl := "https://raw.githubusercontent.com/netology-code/bgo-homeworks/master/10_client/assets/daily.xml"
	resp, err := http.Get(reqUrl)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println(resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Println(cerr)
		}
	}()

	parsedData, err := parseXml(body)
	if err != nil {
		log.Println(err)
		return err
	}

	jsonData := parsedData.convertDataToJson()

	err = writeDataToJsonFile(jsonData)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func parseXml(data []byte) (Curriencies, error) {
	var decoded Curriencies

	err := xml.Unmarshal(data, &decoded)
	if err != nil {
		log.Println(err)
		return Curriencies{
			ValuteIds: nil,
		}, err
	}

	return decoded, nil
}

func (c *Curriencies) convertDataToJson() []JsonData {
	var jsonData []JsonData
	for _, element := range c.ValuteIds{
		jsonData = append(jsonData, JsonData{
			Code:  element.CharCode,
			Name:  element.Name,
			Value: element.Value,
		})
	}

	return jsonData
}

func writeDataToJsonFile(data []JsonData) error {
	file, err := os.Create("currencies.json")
	if err != nil {
		log.Println(err)
		return err
	}

	defer func(c io.Closer) {
		if err := c.Close(); err != nil {
			log.Println(err)
		}
	}(file)


	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type Service struct {
	baseUrl string
	client *http.Client
}

func NewService(url string) *Service {
	timeDur, _ := strconv.Atoi("CONT_TIMEOUT")
	return &Service{
		baseUrl: url,
		client:  &http.Client{Timeout: time.Second * time.Duration(timeDur)},
	}
}

func (s *Service) Encode(words string, sizeInPixels int64) ([]byte, error) {

	url := s.baseUrl + fmt.Sprintf("?data=%s&size=%dx%d", words, sizeInPixels, sizeInPixels)

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println(resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Println(cerr)
			err = cerr
		}
	}()

	return body, nil
}

func byteConvertToPNG(data []byte) error {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Println()
		return err
	}

	out, err := os.Create("./QrCode.png")
	defer func() {
		if cerr := out.Close(); cerr != nil {
			log.Println(cerr)
			err = cerr
		}
	}()

	if err != nil {
		log.Println(err)
		return err
	}

	err = png.Encode(out, img)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}