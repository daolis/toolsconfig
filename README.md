# Toolsconfig

> **Documentation is work in progress!**

Single config file containing credentials and favourites for cli tools using cobra and viper.

Default configuration file: `~/.toolsconfig/config.yaml`.
On Linux systems the directory should have 700 permissions, and the file 600, otherwise an error will be thrown.

## The following kinds of credentials are now available.

- Server credentials per URL/Identifier (Similar to `~/.m2/settings.xml` or `~/.gradle/gradle.properties`) e.g. Docker registry, Artifactory, ...
- Azure Subscription Credentials
- Generic credentials (Simple Key/Value pair)

If a credential is required, and it does not exist in the config file a new entry with empty values will be added to the configuration.

## Additional features:

- Favourite handling
  > Requires adding toolconfig commands and functions to the cobra root command!
  - Every successfully executed call can be saved as a favourite using\
    `mytool [param1] [param2] --flag1 test --save myFirstFavourite`
  - List existing favourites.\
    `mytool fav list`
  - Run save commands\
    `mytool --run myFirstFavourite` => runs `mytool [param1] [param2] --flag1 test`

## Configuration

```yaml
# ~/.toolsconfig/config.yaml
servers:
  - # The server base url. This url is used from the apps to find the credentials for the server
    url: repository.url
    # The username to access the server
    username: [ USERNAME ]
    # The password to access the server
    password: [ PASSWORD ]

    # Another example for a maven repository
  - url: maven.repository.url
    username: [ MAVEN_USERNAME ]
    password: [ MAVENPASSWORD ]

azureSubscriptions:
  - # The subscription name. Either the name or the subscriptionID can be used from the tools to get teh credentials 
    name: azureSubsciption01
    # The subscription ID
    subscriptionID: [ SUBSCRIPTION_ID ]
    # The tenant ID
    tenantID: [ TENANT_ID ]
    # The client ID
    clientID: [ CLIENT_ID ]
    # The client secret
    clientSecret: [ CLIENT_SECRET ]
generics:
  - # The key of a generic value 
    key: system-with-token
    # The value for the generic key (e.g. auth token)
    value: [ SECRET TOKEN ]
favourites:
  toolname1:
    favName:
      name: favName
      args: [ arg01, arg02, arg03]
```

### Environement Variables

You can use environment variables to use instead of the values from config files. Environment variables overrule values from config files,
but only if all values for a entry are available.

**Environment values for the example from above:**

```bash
REPOSITORY.URL_USERNAME=[USERNAME]
REPOSITORY.URL_IO_PASSWORD=[PASSWORD]
MAVEN.REPOSITORY.URL_PASSWORD=[USERNAME] 
MAVEN.REPOSITORY.URL_USERNAME=[PASSWORD]
AZURESUBSCRIPTION01_SUBSCRIPTIONID=[SUBSCRIPTION_ID]
AZURESUBSCRIPTION01_TENANTID=[TENANT_ID]
AZURESUBSCRIPTION01_CLIENTID=[CLIENT_ID]
AZURESUBSCRIPTION01_CLIENTSECRET=[CLIENT_SECRET]
SYSTEM_WITH_TOKEN_VALUE=CLIENT_SECRET
```

## Example

see [Command example](example/main.go)

run with:

```bash
go run example/main.go --help
```
