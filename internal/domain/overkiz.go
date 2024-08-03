package domain

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
)

type Overkiz struct {
	token  string
	apiUrl string
	client *http.Client
}

func NewOverkiz(token string, host string) (*Overkiz, error) {
	o := &Overkiz{
		token:  token,
		apiUrl: fmt.Sprintf("https://%s:8443/enduser-mobile-web/1/enduserAPI", host),
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	o.client = &http.Client{Transport: tr}
	return o, nil
}

func (o *Overkiz) Devices() {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/setup/devices", o.apiUrl), nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Authorization", "Bearer "+o.token)
	resp, err := o.client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		println(string(body))
	}
	println(string(body))
}
