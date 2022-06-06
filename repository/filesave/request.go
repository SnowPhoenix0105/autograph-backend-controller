package filesave

import (
	"autograph-backend-controller/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type responseSchema struct {
	URL      string   `json:"url"`
	Hash     string   `json:"hash"`
	Type     string   `json:"type"`
	TextList []string `json:"text_list"`
}

func saveFile(config *Config, fileData []byte) (*responseSchema, error) {
	req, err := http.NewRequest(http.MethodPost, "http://"+config.FullHost()+"/save", bytes.NewReader(fileData))
	if err != nil {
		return nil, utils.WrapError(err, "create request fail")
	}
	req.Header.Add("Content-Type", "multipart/form-data")

	client := http.Client{
		Timeout: config.TimeOut,
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, utils.WrapError(err, "produce http request fail")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code not 200 [%d]", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("read body fail")
	}

	var schema responseSchema
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, utils.WrapError(err, "unmarshal responseSchema fail")
	}

	return &schema, nil
}

func deleteFile(config *Config, url string) error {
	body, err := json.Marshal(map[string]interface{}{
		"url": url,
	})
	if err != nil {
		return errors.New("json marshal fail")
	}

	req, err := http.NewRequest(http.MethodPost, "http://"+config.FullHost()+"/delete", bytes.NewReader(body))
	if err != nil {
		return utils.WrapError(err, "create request fail")
	}
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		Timeout: config.TimeOut,
	}
	resp, err := client.Do(req)

	if err != nil {
		return utils.WrapError(err, "produce http request fail")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not 200 [%d]", resp.StatusCode)
	}

	return nil
}

func getFile(config *Config, url string) ([]byte, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"url": url,
	})
	if err != nil {
		return nil, errors.New("json marshal fail")
	}

	req, err := http.NewRequest(http.MethodPost, "http://"+config.FullHost()+"/get", bytes.NewReader(reqBody))
	if err != nil {
		return nil, utils.WrapError(err, "create request fail")
	}
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		Timeout: config.TimeOut,
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, utils.WrapError(err, "produce http request fail")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code not 200 [%d]", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, utils.WrapError(err, "read body fail")
	}

	return respBody, nil
}
