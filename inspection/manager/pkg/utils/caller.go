package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	beego "github.com/beego/beego/v2/server/web"
)

func CallHttpPost(url string, jsonData interface{}) error {
	mode := beego.BConfig.RunMode

	callURL, err := beego.AppConfig.String(mode + "::" + url)
	if err != nil {
		return err
	}

	jsonDataBytes, err := json.Marshal(jsonData)
	if err != nil {
		log.Printf("Error marshaling JSON data: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", callURL, bytes.NewBuffer(jsonDataBytes))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error executing HTTP request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("HTTP request failed with status: %d, response: %s", resp.StatusCode, string(body))
		return errors.New("HTTP request failed")
	}

	return nil
}
