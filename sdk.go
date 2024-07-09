package featureflag

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type FeatureFlagSDK struct {
	host      string
	client    *http.Client
	ffDefault bool
}

func NewFeatureFlagSDK(hostFF string) *FeatureFlagSDK {
	return &FeatureFlagSDK{client: &http.Client{}, host: hostFF}
}

func (ff FeatureFlagSDK) WithDefault(ffDefault bool) FeatureFlagSDK {
	ff.ffDefault = ffDefault
	return ff
}

func (ff FeatureFlagSDK) GetFeatureFlag(key string, sessionID ...string) (bool, error) {
	url := fmt.Sprintf("%s/%s", ff.host, key)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ff.ffDefault, err
	}

	req.Header.Add("Accept", "application/json")

	if len(sessionID) != 0 {
		req.Header.Add("session_id", sessionID[0])
	}

	res, err := ff.client.Do(req)
	if err != nil {
		return ff.ffDefault, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ff.ffDefault, err
	}

	var data struct {
		Active bool `json:"active"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return ff.ffDefault, err
	}

	return data.Active, nil
}
