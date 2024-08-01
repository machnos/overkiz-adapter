package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Overkiz struct {
	token    string
	apiUrl   string
	pod      string
	userId   string
	password string
}

func NewOverkiz(regio string, pod string, userId string, password string) (*Overkiz, error) {
	o := &Overkiz{}
	host := ""
	switch regio {
	case "europe", "middle east", "africa":
		host = "ha101-1.overkiz.com"
		break
	case "asia", "pacific":
		host = "ha201-1.overkiz.com"
		break
	case "north america":
		host = "ha401-1.overkiz.com"
	default:
		return nil, fmt.Errorf("unknown overkiz regio: %s", regio)
	}
	o.apiUrl = fmt.Sprintf("https://%s/enduser-mobile-web/enduserAPI", host)
	o.pod = pod
	o.userId = userId
	o.password = password

	token, err := o.refreshToken()
	if err != nil {
		return nil, err
	}
	o.token = token
	return o, nil
}

func (o *Overkiz) refreshToken() (string, error) {
	tokenLabel := "Machnos Overkiz-adapter"

	jar, err := cookiejar.New(nil)
	if err != nil {
		// error handling
	}

	client := &http.Client{
		Jar: jar,
	}
	//Login
	data := url.Values{
		"userId":       {o.userId},
		"userPassword": {o.password},
	}
	resp, err := client.PostForm(o.apiUrl+"/login", data)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}

	// Get tokens
	resp, err = client.Get(fmt.Sprintf("%s/config/%s/local/tokens/devmode", o.apiUrl, o.pod))
	if err == nil && resp.StatusCode == http.StatusOK {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		body, _ = io.ReadAll(resp.Body)
		var responseBody []map[string]any
		err = json.Unmarshal(body, &responseBody)
		if err == nil {
			for _, token := range responseBody {
				if token["label"] == tokenLabel && token["gatewayId"] == o.pod && token["scope"] == "devmode" {
					return fmt.Sprint(token["uuid"]), nil
				}
			}
		}
	}

	// Generate token
	resp, err = client.Get(fmt.Sprintf("%s/config/%s/local/tokens/generate", o.apiUrl, o.pod))
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}

	var responseBody map[string]string
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return "", err
	}
	usertoken := responseBody["token"]

	// Activate token
	type tokenRequest struct {
		Label string `json:"label"`
		Token string `json:"token"`
		Scope string `json:"scope"`
	}

	jsonRequest, err := json.Marshal(&tokenRequest{
		Label: tokenLabel,
		Token: usertoken,
		Scope: "devmode",
	})
	if err != nil {
		return "", err
	}

	resp, err = client.Post(fmt.Sprintf("%s/config/%s/local/tokens", o.apiUrl, o.pod), "application/json", bytes.NewReader(jsonRequest))
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}
	println(string(body))
	return responseBody["token"], nil
}
