package domain

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Overkiz struct {
	token        string
	apiUrl       string
	client       *http.Client
	devices      []*Device
	updateTicker *time.Ticker
}

type Device struct {
	Label     string `json:"label"`
	Class     string `json:"class"`
	DeviceURL string `json:"device_url"`
}

func NewOverkiz(token string, host string, context context.Context) (*Overkiz, error) {
	o := &Overkiz{
		token:  token,
		apiUrl: fmt.Sprintf("https://%s:8443/enduser-mobile-web/1/enduserAPI", host),
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	o.client = &http.Client{Transport: tr}
	o.updateTicker = time.NewTicker(time.Minute * 5)
	devices, err := o.loadDevices()
	if err != nil {
		fmt.Printf("Failed to load devices: %v", err)
	} else {
		o.devices = devices
	}
	go func() {
		for {
			select {
			case <-context.Done():
				return
			case <-o.updateTicker.C:
				devices, err := o.loadDevices()
				if err != nil {
					fmt.Printf("Failed to load devices: %v", err)
				} else {
					o.devices = devices
				}
			}
		}
	}()
	return o, nil
}

func (o *Overkiz) loadDevices() ([]*Device, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/setup/devices", o.apiUrl), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.token)
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		println(string(body))
	}
	var responseBody []map[string]any
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return nil, err
	}
	devices := make([]*Device, 0)
	for _, device := range responseBody {
		devices = append(devices, &Device{
			Label:     device["label"].(string),
			DeviceURL: device["deviceURL"].(string),
			Class:     (device["definition"].(map[string]any))["uiClass"].(string),
		})
	}
	return devices, nil
}

func (o *Overkiz) Devices(class string) []*Device {
	result := make([]*Device, 0)
	if class == "" {
		return o.devices
	}
	for _, device := range o.devices {
		if device.Class == class {
			result = append(result, device)
		}
	}
	return result
}

func (o *Overkiz) RollerShutters(actionName string) (int, error) {
	devices := o.Devices("RollerShutter")
	if len(devices) == 0 {
		return 0, nil
	}
	ar := &actionRequest{
		Label: actionName + "RollerShutters",
	}
	for _, device := range devices {
		ac := &action{
			DeviceURL: device.DeviceURL,
		}
		ac.Commands = append(ac.Commands, &command{
			Name: actionName,
		})
		ar.Actions = append(ar.Actions, ac)
	}
	reqData, err := json.Marshal(ar)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/exec/apply", o.apiUrl), bytes.NewBuffer(reqData))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+o.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.client.Do(req)
	if err != nil {
		return 0, err
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%s", string(body))
	}
	var responseBody map[string]any
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return 0, err
	}
	return len(devices), nil
}
