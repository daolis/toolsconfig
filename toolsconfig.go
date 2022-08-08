package toolsconfig

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	ConfigFormat          = "yaml"
	ConfigDirectory       = ".toolsconfig"
	ConfigFilename        = "config.yaml"
	ConfigFilePermissions = 0600
)

type Configuration interface {
	// SetAzureSubscriptionCredentials set the azure subscription credentials.
	SetAzureSubscriptionCredentials(entry AzureSubscriptionCredential) error
	// GetAzureSubscriptionCredentials get the azure subscription credentials.
	GetAzureSubscriptionCredentials(nameOrID string) (*AzureSubscriptionCredential, error)
	SetServerCredentials(entry ServerCredential) error
	SetGenericCredential(entry GenericCredential) error
	GetServerCredentials(url string) (*ServerCredential, error)
	GetGenericCredentials(key string) (*GenericCredential, error)
	GetGeneric(key string) string
	SetDefaultSubscription(subscriptionName string) error
	SaveFavourite(tool, name string, args []string) error
	GetFavourite(tool, name string) (*Favourite, error)
	GetFavourites(tool string) []Favourite
	RemoveFavourite(tool, name string) error
}

// NewToolConfiguration creates a new configuration object.
// If the option WithConfigDir and WithConfigFile are not provided,
// the configuration is loaded from the default location (~/.toolsconfig/config/yaml).
// Use the options
// * RequiredServer(..)
// * RequiredAzureSubscription(..)
// * RequiredGeneric(..)
// to specify which credentials are required. If the credentials are not available in the configuration,
// an error is returned immediately.
func NewToolConfiguration(options ...ConfigOption) (Configuration, error) {
	opts := ConfigOptions{
		updateConfig:    true,
		configDirectory: ConfigDirectory,
		configFile:      ConfigFilename,
	}
	for _, option := range options {
		option(&opts)
	}

	c := &ToolConfiguration{
		servers:            map[string]*ServerCredential{},
		azureSubscriptions: map[string]*AzureSubscriptionCredential{},
		generics:           map[string]*GenericCredential{},
	}

	file, err := configFile(opts.configDirectory, opts.configFile)
	if err != nil {
		return nil, wrapErr(err)
	}
	viper.SetConfigType(ConfigFormat)
	viper.SetConfigFile(*file)
	viper.SetConfigPermissions(ConfigFilePermissions)

	c.config = readConfiguration()
	err = verifyRequiredValues(c, opts)
	if err != nil {
		dirty := c.config.merge(opts.requiredConfig())
		if dirty && opts.updateConfig {
			err := saveConfiguration(c.config)
			if err != nil {
				return nil, wrapErr(err)
			}
		}
		return nil, err
	}
	return c, err
}

func verifyRequiredValues(c Configuration, opts ConfigOptions) error {
	var missingCredentials []string
	for _, serverURL := range opts.requiredServers {
		serverCredential, err := c.GetServerCredentials(serverURL)
		if err != nil || !serverCredential.valid() {
			missingCredentials = append(missingCredentials, fmt.Sprintf("Server: %s", serverURL))
		}
	}
	for _, idOrName := range opts.requiredAzureSubscriptions {
		subscriptionCredential, err := c.GetAzureSubscriptionCredentials(idOrName)
		if err != nil || !subscriptionCredential.valid() {
			missingCredentials = append(missingCredentials, fmt.Sprintf("AzureSubscriptionCredential: %s", idOrName))
		}
	}
	for _, key := range opts.requiredGenerics {
		generic, err := c.GetGenericCredentials(key)
		if err != nil || !generic.valid() {
			missingCredentials = append(missingCredentials, fmt.Sprintf("GenericCredential: %s", key))
		}
	}
	if len(missingCredentials) > 0 {
		return wrapErr(fmt.Errorf("missing entries"), missingCredentials...)
	}
	return nil
}

func (c *ToolConfiguration) SetAzureSubscriptionCredentials(entry AzureSubscriptionCredential) error {
	if entry.Name == "" {
		return fmt.Errorf("subscription name missing")
	}
	_, index, err := c.config.azureSubscriptionCredential(entry.Name)
	if err != nil {
		c.config.AzureSubscriptions = append(c.config.AzureSubscriptions, entry)
	} else {
		c.config.AzureSubscriptions[*index].SubscriptionID = entry.SubscriptionID
		c.config.AzureSubscriptions[*index].TenantID = entry.TenantID
		c.config.AzureSubscriptions[*index].ClientID = entry.ClientID
		c.config.AzureSubscriptions[*index].ClientSecret = entry.ClientSecret
	}
	return saveConfiguration(c.config)
}

func (c *ToolConfiguration) SetServerCredentials(entry ServerCredential) error {
	if entry.URL == "" {
		return fmt.Errorf("server url missing")
	}
	_, index, err := c.config.serverCredential(entry.URL)
	if err != nil {
		c.config.Servers = append(c.config.Servers, entry)
	} else {
		c.config.Servers[*index].URL = entry.URL
		c.config.Servers[*index].Username = entry.Username
		c.config.Servers[*index].Password = entry.Password
	}
	return saveConfiguration(c.config)
}

