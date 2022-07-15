package toolsconfig

type ConfigOption func(*ConfigOptions)

type ConfigOptions struct {
	requiredServers            []string
	requiredAzureSubscriptions []string
	requiredGenerics           []string
	updateConfig               bool
	configDirectory            string
	configFile                 string
}

func (c ConfigOptions) requiredConfig() *Config {
	config := Config{
		Servers:            make([]ServerCredential, len(c.requiredServers)),
		AzureSubscriptions: make([]AzureSubscriptionCredential, len(c.requiredAzureSubscriptions)),
		Generic:            make([]GenericCredential, len(c.requiredGenerics)),
	}
	for idx, server := range c.requiredServers {
		config.Servers[idx] = ServerCredential{
			URL: server,
		}
	}
	for idx, subscription := range c.requiredAzureSubscriptions {
		config.AzureSubscriptions[idx] = AzureSubscriptionCredential{
			Name:           subscription,
			SubscriptionID: subscription,
		}
	}
	for idx, generic := range c.requiredGenerics {
		config.Generic[idx] = GenericCredential{
			Key: generic,
		}
	}
	return &config
}

// RequiredServer add the server credential with given url as required
func RequiredServer(serverURLs string) ConfigOption {
	return func(c *ConfigOptions) {
		c.requiredServers = append(c.requiredServers, serverURLs)
	}
}

// RequiredSubscription add the subscription credential with given name or ID as required
func RequiredSubscription(nameOrIDs string) ConfigOption {
	return func(c *ConfigOptions) {
		c.requiredAzureSubscriptions = append(c.requiredAzureSubscriptions, nameOrIDs)
	}
}

// RequiredGeneric add the generic credential with given key as required
func RequiredGeneric(key string) ConfigOption {
	return func(c *ConfigOptions) {
		c.requiredGenerics = append(c.requiredGenerics, key)
	}
}

// UpdateConfig defines whether the config should be updated in file. Default is 'true'. If Enabled and an unknown credential is requested
// a new empty entry will be added to the configuration file.
func UpdateConfig(value bool) ConfigOption {
	return func(c *ConfigOptions) {
		c.updateConfig = value
	}
}
