package toolsconfig

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	serverURL01        = "testserver.io"
	serverURL02        = "ser.test02.com"
	subscriptionName01 = "testSubscription01"
	generic01          = "generic"
)

func TestNewConfiguration(t *testing.T) {
	var savedConfig *Config
	type args struct {
		options []ConfigOption
	}
	tests := []struct {
		name     string
		args     args
		prepare  func()
		validate func(t *testing.T, c Configuration, err error)
		cleanup  func()
	}{
		{
			name: "MissingValues",
			args: args{options: []ConfigOption{
				RequiredServer(serverURL01),
				RequiredServer(serverURL02),
				RequiredSubscription(subscriptionName01),
				RequiredGeneric(generic01),
			}},
			prepare: func() {
				readConfiguration = func() *Config {
					return &Config{}
				}
				saveConfiguration = func(config *Config) error {
					return nil
				}
			},
			validate: func(t *testing.T, c Configuration, err error) {
				require.Nil(t, c)
				require.NotNil(t, err)
				var configError *ConfigError
				errors.As(err, &configError)
				require.Error(t, err, "missing values")
				require.Equal(t, len(configError.Missing), 4)
			},
		},
		{
			name: "ExistingValue",
			args: args{options: []ConfigOption{
				RequiredSubscription(subscriptionName01),
				RequiredServer(serverURL01),
				RequiredGeneric(generic01),
			}},
			prepare: func() {
				readConfiguration = func() *Config {
					return &Config{
						Servers: []ServerCredential{
							{URL: serverURL01, Username: "testusername", Password: "testpassword"},
						},
						AzureSubscriptions: []AzureSubscriptionCredential{
							{Name: subscriptionName01, SubscriptionID: "subscription-id", TenantID: "tenant-id", ClientID: "client-id", ClientSecret: "client-secret"},
						},
						Generic: []GenericCredential{
							{Key: generic01, Value: "genericValue"},
						},
					}
				}
				saveConfiguration = func(config *Config) error {
					savedConfig = config
					return nil
				}
			},
			validate: func(t *testing.T, c Configuration, err error) {
				require.Nil(t, err)
				require.NotNil(t, c)
				serverCred, err := c.GetServerCredentials(serverURL01)
				require.Nil(t, err)
				require.Equal(t, "testusername", serverCred.Username)
				require.Equal(t, "testpassword", serverCred.Password)
				subscriptionCred, err := c.GetAzureSubscriptionCredentials(subscriptionName01)
				require.Nil(t, err)
				require.Equal(t, "subscription-id", subscriptionCred.SubscriptionID)
				require.Equal(t, "tenant-id", subscriptionCred.TenantID)
				require.Equal(t, "client-id", subscriptionCred.ClientID)
				require.Equal(t, "client-secret", subscriptionCred.ClientSecret)
				subscriptionCred, err = c.GetAzureSubscriptionCredentials("subscription-id")
				require.Nil(t, err)
				require.Equal(t, subscriptionName01, subscriptionCred.Name)
				genericCred := c.GetGeneric(generic01)
				require.Equal(t, "genericValue", genericCred)

				require.Nil(t, savedConfig)
			},
		},
		{
			name: "AdditionalValueInFile",
			args: args{options: []ConfigOption{
				RequiredSubscription(subscriptionName01),
				RequiredServer(serverURL01),
				RequiredServer("new-server"),
				RequiredGeneric(generic01),
			}},
			prepare: func() {
				readConfiguration = func() *Config {
					return &Config{
						Servers: []ServerCredential{
							{URL: serverURL01, Username: "testusername", Password: "testpassword"},
						},
						AzureSubscriptions: []AzureSubscriptionCredential{
							{Name: subscriptionName01, SubscriptionID: "subscription-id", TenantID: "tenant-id", ClientID: "client-id", ClientSecret: "client-secret"},
						},
					}
				}
				saveConfiguration = func(config *Config) error {
					savedConfig = config
					return nil
				}
			},
			validate: func(t *testing.T, c Configuration, err error) {
				require.Nil(t, c)
				require.NotNil(t, err)
				var configError *ConfigError
				errors.As(err, &configError)
				require.Error(t, err, "missing values")
				require.Equal(t, len(configError.Missing), 2)

				require.NotNil(t, savedConfig)
				require.Equal(t, 2, len(savedConfig.Servers))
				require.Equal(t, 1, len(savedConfig.AzureSubscriptions))
				require.Equal(t, 1, len(savedConfig.Generic))
			},
		},
		{
			name: "CreateFile",
			args: args{options: []ConfigOption{
				RequiredSubscription(subscriptionName01),
				RequiredServer(serverURL01),
				RequiredServer("new-server"),
				RequiredGeneric(generic01),
			}},
			prepare: func() {
				readConfiguration = func() *Config {
					return &Config{
						Servers: []ServerCredential{
							{URL: serverURL01, Username: "testusername", Password: "testpassword"},
						},
						AzureSubscriptions: []AzureSubscriptionCredential{
							{Name: subscriptionName01, SubscriptionID: "subscription-id", TenantID: "tenant-id", ClientID: "client-id", ClientSecret: "client-secret"},
						},
					}
				}
				saveConfiguration = defaultSaveConfiguration
			},
			validate: func(t *testing.T, c Configuration, err error) {
				require.Nil(t, c)
				require.NotNil(t, err)
				var configError *ConfigError
				errors.As(err, &configError)
				require.Error(t, err, "missing values")
				require.Equal(t, len(configError.Missing), 2)

				savedConfig := defaultReadConfiguration()
				require.NotNil(t, savedConfig)
				require.Equal(t, 2, len(savedConfig.Servers))
				require.Equal(t, 1, len(savedConfig.AzureSubscriptions))
				require.Equal(t, 1, len(savedConfig.Generic))
			},
			cleanup: func() {
				//_ = os.Remove("unittestconfig.yaml")
			},
		},
	}
	ConfigFileLocation(".", "unittestconfig.yaml")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare()
			}
			got, err := NewToolConfiguration(tt.args.options...)
			if tt.cleanup != nil {
				defer tt.cleanup()
			}
			tt.validate(t, got, err)
		})
	}
}

