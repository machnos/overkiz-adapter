package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go.nhat.io/cookiejar"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strconv"
	"time"
)

type OverkizTokenApi struct {
	apiUrl string
}

func NewOverkizTokenApi(regio string) (*OverkizTokenApi, error) {
	o := &OverkizTokenApi{}
	tokenHost := ""
	switch regio {
	case "europe", "middle east", "africa":
		tokenHost = "ha101-1.overkiz.com"
		break
	case "asia", "pacific":
		tokenHost = "ha201-1.overkiz.com"
		break
	case "north america":
		tokenHost = "ha401-1.overkiz.com"
	default:
		return nil, fmt.Errorf("unknown overkiz regio: %s", regio)
	}
	o.apiUrl = fmt.Sprintf("https://%s/enduser-mobile-web/enduserAPI", tokenHost)
	return o, nil
}

func (o *OverkizTokenApi) client() (*http.Client, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	cookieDir := usr.HomeDir + "/.machnos/overkiz-token/"
	err = os.MkdirAll(cookieDir, 0770)
	if err != nil {
		return nil, err
	}

	jar := cookiejar.NewPersistentJar(
		cookiejar.WithFilePath(cookieDir+"cookies.json"),
		cookiejar.WithFilePerm(0660),
		cookiejar.WithAutoSync(true),
	)

	return &http.Client{
		Jar: jar,
	}, nil
}

func (o *OverkizTokenApi) Login(username string, password string) error {
	client, err := o.client()
	if err != nil {
		return err
	}
	// Check already logged in
	resp, err := client.Get(o.apiUrl + "/authenticated")
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var responseBody map[string]bool
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return err
	}
	if responseBody["authenticated"] {
		fmt.Println("Already logged in")
		return nil
	}

	//Login
	data := url.Values{
		"userId":       {username},
		"userPassword": {password},
	}
	resp, err = client.PostForm(o.apiUrl+"/login", data)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		println(string(body))
		return fmt.Errorf("unable to login to %s. status code %d", resp.Request.URL, resp.StatusCode)
	}
	fmt.Println("Logged in!")
	return nil
}

func (o *OverkizTokenApi) Logout() error {
	client, err := o.client()
	if err != nil {
		return err
	}
	resp, err := client.PostForm(o.apiUrl+"/logout", url.Values{})
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println(string(body))
		return err
	}
	fmt.Println("Logged out")
	return nil
}

func (o *OverkizTokenApi) PrintTokens(pod string) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("%s/config/%s/local/tokens/devmode", o.apiUrl, pod))
	if err == nil && resp.StatusCode == http.StatusOK {
		type listToken struct {
			Label        string
			CreationTime string
			UUID         string
			Scope        string
		}

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var responseBody []map[string]any
		err = json.Unmarshal(body, &responseBody)
		if err != nil {
			return err
		}
		tokens := make([]listToken, 0)
		for _, token := range responseBody {
			if token["gatewayId"] == pod {
				creationTime, err := strconv.ParseInt(fmt.Sprintf("%.0f", token["gatewayCreationTime"]), 10, 64)
				if err != nil {
					return err
				}
				tokens = append(tokens, listToken{
					Label:        fmt.Sprint(token["label"]),
					CreationTime: fmt.Sprintf("%v", time.Unix(creationTime/1000, 0)),
					UUID:         fmt.Sprint(token["uuid"]),
					Scope:        fmt.Sprint(token["scope"]),
				})
			}
		}
		spacing := 2
		headerLabel := "Label"
		headerCreationTime := "Creation time"
		headerUUID := "UUID"
		headerScope := "Scope"
		maxLabelLength := len(headerLabel) + spacing
		maxCreationTimeLength := len(headerCreationTime) + spacing
		maxUUIDLength := len(headerUUID) + spacing
		maxScopeLength := len(headerScope) + spacing
		for _, token := range tokens {
			if len(token.Label) > maxLabelLength {
				maxLabelLength = len(token.Label) + spacing
			}
			if len(token.CreationTime) > maxCreationTimeLength {
				maxCreationTimeLength = len(token.CreationTime) + spacing
			}
			if len(token.UUID) > maxUUIDLength {
				maxUUIDLength = len(token.UUID) + spacing
			}
			if len(token.Scope) > maxScopeLength {
				maxScopeLength = len(token.Scope) + spacing
			}
		}
		lengths := []int{maxLabelLength, maxCreationTimeLength, maxUUIDLength, maxScopeLength}
		o.printLine([]string{headerLabel, headerCreationTime, headerUUID, headerScope}, lengths, ' ')
		o.printLine([]string{"=", "=", "=", "="}, lengths, '=')
		for _, token := range tokens {
			o.printLine([]string{token.Label, fmt.Sprintf("%v", token.CreationTime), token.UUID, token.Scope}, lengths, ' ')
		}
	}
	return nil
}

func (o *OverkizTokenApi) printLine(columns []string, columnLength []int, fillChar rune) {
	for ix, column := range columns {
		fmt.Print(column)
		fill := string(fillChar)
		for i := len(column); i < columnLength[ix]; i++ {
			fmt.Print(fill)
		}
	}
	fmt.Println()
}

func (o *OverkizTokenApi) CreateAndPrintToken(pod string, label string) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	// Generate token
	resp, err := client.Get(fmt.Sprintf("%s/config/%s/local/tokens/generate", o.apiUrl, pod))
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	var responseBody map[string]string
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return err
	}
	usertoken := responseBody["token"]

	// Activate token
	type tokenRequest struct {
		Label string `json:"label"`
		Token string `json:"token"`
		Scope string `json:"scope"`
	}

	jsonRequest, err := json.Marshal(&tokenRequest{
		Label: label,
		Token: usertoken,
		Scope: "devmode",
	})
	if err != nil {
		return err
	}
	resp, err = client.Post(fmt.Sprintf("%s/config/%s/local/tokens", o.apiUrl, pod), "application/json", bytes.NewReader(jsonRequest))
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return err
	}
	fmt.Printf("Token: %s", responseBody["token"])
	fmt.Println()
	fmt.Println("Please store the token at a safe place!! You will never be able to view it again.")
	return nil
}

func (o *OverkizTokenApi) DeleteToken(pod string, uuid string) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/config/%s/local/tokens/%s", o.apiUrl, pod, uuid), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}
	fmt.Println("Token deleted")
	return nil
}

func (o *OverkizTokenApi) Doc() error {
	client, err := o.client()
	if err != nil {
		return err
	}

	// Generate token
	resp, err := client.Get(fmt.Sprintf("%s/doc", o.apiUrl))
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}
	fmt.Println(string(body))
	return nil
}