func (c *ToolConfiguration) SetGenericCredential(entry GenericCredential) error {
	if entry.Key == "" {
		return fmt.Errorf("generic credential key missing")
	}
	_, index, err := c.config.genericCredential(entry.Key)
	if err != nil {
		c.config.Generic = append(c.config.Generic, entry)
	} else {
		c.config.Generic[*index].Key = entry.Key
		c.config.Generic[*index].Value = entry.Value
	}
	return saveConfiguration(c.config)
}

// GetServerCredentials find the credentials for the given url. Returns errNotFound if not found.
func (c *ToolConfiguration) GetServerCredentials(url string) (*ServerCredential, error) {
	if fromEnv := (ServerCredential{}.FromEnv(url)); fromEnv != nil {
		return fromEnv, nil
	}
	if serverCred, ok := c.servers[url]; ok {
		return serverCred, nil
	}
	credential, _, err := c.config.serverCredential(url)
	if err != nil {
		return nil, err
	}
	c.servers[url] = credential
	return credential, nil
}

// GetAzureSubscriptionCredentials find the credentials for the given name or subscription id. Returns errNotFound if not found.
func (c *ToolConfiguration) GetAzureSubscriptionCredentials(nameOrID string) (*AzureSubscriptionCredential, error) {
	if fromEnv := (AzureSubscriptionCredential{}.FromEnv(nameOrID)); fromEnv != nil {
		return fromEnv, nil
	}
	searchNameOrId := nameOrID
	if nameOrID == "" && c.config.DefaultAzureSubscription != "" {
		searchNameOrId = c.config.DefaultAzureSubscription
	}
	if azureCred, ok := c.azureSubscriptions[searchNameOrId]; ok {
		return azureCred, nil
	}
	credential, _, err := c.config.azureSubscriptionCredential(searchNameOrId)
	if err != nil {
		return nil, err
	}
	c.azureSubscriptions[searchNameOrId] = credential
	return credential, nil
}

// GetGenericCredentials find the credentials for the given key. Returns errNotFound if not found.
func (c *ToolConfiguration) GetGenericCredentials(key string) (*GenericCredential, error) {
	if fromEnv := (GenericCredential{}.FromEnv(key)); fromEnv != nil {
		return fromEnv, nil
	}
	if genericCred, ok := c.generics[key]; ok {
		return genericCred, nil
	}
	credential, _, err := c.config.genericCredential(key)
	if err != nil {
		return nil, err
	}
	c.generics[key] = credential
	return credential, nil
}

// GetGeneric is a simple call to get only the value of a generic key. Empty string if not exists.
func (c *ToolConfiguration) GetGeneric(key string) string {
	credentials, err := c.GetGenericCredentials(key)
	if err != nil {
		return ""
	}
	return credentials.Value
}

// SetDefaultSubscription updates the default subscription value in the configuration. GetAzureSubscriptionCredentials returns the
// subscription credentials with this name or id if the given identifier is empty.
func (c *ToolConfiguration) SetDefaultSubscription(subscriptionName string) error {
	_, _, err := c.config.azureSubscriptionCredential(subscriptionName)
	if err != nil {
		return wrapErr(err, "subscription does not exist")
	}
	c.config.DefaultAzureSubscription = subscriptionName
	err = saveConfiguration(c.config)
	if err != nil {
		return wrapErr(err)
	}
	return nil
}

func (c *ToolConfiguration) SaveFavourite(tool, name string, args []string) error {
	if c.config.Favourites == nil {
		c.config.Favourites = make(map[string]map[string]Favourite, 1)
	}
	if c.config.Favourites[tool] == nil {
		c.config.Favourites[tool] = make(map[string]Favourite, 1)
	}
	c.config.Favourites[tool][name] = Favourite{Name: name, Args: args}
	err := saveConfiguration(c.config)
	if err != nil {
		return wrapErr(err)
	}
	return nil
}

func (c *ToolConfiguration) GetFavourite(tool, name string) (*Favourite, error) {
	if tool, ok := c.config.Favourites[tool]; ok {
		if command, ok := tool[name]; ok {
			return &command, nil
		}
	}
	return nil, wrapErr(errNotFound)
}

func (c *ToolConfiguration) GetFavourites(tool string) []Favourite {
	if tool, ok := c.config.Favourites[tool]; ok {
		result := make([]Favourite, len(tool))
		idx := 0
		for _, favourite := range tool {
			result[idx] = favourite
			idx++
		}
		return result
	}
	return []Favourite{}
}

func (c *ToolConfiguration) RemoveFavourite(tool, name string) error {
	if c.config.Favourites == nil {
		return wrapErr(fmt.Errorf("no saved favourites"))
	}
	if c.config.Favourites[tool] == nil {
		return wrapErr(fmt.Errorf("no favourites exist for tool %s", tool))
	}
	delete(c.config.Favourites[tool], name)
	err := saveConfiguration(c.config)
	if err != nil {
		return wrapErr(err)
	}
	return nil
}