func TestGettingConfigurationEnvOnly(t *testing.T) {
	var saveCalled bool
	readConfiguration = func() *Config {
		return &Config{}
	}
	saveConfiguration = func(config *Config) error {
		saveCalled = true
		return nil
	}
	require.NoError(t, os.Setenv(toEnvironmentKey(serverURL01, "username"), "envUsername"))
	require.NoError(t, os.Setenv(toEnvironmentKey(serverURL01, "password"), "envPassword"))

	configuration, err := NewToolConfiguration(RequiredServer(serverURL01), UpdateConfig(true))
	require.NoError(t, err)
	serverCredentials, err := configuration.GetServerCredentials(serverURL01)
	require.NoError(t, err)
	require.Equal(t, serverURL01, serverCredentials.URL)
	require.Equal(t, "envUsername", serverCredentials.Username)
	require.Equal(t, "envPassword", serverCredentials.Password)
	require.False(t, saveCalled)
	require.NoError(t, os.Unsetenv(toEnvironmentKey(serverURL01, "username")))
	require.NoError(t, os.Unsetenv(toEnvironmentKey(serverURL01, "password")))
}

func TestGettingConfiguration(t *testing.T) {
	var savedConfig = &Config{
		DefaultAzureSubscription: "",
		Servers: []ServerCredential{
			{URL: serverURL01, Username: "testusername", Password: "testpassword"},
		},
		AzureSubscriptions: []AzureSubscriptionCredential{
			{Name: subscriptionName01, SubscriptionID: "subscription-id", TenantID: "tenant-id", ClientID: "client-id", ClientSecret: "client-secret"},
		},
		Generic: []GenericCredential{
			{Key: generic01, Value: "genericValue"},
		},
	}

	readConfiguration = func() *Config {
		return savedConfig
	}
	saveConfiguration = func(config *Config) error {
		savedConfig = config
		return nil
	}

	configuration, err := NewToolConfiguration()
	t.Run("GetServerConfig", func(t *testing.T) {
		require.NoError(t, err)
		serverCredentials, err := configuration.GetServerCredentials(serverURL01)
		require.NoError(t, err)
		require.Equal(t, "testusername", serverCredentials.Username)
		require.Equal(t, "testpassword", serverCredentials.Password)
	})

	t.Run("GetSubscriptionCredentials", func(t *testing.T) {
		subscriptionCredentials, err := configuration.GetAzureSubscriptionCredentials(subscriptionName01)
		require.NoError(t, err)
		require.Equal(t, "subscription-id", subscriptionCredentials.SubscriptionID)
		require.Equal(t, "tenant-id", subscriptionCredentials.TenantID)
		require.Equal(t, "client-id", subscriptionCredentials.ClientID)
		require.Equal(t, "client-secret", subscriptionCredentials.ClientSecret)
	})

	t.Run("GetGenericCredentials", func(t *testing.T) {
		genericCredentials, err := configuration.GetGenericCredentials(generic01)
		require.NoError(t, err)
		require.Equal(t, generic01, genericCredentials.Key)
		require.Equal(t, "genericValue", genericCredentials.Value)
	})

	t.Run("GetSubscriptionDefaultCredentialsWithoutDefaultValueSet", func(t *testing.T) {
		_, err = configuration.GetAzureSubscriptionCredentials("")
		require.Error(t, err)
	})

	t.Run("SetDefaultAndGetSubscriptionDefaultCredentials", func(t *testing.T) {
		err = configuration.SetDefaultSubscription(subscriptionName01)
		require.NoError(t, err)

		subscriptionCredentials, err := configuration.GetAzureSubscriptionCredentials("")
		require.NoError(t, err)
		require.Equal(t, "subscription-id", subscriptionCredentials.SubscriptionID)
		require.Equal(t, "tenant-id", subscriptionCredentials.TenantID)
		require.Equal(t, "client-id", subscriptionCredentials.ClientID)
		require.Equal(t, "client-secret", subscriptionCredentials.ClientSecret)
	})

	t.Run("GetServerConfigFromEnv", func(t *testing.T) {
		require.NoError(t, os.Setenv(toEnvironmentKey(serverURL01, "username"), "envUsername"))
		require.NoError(t, os.Setenv(toEnvironmentKey(serverURL01, "password"), "envPassword"))
		require.NoError(t, err)
		serverCredentials, err := configuration.GetServerCredentials(serverURL01)
		require.NoError(t, err)
		require.Equal(t, serverURL01, serverCredentials.URL)
		require.Equal(t, "envUsername", serverCredentials.Username)
		require.Equal(t, "envPassword", serverCredentials.Password)
	})

	t.Run("GetSubscriptionCredentialsFromEnv", func(t *testing.T) {
		require.NoError(t, os.Setenv(toEnvironmentKey(subscriptionName01, "subscriptionId"), "envSubId"))
		require.NoError(t, os.Setenv(toEnvironmentKey(subscriptionName01, "tenantID"), "envTenantId"))
		require.NoError(t, os.Setenv(toEnvironmentKey(subscriptionName01, "clientID"), "envClientId"))
		require.NoError(t, os.Setenv(toEnvironmentKey(subscriptionName01, "clientSecret"), "envClientSecret"))
		subscriptionCredentials, err := configuration.GetAzureSubscriptionCredentials(subscriptionName01)
		require.NoError(t, err)
		require.Equal(t, subscriptionName01, subscriptionCredentials.Name)
		require.Equal(t, "envSubId", subscriptionCredentials.SubscriptionID)
		require.Equal(t, "envTenantId", subscriptionCredentials.TenantID)
		require.Equal(t, "envClientId", subscriptionCredentials.ClientID)
		require.Equal(t, "envClientSecret", subscriptionCredentials.ClientSecret)
	})

	t.Run("GetGenericCredentialsFromEnv", func(t *testing.T) {
		require.NoError(t, os.Setenv("GENERIC_VALUE", "envGenValue"))
		genericCredentials, err := configuration.GetGenericCredentials(generic01)
		require.NoError(t, err)
		require.Equal(t, generic01, genericCredentials.Key)
		require.Equal(t, "envGenValue", genericCredentials.Value)
	})

}

