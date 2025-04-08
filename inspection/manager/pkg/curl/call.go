package curl

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"

	beego "github.com/beego/beego/v2/server/web"
)

func CallCurlPost(url string, jsonData interface{}) error {
	mode := beego.BConfig.RunMode

	callURL, err := beego.AppConfig.String(mode + "::" + url)
	if err != nil {
		return err
	}

	// 将数据编码为 JSON
	jsonDataBytes, err := json.Marshal(jsonData)
	if err != nil {
		log.Printf("Error marshaling JSON data: %v\n", err)
		return err
	}

	curlArgs := []string{"-v", "-H", "Content-Type: application/json", "-d", string(jsonDataBytes), callURL}

	var cmd *exec.Cmd
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd = exec.Command("curl", curlArgs...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Error executing curl command: %v\n", err)
		log.Printf("Request URL: %s\n", callURL)
		log.Printf("Request Body: %s\n", string(jsonDataBytes))
		log.Printf("Curl stderr: %s\n", stderr.String())
		return err
	}

	return nil
}
