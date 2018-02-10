package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HttpPost(url, token string, payload []byte) ([]byte, error) {
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(payload)),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("--- resp: \n%s\n", string(b))
	defer resp.Body.Close()

	return b, err
}