func TestSetConfiguration_Update(t *testing.T) {
	var savedConfig = &Config{
		DefaultAzureSubscription: "",
		AzureSubscriptions: []AzureSubscriptionCredential{
			{Name: subscriptionName01, SubscriptionID: "subscription-id", TenantID: "tenant-id", ClientID: "client-id", ClientSecret: "client-secret"},
		},
		Servers: []ServerCredential{
			{URL: serverURL01, Username: "testusername", Password: "testpassword"},
		},
		Generic: []GenericCredential{
			{Key: generic01, Value: "genericValue"},
		},
	}

	readConfiguration = func() *Config {
		return savedConfig
	}
	saveConfiguration = func(config *Config) error {
		savedConfig = config
		return nil
	}

	configuration, err := NewToolConfiguration()
	t.Run("SetExistingSubscriptionCredentials", func(t *testing.T) {
		require.NoError(t, err)
		err := configuration.SetAzureSubscriptionCredentials(AzureSubscriptionCredential{
			Name:           subscriptionName01,
			SubscriptionID: "new-subscription-id",
			TenantID:       "new-tenant-id",
			ClientID:       "new-client-id",
			ClientSecret:   "new-client-secret",
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(savedConfig.AzureSubscriptions))
		require.Equal(t, 1, len(savedConfig.Servers))
		require.Equal(t, 1, len(savedConfig.Generic))
		require.Equal(t, subscriptionName01, savedConfig.AzureSubscriptions[0].Name)
		require.Equal(t, "new-subscription-id", savedConfig.AzureSubscriptions[0].SubscriptionID)
		require.Equal(t, "new-tenant-id", savedConfig.AzureSubscriptions[0].TenantID)
		require.Equal(t, "new-client-id", savedConfig.AzureSubscriptions[0].ClientID)
		require.Equal(t, "new-client-secret", savedConfig.AzureSubscriptions[0].ClientSecret)
	})

	t.Run("SetExistingServerCredentials", func(t *testing.T) {
		err := configuration.SetServerCredentials(ServerCredential{
			URL:      serverURL01,
			Username: "newTestusername",
			Password: "newTestpassword",
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(savedConfig.AzureSubscriptions))
		require.Equal(t, 1, len(savedConfig.Servers))
		require.Equal(t, 1, len(savedConfig.Generic))
		require.Equal(t, serverURL01, savedConfig.Servers[0].URL)
		require.Equal(t, "newTestusername", savedConfig.Servers[0].Username)
		require.Equal(t, "newTestpassword", savedConfig.Servers[0].Password)
	})

	t.Run("SetExistingGenericCredentials", func(t *testing.T) {
		err := configuration.SetGenericCredentials(GenericCredential{
			Key:   generic01,
			Value: "newGenericValue",
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(savedConfig.AzureSubscriptions))
		require.Equal(t, 1, len(savedConfig.Servers))
		require.Equal(t, 1, len(savedConfig.Generic))
		require.Equal(t, generic01, savedConfig.Generic[0].Key)
		require.Equal(t, "newGenericValue", savedConfig.Generic[0].Value)
	})

	t.Run("SetNotExisingDefaultSubscription", func(t *testing.T) {
		err = configuration.SetDefaultSubscription("lalala")
		require.Error(t, err)
		require.Equal(t, "", savedConfig.DefaultAzureSubscription)
	})

	t.Run("SetExisingDefaultSubscription", func(t *testing.T) {
		err = configuration.SetDefaultSubscription(subscriptionName01)
		require.NoError(t, err)
		require.Equal(t, subscriptionName01, savedConfig.DefaultAzureSubscription)
	})
}

func TestFavourites(t *testing.T) {
	var savedConfig = &Config{
		DefaultAzureSubscription: "",
	}

	readConfiguration = func() *Config {
		return savedConfig
	}
	saveConfiguration = func(config *Config) error {
		savedConfig = config
		return nil
	}

	configuration, err := NewToolConfiguration()
	const toolName = "testtool"
	const firstFavName = "testFav1"
	const secondFavName = "testFav2"
	t.Run("CheckEmptyFavouritesTools", func(t *testing.T) {
		require.NoError(t, err)
		tools := configuration.GetFavourites(toolName)
		require.Equal(t, 0, len(tools))
	})

	t.Run("AddFavourite", func(t *testing.T) {
		err := configuration.SaveFavourite(toolName, firstFavName, []string{"arg1", "arg2", "arg3", "arg4"})
		require.NoError(t, err)
		require.Nil(t, savedConfig.AzureSubscriptions)
		require.Nil(t, savedConfig.Servers)
		require.Nil(t, savedConfig.Generic)
		require.Equal(t, 1, len(savedConfig.Favourites))
		t.Run("GetFavourite-Directly", func(t *testing.T) {
			toolFavs := savedConfig.Favourites[toolName]
			require.NotNil(t, toolFavs)
			firstFav := toolFavs[firstFavName]
			require.NotNil(t, firstFav)
			require.Equal(t, firstFavName, firstFav.Name)
			require.Equal(t, []string{"arg1", "arg2", "arg3", "arg4"}, firstFav.Args)
		})
		t.Run("GetFavourite-Function", func(t *testing.T) {
			favourite, err := configuration.GetFavourite(toolName, firstFavName)
			require.NoError(t, err)
			require.NotNil(t, favourite)
			require.Equal(t, firstFavName, favourite.Name)
			require.Equal(t, []string{"arg1", "arg2", "arg3", "arg4"}, favourite.Args)
		})
		t.Run("AddFavourite-OneMore", func(t *testing.T) {
			err := configuration.SaveFavourite(toolName, secondFavName, []string{"x1", "x2", "x3", "x4", "x5"})
			require.NoError(t, err)
			allFavs := configuration.GetFavourites(toolName)
			require.NotNil(t, allFavs)
			require.Equal(t, 2, len(allFavs))
			favourite, err := configuration.GetFavourite(toolName, secondFavName)
			require.NoError(t, err)
			require.NotNil(t, favourite)
			require.Equal(t, secondFavName, favourite.Name)
			require.Equal(t, []string{"x1", "x2", "x3", "x4", "x5"}, favourite.Args)
		})
	})
}
