package client

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

const credentialsFile = "credentials.json"
const configFile = "config.json"

var cfgDir = ""

type (
	ProviderCredentials struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
	}

	Credentials struct {
		Credentials     map[string]*ProviderCredentials `json:"credentials"`
		DefaultProvider string                          `json:"default_provider"`
	}

	Configuration struct {
		BaseURL string
	}
)

const defaultServerUrl = "https://timelapse.wangpedersen.com"

var configurationOverrides Configuration

func init() {
	rootCmd.PersistentFlags().StringVar(&configurationOverrides.BaseURL, "server-url", "", "Server URL, default "+defaultServerUrl)
}

func GetCredentials() (*Credentials, error) {
	var c Credentials
	err := internalLoadConfig(credentialsFile, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func GetConfiguration() (*Configuration, error) {
	var c Configuration
	err := internalLoadConfig(configFile, &c)
	if err != nil {
		return nil, err
	}

	if configurationOverrides.BaseURL != "" {
		c.BaseURL = configurationOverrides.BaseURL
	}
	if c.BaseURL == "" {
		c.BaseURL = defaultServerUrl
	}
	return &c, nil
}

func (credentials *Credentials) Store() error {
	return internalStoreConfig(credentialsFile, credentials, 0600)
}

func (credentials *Credentials) SetProviderCredentialsFromToken(providerName string, token *tokenResponse) {
	var creds ProviderCredentials
	creds.AccessToken = token.AccessToken
	creds.RefreshToken = token.RefreshToken
	creds.IDToken = token.IDToken

	if credentials.Credentials == nil {
		credentials.Credentials = make(map[string]*ProviderCredentials)
	}
	credentials.Credentials[providerName] = &creds
	if credentials.DefaultProvider == "" {
		credentials.DefaultProvider = providerName
	}
}

func (config *Configuration) Store() error {
	return internalStoreConfig(configFile, config, 0644)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgDir, "config", "", "config directory location (default is "+configDir()+"/)")
}

func ensureConfigDirExists() error {
	return os.MkdirAll(configDir(), 0755)
}

func internalStoreConfig(filename string, config interface{}, perm os.FileMode) error {
	err := ensureConfigDirExists()
	if err != nil {
		return err
	}

	filePath := filepath.Join(configDir(), filename)

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, data, perm)
	if err != nil {
		return err
	}

	return nil

}

func internalLoadConfig(filename string, dest interface{}) error {
	filePath := filepath.Join(configDir(), filename)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return err
	}

	return nil
}

func loadCredentials() (*Credentials, error) {
	var c Credentials
	err := internalLoadConfig(credentialsFile, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func loadConfig() (*Configuration, error) {
	var c Configuration
	err := internalLoadConfig(configFile, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
