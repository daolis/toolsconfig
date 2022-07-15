package toolsconfig

import (
	"fmt"
	"os"
	"strings"
)

var _ Configuration = new(ToolConfiguration)

type ToolConfiguration struct {
	config             *Config
	servers            map[string]*ServerCredential
	azureSubscriptions map[string]*AzureSubscriptionCredential
	generics           map[string]*GenericCredential
	configReader       func() (*Configuration, error)
}

type Config struct {
	DefaultAzureSubscription string                          `yaml:"defaultAzureSubscription,omitempty"`
	Servers                  []ServerCredential              `yaml:"servers"`
	AzureSubscriptions       []AzureSubscriptionCredential   `yaml:"azureSubscriptions"`
	Generic                  []GenericCredential             `yaml:"generics"`
	Favourites               map[string]map[string]Favourite `yaml:"favourites"`
}

type ServerCredential struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type AzureSubscriptionCredential struct {
	Name           string `yaml:"name"`
	SubscriptionID string `yaml:"subscriptionID"`
	TenantID       string `yaml:"tenantID"`
	ClientID       string `yaml:"clientID"`
	ClientSecret   string `yaml:"clientSecret"`
}

type GenericCredential struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Favourite struct {
	Name string   `yaml:"name"`
	Args []string `yaml:"args,flow"`
}

func (s Favourite) String() string {
	return fmt.Sprintf("%s - '%s'", s.Name, strings.Join(s.Args, " "))
}

func (c *Config) merge(required *Config) bool {
	var dirty bool
	for _, server := range required.Servers {
		_, _, err := c.serverCredential(server.URL)
		if err != nil {
			c.Servers = append(c.Servers, server)
			dirty = true
		}
	}
	for _, subscription := range required.AzureSubscriptions {
		_, _, err := c.azureSubscriptionCredential(subscription.SubscriptionID)
		if err != nil {
			c.AzureSubscriptions = append(c.AzureSubscriptions, subscription)
			dirty = true
		}
	}
	for _, generic := range required.Generic {
		_, _, err := c.genericCredential(generic.Key)
		if err != nil {
			c.Generic = append(c.Generic, generic)
			dirty = true
		}
	}
	return dirty
}

func (c Config) serverCredential(url string) (*ServerCredential, *int, error) {
	for index, server := range c.Servers {
		if server.URL == url {
			return &server, &index, nil
		}
	}
	return nil, nil, wrapErr(errNotFound, "server '"+url+"'")
}

func (c Config) azureSubscriptionCredential(nameOrID string) (*AzureSubscriptionCredential, *int, error) {
	for index, subscription := range c.AzureSubscriptions {
		if subscription.Name == nameOrID || subscription.SubscriptionID == nameOrID {
			return &subscription, &index, nil
		}
	}
	return nil, nil, wrapErr(errNotFound, "subscription '"+nameOrID+"'")
}

func (c Config) genericCredential(key string) (*GenericCredential, *int, error) {
	for index, generic := range c.Generic {
		if generic.Key == key {
			return &generic, &index, nil
		}
	}
	return nil, nil, wrapErr(errNotFound, "generic '"+key+"'")
}

func (c ServerCredential) valid() bool {
	return c.Username != "" && c.Password != ""
}

func (c ServerCredential) FromEnv(url string) *ServerCredential {
	username := os.Getenv(toEnvironmentKey(url, "username"))
	password := os.Getenv(toEnvironmentKey(url, "password"))
	result := ServerCredential{
		URL:      url,
		Username: username,
		Password: password,
	}
	if result.valid() {
		return &result
	}
	return nil
}

func (c AzureSubscriptionCredential) valid() bool {
	return c.SubscriptionID != "" && c.TenantID != "" && c.ClientID != "" && c.ClientSecret != ""
}

func (c AzureSubscriptionCredential) FromEnv(name string) *AzureSubscriptionCredential {
	subscriptionId := os.Getenv(toEnvironmentKey(name, "subscriptionId"))
	tenantId := os.Getenv(toEnvironmentKey(name, "tenantId"))
	clientId := os.Getenv(toEnvironmentKey(name, "clientId"))
	clientSecret := os.Getenv(toEnvironmentKey(name, "clientSecret"))
	result := AzureSubscriptionCredential{
		Name:           name,
		SubscriptionID: subscriptionId,
		TenantID:       tenantId,
		ClientID:       clientId,
		ClientSecret:   clientSecret,
	}
	if result.valid() {
		return &result
	}
	return nil
}

func (c GenericCredential) valid() bool {
	return c.Value != ""
}
func (c GenericCredential) FromEnv(key string) *GenericCredential {
	value := os.Getenv(toEnvironmentKey(key, "value"))
	result := GenericCredential{
		Key:   key,
		Value: value,
	}
	if result.valid() {
		return &result
	}
	return nil
}
