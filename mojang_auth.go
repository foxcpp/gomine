package gomine

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type AuthProvider interface {
	// Login initiates new session using specified credentials.
	Login(user, pass string) (AuthData, error)

	// Refresh updates login information in passed structure to keep user logged in.
	// Old information is no longer usable.
	Refresh(ad *AuthData) error

	// Validate checks whether information in AuthData structure is still usable.
	Validate(ad AuthData) (bool, error)

	// Invalidate terminates session.
	Invalidate(ad AuthData)
}

// MojangAuth implements AuthProvider using "Yggdrasil" authentication scheme.
type MojangAuth struct {
	// ClientID must be set to persistent value unique for this client.
	// Changing ClientID will make all old sessions unusable.
	ClientID string

	// URL that will be prepended ot each API endpoint path.
	// Defaults to https://authserver.mojang.com.
	AuthURL string
}

type mojangProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type mojangUserProp struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type mojangUser struct {
	ID    string           `json:"id"`
	Props []mojangUserProp `json:"properties"`
}

type mojangAuthError struct {
	ErrorCode string `json:"error"`
	ErrorMsg  string `json:"errorMessage"`
}

func (mae mojangAuthError) Error() string {
	return "auth fail: " + mae.ErrorMsg
}

type mojangAuthResp struct {
	AccessToken       string          `json:"accessToken"`
	ClientToken       string          `json:"clientToken"`
	AvailableProfiles []mojangProfile `json:"availableProfiles"`
	SelectedProfile   mojangProfile   `json:"sleectedProfile"`
	User              mojangUser      `json:"user"`
	mojangAuthError
}

func (ma MojangAuth) Login(user, pass string) (AuthData, error) {
	if ma.AuthURL == "" {
		ma.AuthURL = "https://authserver.mojang.com"
	}

	req := map[string]interface{}{
		"agent": map[string]interface{}{
			"name": "Minecraft",
			"version": 1,
		},
		"username": user,
		"password": pass,
		"clientToken": ma.ClientID,
		"requestUser": false,
	}

	blob, err := json.Marshal(req)
	if err != nil {
		return AuthData{}, errors.Wrap(err, "failed to encode auth request")
	}

	resp, err := http.Post(ma.AuthURL+"/authenticate", "application/json", bytes.NewReader(blob))
	if err != nil {
		return AuthData{}, errors.Wrap(err, "failed to send auth request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return AuthData{}, errors.New("auth request rejected: HTTP " + resp.Status)
	}

	respJson := mojangAuthResp{}
	if err := json.NewDecoder(resp.Body).Decode(&respJson); err != nil {
		return AuthData{}, errors.Wrap(err, "failed to parse auth response")
	}

	if respJson.ErrorCode != "" {
		return AuthData{}, respJson.mojangAuthError
	}

	if len(respJson.AvailableProfiles) == 0 {
		return AuthData{
			UserType: "mojang",
			PlayerName: user,
			UUID: "DEMO-USER",
			Token: "DEMO-USER",
		}, errors.New("not premium")
	}

	return AuthData{
		UserType: "mojang",
		PlayerName: respJson.SelectedProfile.Name,
		UUID: respJson.SelectedProfile.ID,
		Token: respJson.AccessToken,
	}, nil
}

func (ma MojangAuth) Refresh(ad *AuthData) error {
	if ma.AuthURL == "" {
		ma.AuthURL = "https://authserver.mojang.com"
	}

	req := map[string]interface{}{
		"accessToken": ad.Token,
		"clientToken": ma.ClientID,
		"selectedProfile": map[string]interface{}{
			"id": ad.UUID,
			"name": ad.PlayerName,
		},
		"requestUser": false,
	}

	blob, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to encode refresh request")
	}

	resp, err := http.Post(ma.AuthURL+"/refresh", "application/json", bytes.NewReader(blob))
	if err != nil {
		return errors.Wrap(err, "failed to send refresh request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("refresh request rejected: HTTP " + resp.Status)
	}

	respJson := mojangAuthResp{}
	if err := json.NewDecoder(resp.Body).Decode(&respJson); err != nil {
		return errors.Wrap(err, "failed to parse refresh response")
	}

	if respJson.ErrorCode != "" {
		return respJson.mojangAuthError
	}

	ad.PlayerName = respJson.SelectedProfile.Name
	ad.UUID = respJson.SelectedProfile.ID
	ad.Token = respJson.AccessToken
	return nil
}

func (ma MojangAuth) Validate(ad AuthData) (bool, error) {
	if ma.AuthURL == "" {
		ma.AuthURL = "https://authserver.mojang.com"
	}

	req := map[string]interface{}{
		"accessToken": ad.Token,
		"clientToken": ma.ClientID,
	}

	blob, err := json.Marshal(req)
	if err != nil {
		return false, errors.Wrap(err, "failed to encode validate request")
	}

	resp, err := http.Post(ma.AuthURL+"/validate", "application/json", bytes.NewReader(blob))
	if err != nil {
		return false, errors.Wrap(err, "failed to send validate request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return true, nil
	}
	if resp.StatusCode == 403 {
		return false, nil
	}

	respErr := mojangAuthError{}
	if err := json.NewDecoder(resp.Body).Decode(&respErr); err != nil {
		// we ignore parse errors and just return only HTTP code
		return false, errors.Errorf("HTTP %s (JSON body parse failed: %s)", resp.Status, err.Error())
	}
	return false, respErr
}

func (ma MojangAuth) Invalidate(ad AuthData) error {
	if ma.AuthURL == "" {
		ma.AuthURL = "https://authserver.mojang.com"
	}

	req := map[string]interface{}{
		"accessToken": ad.Token,
		"clientToken": ma.ClientID,
	}

	blob, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to encode invalidate request")
	}

	resp, err := http.Post(ma.AuthURL+"/invalidate", "application/json", bytes.NewReader(blob))
	if err != nil {
		return errors.Wrap(err, "failed to send invalidate request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respErr := mojangAuthError{}
		if err := json.NewDecoder(resp.Body).Decode(&respErr); err != nil {
			// we ignore parse errors and just return only HTTP code
			return errors.Errorf("HTTP %s (JSON body parse failed: %s)", resp.Status, err.Error())
		}
		return respErr
	}

	return nil
}