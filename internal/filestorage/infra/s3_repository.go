Go
Skip to Main Content
Search packages or symbols

Why Gosubmenu dropdown icon
Learn
Docssubmenu dropdown icon
Packages
Communitysubmenu dropdown icon
Discover Packages
 
github.com/aws/aws-sdk-go-v2/config

Go
config
package
module

Version: v1.31.0 Latest 
Published: Aug 11, 2025 
License: Apache-2.0 
Imports: 35 
Imported by: 8,576
Details
checked Valid go.mod file 
checked Redistributable license 
checked Tagged version 
checked Stable version 
Learn more about best practices
Repository
github.com/aws/aws-sdk-go-v2
Links
Open Source Insights Logo Open Source Insights
Jump to ...
 Documentation ¶
Rendered for 
linux/amd64
Overview ¶
Package config provides utilities for loading configuration from multiple sources that can be used to configure the SDK's API clients, and utilities.

The config package will load configuration from environment variables, AWS shared configuration file (~/.aws/config), and AWS shared credentials file (~/.aws/credentials).

Use the LoadDefaultConfig to load configuration from all the SDK's supported sources, and resolve credentials using the SDK's default credential chain.

LoadDefaultConfig allows for a variadic list of additional Config sources that can provide one or more configuration values which can be used to programmatically control the resolution of a specific value, or allow for broader range of additional configuration sources not supported by the SDK. A Config source implements one or more provider interfaces defined in this package. Config sources passed in will take precedence over the default environment and shared config sources used by the SDK. If one or more Config sources implement the same provider interface, priority will be handled by the order in which the sources were passed in.

A number of helpers (prefixed by “With“) are provided in this package that implement their respective provider interface. These helpers should be used for overriding configuration programmatically at runtime.

Example ¶
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func main() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	client := sts.NewFromConfig(cfg)
	identity, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Account: %s, Arn: %s", aws.ToString(identity.Account), aws.ToString(identity.Arn))
}

Share
Format
Run
Example (Custom_config) ¶
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func main() {
	ctx := context.TODO()

	// Config sources can be passed to LoadDefaultConfig, these sources can implement one or more
	// provider interfaces. These sources take priority over the standard environment and shared configuration values.
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-west-2"),
		config.WithSharedConfigProfile("customProfile"),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := sts.NewFromConfig(cfg)
	identity, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Account: %s, Arn: %s", aws.ToString(identity.Account), aws.ToString(identity.Arn))
}

Share
Format
Run
Index ¶
Constants
Variables
func DefaultSharedConfigFilename() string
func DefaultSharedCredentialsFilename() string
func GetIgnoreConfiguredEndpoints(ctx context.Context, configs []interface{}) (value bool, found bool, err error)
func LoadDefaultConfig(ctx context.Context, optFns ...func(*LoadOptions) error) (cfg aws.Config, err error)
type AssumeRoleTokenProviderNotSetError
func (e AssumeRoleTokenProviderNotSetError) Error() string
type Config
type CredentialRequiresARNError
func (e CredentialRequiresARNError) Error() string
type DefaultsModeOptions
type EnvConfig
func NewEnvConfig() (EnvConfig, error)
func (c EnvConfig) GetEC2IMDSClientEnableState() (imds.ClientEnableState, bool, error)
func (c EnvConfig) GetEC2IMDSEndpoint() (string, bool, error)
func (c EnvConfig) GetEC2IMDSEndpointMode() (imds.EndpointModeState, bool, error)
func (c EnvConfig) GetEC2IMDSV1FallbackDisabled() (bool, bool)
func (c EnvConfig) GetEnableEndpointDiscovery(ctx context.Context) (value aws.EndpointDiscoveryEnableState, found bool, err error)
func (c EnvConfig) GetIgnoreConfiguredEndpoints(context.Context) (bool, bool, error)
func (c EnvConfig) GetRetryMaxAttempts(ctx context.Context) (int, bool, error)
func (c EnvConfig) GetRetryMode(ctx context.Context) (aws.RetryMode, bool, error)
func (c EnvConfig) GetS3DisableExpressAuth() (value, ok bool)
func (c EnvConfig) GetS3DisableMultiRegionAccessPoints(ctx context.Context) (value, ok bool, err error)
func (c EnvConfig) GetS3UseARNRegion(ctx context.Context) (value, ok bool, err error)
func (c EnvConfig) GetServiceBaseEndpoint(ctx context.Context, sdkID string) (string, bool, error)
func (c EnvConfig) GetUseDualStackEndpoint(ctx context.Context) (value aws.DualStackEndpointState, found bool, err error)
func (c EnvConfig) GetUseFIPSEndpoint(ctx context.Context) (value aws.FIPSEndpointState, found bool, err error)
type HTTPClient
type IgnoreConfiguredEndpointsProvider
type LoadOptions
func (o LoadOptions) GetEC2IMDSClientEnableState() (imds.ClientEnableState, bool, error)
func (o LoadOptions) GetEC2IMDSEndpoint() (string, bool, error)
func (o LoadOptions) GetEC2IMDSEndpointMode() (imds.EndpointModeState, bool, error)
func (o LoadOptions) GetEnableEndpointDiscovery(ctx context.Context) (value aws.EndpointDiscoveryEnableState, ok bool, err error)
func (o LoadOptions) GetRetryMaxAttempts(ctx context.Context) (int, bool, error)
func (o LoadOptions) GetRetryMode(ctx context.Context) (aws.RetryMode, bool, error)
func (o LoadOptions) GetS3DisableExpressAuth() (value, ok bool)
func (o LoadOptions) GetS3DisableMultiRegionAccessPoints(ctx context.Context) (v bool, found bool, err error)
func (o LoadOptions) GetS3UseARNRegion(ctx context.Context) (v bool, found bool, err error)
func (o LoadOptions) GetServiceBaseEndpoint(context.Context, string) (string, bool, error)
func (o LoadOptions) GetUseDualStackEndpoint(ctx context.Context) (value aws.DualStackEndpointState, found bool, err error)
func (o LoadOptions) GetUseFIPSEndpoint(ctx context.Context) (value aws.FIPSEndpointState, found bool, err error)
type LoadOptionsFunc
func WithAPIOptions(v []func(*middleware.Stack) error) LoadOptionsFunc
func WithAccountIDEndpointMode(m aws.AccountIDEndpointMode) LoadOptionsFunc
func WithAfterAttempt(i smithyhttp.AfterAttemptInterceptor) LoadOptionsFunc
func WithAfterDeserialization(i smithyhttp.AfterDeserializationInterceptor) LoadOptionsFunc
func WithAfterExecution(i smithyhttp.AfterExecutionInterceptor) LoadOptionsFunc
func WithAfterSerialization(i smithyhttp.AfterSerializationInterceptor) LoadOptionsFunc
func WithAfterSigning(i smithyhttp.AfterSigningInterceptor) LoadOptionsFunc
func WithAfterTransmit(i smithyhttp.AfterTransmitInterceptor) LoadOptionsFunc
func WithAppID(ID string) LoadOptionsFunc
func WithAssumeRoleCredentialOptions(v func(*stscreds.AssumeRoleOptions)) LoadOptionsFunc
func WithAuthSchemePreference(schemeIDs ...string) LoadOptionsFunc
func WithBaseEndpoint(v string) LoadOptionsFunc
func WithBearerAuthTokenCacheOptions(v func(*smithybearer.TokenCacheOptions)) LoadOptionsFunc
func WithBearerAuthTokenProvider(v smithybearer.TokenProvider) LoadOptionsFunc
func WithBeforeAttempt(i smithyhttp.BeforeAttemptInterceptor) LoadOptionsFunc
func WithBeforeDeserialization(i smithyhttp.BeforeDeserializationInterceptor) LoadOptionsFunc
func WithBeforeExecution(i smithyhttp.BeforeExecutionInterceptor) LoadOptionsFunc
func WithBeforeRetryLoop(i smithyhttp.BeforeRetryLoopInterceptor) LoadOptionsFunc
func WithBeforeSerialization(i smithyhttp.BeforeSerializationInterceptor) LoadOptionsFunc
func WithBeforeSigning(i smithyhttp.BeforeSigningInterceptor) LoadOptionsFunc
func WithBeforeTransmit(i smithyhttp.BeforeTransmitInterceptor) LoadOptionsFunc
func WithClientLogMode(v aws.ClientLogMode) LoadOptionsFunc
func WithCredentialsCacheOptions(v func(*aws.CredentialsCacheOptions)) LoadOptionsFunc
func WithCredentialsProvider(v aws.CredentialsProvider) LoadOptionsFunc
func WithCustomCABundle(v io.Reader) LoadOptionsFunc
func WithDefaultRegion(v string) LoadOptionsFunc
func WithDefaultsMode(mode aws.DefaultsMode, optFns ...func(options *DefaultsModeOptions)) LoadOptionsFunc
func WithDisableRequestCompression(DisableRequestCompression *bool) LoadOptionsFunc
func WithEC2IMDSClientEnableState(v imds.ClientEnableState) LoadOptionsFunc
func WithEC2IMDSEndpoint(v string) LoadOptionsFunc
func WithEC2IMDSEndpointMode(v imds.EndpointModeState) LoadOptionsFunc
func WithEC2IMDSRegion(fnOpts ...func(o *UseEC2IMDSRegion)) LoadOptionsFunc
func WithEC2RoleCredentialOptions(v func(*ec2rolecreds.Options)) LoadOptionsFunc
func WithEndpointCredentialOptions(v func(*endpointcreds.Options)) LoadOptionsFunc
func WithEndpointDiscovery(v aws.EndpointDiscoveryEnableState) LoadOptionsFunc
func WithEndpointResolver(v aws.EndpointResolver) LoadOptionsFuncdeprecated
func WithEndpointResolverWithOptions(v aws.EndpointResolverWithOptions) LoadOptionsFuncdeprecated
func WithHTTPClient(v HTTPClient) LoadOptionsFunc
func WithLogConfigurationWarnings(v bool) LoadOptionsFunc
func WithLogger(v logging.Logger) LoadOptionsFunc
func WithProcessCredentialOptions(v func(*processcreds.Options)) LoadOptionsFunc
func WithRegion(v string) LoadOptionsFunc
func WithRequestChecksumCalculation(c aws.RequestChecksumCalculation) LoadOptionsFunc
func WithRequestMinCompressSizeBytes(RequestMinCompressSizeBytes *int64) LoadOptionsFunc
func WithResponseChecksumValidation(v aws.ResponseChecksumValidation) LoadOptionsFunc
func WithRetryMaxAttempts(v int) LoadOptionsFunc
func WithRetryMode(v aws.RetryMode) LoadOptionsFunc
func WithRetryer(v func() aws.Retryer) LoadOptionsFunc
func WithS3DisableExpressAuth(v bool) LoadOptionsFunc
func WithS3DisableMultiRegionAccessPoints(v bool) LoadOptionsFunc
func WithS3UseARNRegion(v bool) LoadOptionsFunc
func WithSSOProviderOptions(v func(*ssocreds.Options)) LoadOptionsFunc
func WithSSOTokenProviderOptions(v func(*ssocreds.SSOTokenProviderOptions)) LoadOptionsFunc
func WithServiceOptions(callbacks ...func(string, any)) LoadOptionsFunc
func WithSharedConfigFiles(v []string) LoadOptionsFunc
func WithSharedConfigProfile(v string) LoadOptionsFunc
func WithSharedCredentialsFiles(v []string) LoadOptionsFunc
func WithUseDualStackEndpoint(v aws.DualStackEndpointState) LoadOptionsFunc
func WithUseFIPSEndpoint(v aws.FIPSEndpointState) LoadOptionsFunc
func WithWebIdentityRoleCredentialOptions(v func(*stscreds.WebIdentityRoleOptions)) LoadOptionsFunc
type LoadSharedConfigOptions
type SSOSession
type Services
type SharedConfig
func LoadSharedConfigProfile(ctx context.Context, profile string, optFns ...func(*LoadSharedConfigOptions)) (SharedConfig, error)
func (c SharedConfig) GetEC2IMDSEndpoint() (string, bool, error)
func (c SharedConfig) GetEC2IMDSEndpointMode() (imds.EndpointModeState, bool, error)
func (c SharedConfig) GetEC2IMDSV1FallbackDisabled() (bool, bool)
func (c SharedConfig) GetEnableEndpointDiscovery(ctx context.Context) (value aws.EndpointDiscoveryEnableState, ok bool, err error)
func (c SharedConfig) GetIgnoreConfiguredEndpoints(context.Context) (bool, bool, error)
func (c SharedConfig) GetRetryMaxAttempts(ctx context.Context) (value int, ok bool, err error)
func (c SharedConfig) GetRetryMode(ctx context.Context) (value aws.RetryMode, ok bool, err error)
func (c SharedConfig) GetS3DisableExpressAuth() (value, ok bool)
func (c SharedConfig) GetS3DisableMultiRegionAccessPoints(ctx context.Context) (value, ok bool, err error)
func (c SharedConfig) GetS3UseARNRegion(ctx context.Context) (value, ok bool, err error)
func (c SharedConfig) GetServiceBaseEndpoint(ctx context.Context, sdkID string) (string, bool, error)
func (c SharedConfig) GetUseDualStackEndpoint(ctx context.Context) (value aws.DualStackEndpointState, found bool, err error)
func (c SharedConfig) GetUseFIPSEndpoint(ctx context.Context) (value aws.FIPSEndpointState, found bool, err error)
type SharedConfigAssumeRoleError
func (e SharedConfigAssumeRoleError) Error() string
func (e SharedConfigAssumeRoleError) Unwrap() error
type SharedConfigLoadError
func (e SharedConfigLoadError) Error() string
func (e SharedConfigLoadError) Unwrap() error
type SharedConfigProfileNotExistError
func (e SharedConfigProfileNotExistError) Error() string
func (e SharedConfigProfileNotExistError) Unwrap() error
type UseEC2IMDSRegion
Examples ¶
Package
Package (Custom_config)
WithAPIOptions
WithAssumeRoleCredentialOptions
WithCredentialsCacheOptions
WithCredentialsProvider
WithEC2IMDSRegion
WithEndpointResolver
WithEndpointResolverWithOptions
WithHTTPClient
WithRegion
WithSharedConfigProfile
WithWebIdentityRoleCredentialOptions
Constants ¶
View Source
const CredentialsSourceName = "EnvConfigCredentials"
CredentialsSourceName provides a name of the provider when config is loaded from environment.

View Source
const (

	// DefaultSharedConfigProfile is the default profile to be used when
	// loading configuration from the config files if another profile name
	// is not provided.
	DefaultSharedConfigProfile = `default`
)
Variables ¶
View Source
var DefaultSharedConfigFiles = []string{
	DefaultSharedConfigFilename(),
}
DefaultSharedConfigFiles is a slice of the default shared config files that the will be used in order to load the SharedConfig.

View Source
var DefaultSharedCredentialsFiles = []string{
	DefaultSharedCredentialsFilename(),
}
DefaultSharedCredentialsFiles is a slice of the default shared credentials files that the will be used in order to load the SharedConfig.

Functions ¶
func DefaultSharedConfigFilename ¶
func DefaultSharedConfigFilename() string
DefaultSharedConfigFilename returns the SDK's default file path for the shared config file.

Builds the shared config file path based on the OS's platform.

Linux/Unix: $HOME/.aws/config
Windows: %USERPROFILE%\.aws\config
func DefaultSharedCredentialsFilename ¶
func DefaultSharedCredentialsFilename() string
DefaultSharedCredentialsFilename returns the SDK's default file path for the shared credentials file.

Builds the shared config file path based on the OS's platform.

Linux/Unix: $HOME/.aws/credentials
Windows: %USERPROFILE%\.aws\credentials
func GetIgnoreConfiguredEndpoints ¶
added in v1.19.1
func GetIgnoreConfiguredEndpoints(ctx context.Context, configs []interface{}) (value bool, found bool, err error)
GetIgnoreConfiguredEndpoints is used in knowing when to disable configured endpoints feature.

func LoadDefaultConfig ¶
func LoadDefaultConfig(ctx context.Context, optFns ...func(*LoadOptions) error) (cfg aws.Config, err error)
LoadDefaultConfig reads the SDK's default external configurations, and populates an AWS Config with the values from the external configurations.

An optional variadic set of additional Config values can be provided as input that will be prepended to the configs slice. Use this to add custom configuration. The custom configurations must satisfy the respective providers for their data or the custom data will be ignored by the resolvers and config loaders.

cfg, err := config.LoadDefaultConfig( context.TODO(),
   config.WithSharedConfigProfile("test-profile"),
)
if err != nil {
   panic(fmt.Sprintf("failed loading config, %v", err))
}
The default configuration sources are: * Environment Variables * Shared Configuration and Shared Credentials files.

Types ¶
type AssumeRoleTokenProviderNotSetError ¶
type AssumeRoleTokenProviderNotSetError struct{}
AssumeRoleTokenProviderNotSetError is an error returned when creating a session when the MFAToken option is not set when shared config is configured load assume a role with an MFA token.

func (AssumeRoleTokenProviderNotSetError) Error ¶
func (e AssumeRoleTokenProviderNotSetError) Error() string
Error is the error message

type Config ¶
type Config interface{}
A Config represents a generic configuration value or set of values. This type will be used by the AWSConfigResolvers to extract

General the Config type will use type assertion against the Provider interfaces to extract specific data from the Config.

type CredentialRequiresARNError ¶
type CredentialRequiresARNError struct {
	// type of credentials that were configured.
	Type string

	// Profile name the credentials were in.
	Profile string
}
CredentialRequiresARNError provides the error for shared config credentials that are incorrectly configured in the shared config or credentials file.

func (CredentialRequiresARNError) Error ¶
func (e CredentialRequiresARNError) Error() string
Error satisfies the error interface.

type DefaultsModeOptions ¶
added in v1.13.0
type DefaultsModeOptions struct {
	// The SDK configuration defaults mode. Defaults to legacy if not specified.
	//
	// Supported modes are: auto, cross-region, in-region, legacy, mobile, standard
	Mode aws.DefaultsMode

	// The EC2 Instance Metadata Client that should be used when performing environment
	// discovery when aws.DefaultsModeAuto is set.
	//
	// If not specified the SDK will construct a client if the instance metadata service has not been disabled by
	// the AWS_EC2_METADATA_DISABLED environment variable.
	IMDSClient *imds.Client
}
DefaultsModeOptions is the set of options that are used to configure

type EnvConfig ¶
type EnvConfig struct {
	// Environment configuration values. If set both Access Key ID and Secret Access
	// Key must be provided. Session Token and optionally also be provided, but is
	// not required.
	//
	//	# Access Key ID
	//	AWS_ACCESS_KEY_ID=AKID
	//	AWS_ACCESS_KEY=AKID # only read if AWS_ACCESS_KEY_ID is not set.
	//
	//	# Secret Access Key
	//	AWS_SECRET_ACCESS_KEY=SECRET
	//	AWS_SECRET_KEY=SECRET # only read if AWS_SECRET_ACCESS_KEY is not set.
	//
	//	# Session Token
	//	AWS_SESSION_TOKEN=TOKEN
	Credentials aws.Credentials

	// ContainerCredentialsEndpoint value is the HTTP enabled endpoint to retrieve credentials
	// using the endpointcreds.Provider
	ContainerCredentialsEndpoint string

	// ContainerCredentialsRelativePath is the relative URI path that will be used when attempting to retrieve
	// credentials from the container endpoint.
	ContainerCredentialsRelativePath string

	// ContainerAuthorizationToken is the authorization token that will be included in the HTTP Authorization
	// header when attempting to retrieve credentials from the container credentials endpoint.
	ContainerAuthorizationToken string

	// Region value will instruct the SDK where to make service API requests to. If is
	// not provided in the environment the region must be provided before a service
	// client request is made.
	//
	//	AWS_REGION=us-west-2
	//	AWS_DEFAULT_REGION=us-west-2
	Region string

	// Profile name the SDK should load use when loading shared configuration from the
	// shared configuration files. If not provided "default" will be used as the
	// profile name.
	//
	//	AWS_PROFILE=my_profile
	//	AWS_DEFAULT_PROFILE=my_profile
	SharedConfigProfile string

	// Shared credentials file path can be set to instruct the SDK to use an alternate
	// file for the shared credentials. If not set the file will be loaded from
	// $HOME/.aws/credentials on Linux/Unix based systems, and
	// %USERPROFILE%\.aws\credentials on Windows.
	//
	//	AWS_SHARED_CREDENTIALS_FILE=$HOME/my_shared_credentials
	SharedCredentialsFile string

	// Shared config file path can be set to instruct the SDK to use an alternate
	// file for the shared config. If not set the file will be loaded from
	// $HOME/.aws/config on Linux/Unix based systems, and
	// %USERPROFILE%\.aws\config on Windows.
	//
	//	AWS_CONFIG_FILE=$HOME/my_shared_config
	SharedConfigFile string

	// Sets the path to a custom Credentials Authority (CA) Bundle PEM file
	// that the SDK will use instead of the system's root CA bundle.
	// Only use this if you want to configure the SDK to use a custom set
	// of CAs.
	//
	// Enabling this option will attempt to merge the Transport
	// into the SDK's HTTP client. If the client's Transport is
	// not a http.Transport an error will be returned. If the
	// Transport's TLS config is set this option will cause the
	// SDK to overwrite the Transport's TLS config's  RootCAs value.
	//
	// Setting a custom HTTPClient in the aws.Config options will override this setting.
	// To use this option and custom HTTP client, the HTTP client needs to be provided
	// when creating the config. Not the service client.
	//
	//  AWS_CA_BUNDLE=$HOME/my_custom_ca_bundle
	CustomCABundle string

	// Enables endpoint discovery via environment variables.
	//
	//	AWS_ENABLE_ENDPOINT_DISCOVERY=true
	EnableEndpointDiscovery aws.EndpointDiscoveryEnableState

	// Specifies the WebIdentity token the SDK should use to assume a role
	// with.
	//
	//  AWS_WEB_IDENTITY_TOKEN_FILE=file_path
	WebIdentityTokenFilePath string

	// Specifies the IAM role arn to use when assuming an role.
	//
	//  AWS_ROLE_ARN=role_arn
	RoleARN string

	// Specifies the IAM role session name to use when assuming a role.
	//
	//  AWS_ROLE_SESSION_NAME=session_name
	RoleSessionName string

	// Specifies if the S3 service should allow ARNs to direct the region
	// the client's requests are sent to.
	//
	// AWS_S3_USE_ARN_REGION=true
	S3UseARNRegion *bool

	// Specifies if the EC2 IMDS service client is enabled.
	//
	// AWS_EC2_METADATA_DISABLED=true
	EC2IMDSClientEnableState imds.ClientEnableState

	// Specifies if EC2 IMDSv1 fallback is disabled.
	//
	// AWS_EC2_METADATA_V1_DISABLED=true
	EC2IMDSv1Disabled *bool

	// Specifies the EC2 Instance Metadata Service default endpoint selection mode (IPv4 or IPv6)
	//
	// AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE=IPv6
	EC2IMDSEndpointMode imds.EndpointModeState

	// Specifies the EC2 Instance Metadata Service endpoint to use. If specified it overrides EC2IMDSEndpointMode.
	//
	// AWS_EC2_METADATA_SERVICE_ENDPOINT=http://fd00:ec2::254
	EC2IMDSEndpoint string

	// Specifies if the S3 service should disable multi-region access points
	// support.
	//
	// AWS_S3_DISABLE_MULTIREGION_ACCESS_POINTS=true
	S3DisableMultiRegionAccessPoints *bool

	// Specifies that SDK clients must resolve a dual-stack endpoint for
	// services.
	//
	// AWS_USE_DUALSTACK_ENDPOINT=true
	UseDualStackEndpoint aws.DualStackEndpointState

	// Specifies that SDK clients must resolve a FIPS endpoint for
	// services.
	//
	// AWS_USE_FIPS_ENDPOINT=true
	UseFIPSEndpoint aws.FIPSEndpointState

	// Specifies the SDK Defaults Mode used by services.
	//
	// AWS_DEFAULTS_MODE=standard
	DefaultsMode aws.DefaultsMode

	// Specifies the maximum number attempts an API client will call an
	// operation that fails with a retryable error.
	//
	// AWS_MAX_ATTEMPTS=3
	RetryMaxAttempts int

	// Specifies the retry model the API client will be created with.
	//
	// aws_retry_mode=standard
	RetryMode aws.RetryMode

	// aws sdk app ID that can be added to user agent header string
	AppID string

	// Flag used to disable configured endpoints.
	IgnoreConfiguredEndpoints *bool

	// Value to contain configured endpoints to be propagated to
	// corresponding endpoint resolution field.
	BaseEndpoint string

	// determine if request compression is allowed, default to false
	// retrieved from env var AWS_DISABLE_REQUEST_COMPRESSION
	DisableRequestCompression *bool

	// inclusive threshold request body size to trigger compression,
	// default to 10240 and must be within 0 and 10485760 bytes inclusive
	// retrieved from env var AWS_REQUEST_MIN_COMPRESSION_SIZE_BYTES
	RequestMinCompressSizeBytes *int64

	// Whether S3Express auth is disabled.
	//
	// This will NOT prevent requests from being made to S3Express buckets, it
	// will only bypass the modified endpoint routing and signing behaviors
	// associated with the feature.
	S3DisableExpressAuth *bool

	// Indicates whether account ID will be required/ignored in endpoint2.0 routing
	AccountIDEndpointMode aws.AccountIDEndpointMode

	// Indicates whether request checksum should be calculated
	RequestChecksumCalculation aws.RequestChecksumCalculation

	// Indicates whether response checksum should be validated
	ResponseChecksumValidation aws.ResponseChecksumValidation

	// Priority list of preferred auth scheme names (e.g. sigv4a).
	AuthSchemePreference []string
}
EnvConfig is a collection of environment values the SDK will read setup config from. All environment values are optional. But some values such as credentials require multiple values to be complete or the values will be ignored.

func NewEnvConfig ¶
func NewEnvConfig() (EnvConfig, error)
NewEnvConfig retrieves the SDK's environment configuration. See `EnvConfig` for the values that will be retrieved.

func (EnvConfig) GetEC2IMDSClientEnableState ¶
added in v1.5.0
func (c EnvConfig) GetEC2IMDSClientEnableState() (imds.ClientEnableState, bool, error)
GetEC2IMDSClientEnableState implements a EC2IMDSClientEnableState options resolver interface.

func (EnvConfig) GetEC2IMDSEndpoint ¶
added in v1.5.0
func (c EnvConfig) GetEC2IMDSEndpoint() (string, bool, error)
GetEC2IMDSEndpoint implements a EC2IMDSEndpoint option resolver interface.

func (EnvConfig) GetEC2IMDSEndpointMode ¶
added in v1.5.0
func (c EnvConfig) GetEC2IMDSEndpointMode() (imds.EndpointModeState, bool, error)
GetEC2IMDSEndpointMode implements a EC2IMDSEndpointMode option resolver interface.

func (EnvConfig) GetEC2IMDSV1FallbackDisabled ¶
added in v1.22.0
func (c EnvConfig) GetEC2IMDSV1FallbackDisabled() (bool, bool)
GetEC2IMDSV1FallbackDisabled implements an EC2IMDSV1FallbackDisabled option resolver interface.

func (EnvConfig) GetEnableEndpointDiscovery ¶
func (c EnvConfig) GetEnableEndpointDiscovery(ctx context.Context) (value aws.EndpointDiscoveryEnableState, found bool, err error)
GetEnableEndpointDiscovery returns resolved value for EnableEndpointDiscovery env variable setting.

func (EnvConfig) GetIgnoreConfiguredEndpoints ¶
added in v1.21.0
func (c EnvConfig) GetIgnoreConfiguredEndpoints(context.Context) (bool, bool, error)
GetIgnoreConfiguredEndpoints is used in knowing when to disable configured endpoints feature.

func (EnvConfig) GetRetryMaxAttempts ¶
added in v1.14.0
func (c EnvConfig) GetRetryMaxAttempts(ctx context.Context) (int, bool, error)
GetRetryMaxAttempts returns the value of AWS_MAX_ATTEMPTS if was specified, and not 0.

func (EnvConfig) GetRetryMode ¶
added in v1.14.0
func (c EnvConfig) GetRetryMode(ctx context.Context) (aws.RetryMode, bool, error)
GetRetryMode returns the RetryMode of AWS_RETRY_MODE if was specified, and a valid value.

func (EnvConfig) GetS3DisableExpressAuth ¶
added in v1.25.7
func (c EnvConfig) GetS3DisableExpressAuth() (value, ok bool)
GetS3DisableExpressAuth returns the configured value for [EnvConfig.S3DisableExpressAuth].

func (EnvConfig) GetS3DisableMultiRegionAccessPoints ¶
added in v1.18.34
func (c EnvConfig) GetS3DisableMultiRegionAccessPoints(ctx context.Context) (value, ok bool, err error)
GetS3DisableMultiRegionAccessPoints returns whether to disable multi-region access point support for the S3 client.

func (EnvConfig) GetS3UseARNRegion ¶
func (c EnvConfig) GetS3UseARNRegion(ctx context.Context) (value, ok bool, err error)
GetS3UseARNRegion returns whether to allow ARNs to direct the region the S3 client's requests are sent to.

func (EnvConfig) GetServiceBaseEndpoint ¶
added in v1.21.0
func (c EnvConfig) GetServiceBaseEndpoint(ctx context.Context, sdkID string) (string, bool, error)
GetServiceBaseEndpoint is used to retrieve a normalized SDK ID for use with configured endpoints.

func (EnvConfig) GetUseDualStackEndpoint ¶
added in v1.10.0
func (c EnvConfig) GetUseDualStackEndpoint(ctx context.Context) (value aws.DualStackEndpointState, found bool, err error)
GetUseDualStackEndpoint returns whether the service's dual-stack endpoint should be used for requests.

func (EnvConfig) GetUseFIPSEndpoint ¶
added in v1.10.0
func (c EnvConfig) GetUseFIPSEndpoint(ctx context.Context) (value aws.FIPSEndpointState, found bool, err error)
GetUseFIPSEndpoint returns whether the service's FIPS endpoint should be used for requests.

type HTTPClient ¶
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}
HTTPClient is an HTTP client implementation

type IgnoreConfiguredEndpointsProvider ¶
added in v1.19.1
type IgnoreConfiguredEndpointsProvider interface {
	GetIgnoreConfiguredEndpoints(ctx context.Context) (bool, bool, error)
}
IgnoreConfiguredEndpointsProvider is needed to search for all providers that provide a flag to disable configured endpoints.

type LoadOptions ¶
added in v0.4.0
type LoadOptions struct {

	// Region is the region to send requests to.
	Region string

	// Credentials object to use when signing requests.
	Credentials aws.CredentialsProvider

	// Token provider for authentication operations with bearer authentication.
	BearerAuthTokenProvider smithybearer.TokenProvider

	// HTTPClient the SDK's API clients will use to invoke HTTP requests.
	HTTPClient HTTPClient

	// EndpointResolver that can be used to provide or override an endpoint for
	// the given service and region.
	//
	// See the `aws.EndpointResolver` documentation on usage.
	//
	// Deprecated: See EndpointResolverWithOptions
	EndpointResolver aws.EndpointResolver

	// EndpointResolverWithOptions that can be used to provide or override an
	// endpoint for the given service and region.
	//
	// See the `aws.EndpointResolverWithOptions` documentation on usage.
	EndpointResolverWithOptions aws.EndpointResolverWithOptions

	// RetryMaxAttempts specifies the maximum number attempts an API client
	// will call an operation that fails with a retryable error.
	//
	// This value will only be used if Retryer option is nil.
	RetryMaxAttempts int

	// RetryMode specifies the retry model the API client will be created with.
	//
	// This value will only be used if Retryer option is nil.
	RetryMode aws.RetryMode

	// Retryer is a function that provides a Retryer implementation. A Retryer
	// guides how HTTP requests should be retried in case of recoverable
	// failures.
	//
	// If not nil, RetryMaxAttempts, and RetryMode will be ignored.
	Retryer func() aws.Retryer

	// APIOptions provides the set of middleware mutations modify how the API
	// client requests will be handled. This is useful for adding additional
	// tracing data to a request, or changing behavior of the SDK's client.
	APIOptions []func(*middleware.Stack) error

	// Logger writer interface to write logging messages to.
	Logger logging.Logger

	// ClientLogMode is used to configure the events that will be sent to the
	// configured logger. This can be used to configure the logging of signing,
	// retries, request, and responses of the SDK clients.
	//
	// See the ClientLogMode type documentation for the complete set of logging
	// modes and available configuration.
	ClientLogMode *aws.ClientLogMode

	// SharedConfigProfile is the profile to be used when loading the SharedConfig
	SharedConfigProfile string

	// SharedConfigFiles is the slice of custom shared config files to use when
	// loading the SharedConfig. A non-default profile used within config file
	// must have name defined with prefix 'profile '. eg [profile xyz]
	// indicates a profile with name 'xyz'. To read more on the format of the
	// config file, please refer the documentation at
	// https://docs.aws.amazon.com/credref/latest/refdocs/file-format.html#file-format-config
	//
	// If duplicate profiles are provided within the same, or across multiple
	// shared config files, the next parsed profile will override only the
	// properties that conflict with the previously defined profile. Note that
	// if duplicate profiles are provided within the SharedCredentialsFiles and
	// SharedConfigFiles, the properties defined in shared credentials file
	// take precedence.
	SharedConfigFiles []string

	// SharedCredentialsFile is the slice of custom shared credentials files to
	// use when loading the SharedConfig. The profile name used within
	// credentials file must not prefix 'profile '. eg [xyz] indicates a
	// profile with name 'xyz'. Profile declared as [profile xyz] will be
	// ignored. To read more on the format of the credentials file, please
	// refer the documentation at
	// https://docs.aws.amazon.com/credref/latest/refdocs/file-format.html#file-format-creds
	//
	// If duplicate profiles are provided with a same, or across multiple
	// shared credentials files, the next parsed profile will override only
	// properties that conflict with the previously defined profile. Note that
	// if duplicate profiles are provided within the SharedCredentialsFiles and
	// SharedConfigFiles, the properties defined in shared credentials file
	// take precedence.
	SharedCredentialsFiles []string

	// CustomCABundle is CA bundle PEM bytes reader
	CustomCABundle io.Reader

	// DefaultRegion is the fall back region, used if a region was not resolved
	// from other sources
	DefaultRegion string

	// UseEC2IMDSRegion indicates if SDK should retrieve the region
	// from the EC2 Metadata service
	UseEC2IMDSRegion *UseEC2IMDSRegion

	// CredentialsCacheOptions is a function for setting the
	// aws.CredentialsCacheOptions
	CredentialsCacheOptions func(*aws.CredentialsCacheOptions)

	// BearerAuthTokenCacheOptions is a function for setting the smithy-go
	// auth/bearer#TokenCacheOptions
	BearerAuthTokenCacheOptions func(*smithybearer.TokenCacheOptions)

	// SSOTokenProviderOptions is a function for setting the
	// credentials/ssocreds.SSOTokenProviderOptions
	SSOTokenProviderOptions func(*ssocreds.SSOTokenProviderOptions)

	// ProcessCredentialOptions is a function for setting
	// the processcreds.Options
	ProcessCredentialOptions func(*processcreds.Options)

	// EC2RoleCredentialOptions is a function for setting
	// the ec2rolecreds.Options
	EC2RoleCredentialOptions func(*ec2rolecreds.Options)

	// EndpointCredentialOptions is a function for setting
	// the endpointcreds.Options
	EndpointCredentialOptions func(*endpointcreds.Options)

	// WebIdentityRoleCredentialOptions is a function for setting
	// the stscreds.WebIdentityRoleOptions
	WebIdentityRoleCredentialOptions func(*stscreds.WebIdentityRoleOptions)

	// AssumeRoleCredentialOptions is a function for setting the
	// stscreds.AssumeRoleOptions
	AssumeRoleCredentialOptions func(*stscreds.AssumeRoleOptions)

	// SSOProviderOptions is a function for setting
	// the ssocreds.Options
	SSOProviderOptions func(options *ssocreds.Options)

	// LogConfigurationWarnings when set to true, enables logging
	// configuration warnings
	LogConfigurationWarnings *bool

	// S3UseARNRegion specifies if the S3 service should allow ARNs to direct
	// the region, the client's requests are sent to.
	S3UseARNRegion *bool

	// S3DisableMultiRegionAccessPoints specifies if the S3 service should disable
	// the S3 Multi-Region access points feature.
	S3DisableMultiRegionAccessPoints *bool

	// EnableEndpointDiscovery specifies if endpoint discovery is enable for
	// the client.
	EnableEndpointDiscovery aws.EndpointDiscoveryEnableState

	// Specifies if the EC2 IMDS service client is enabled.
	//
	// AWS_EC2_METADATA_DISABLED=true
	EC2IMDSClientEnableState imds.ClientEnableState

	// Specifies the EC2 Instance Metadata Service default endpoint selection
	// mode (IPv4 or IPv6)
	EC2IMDSEndpointMode imds.EndpointModeState

	// Specifies the EC2 Instance Metadata Service endpoint to use. If
	// specified it overrides EC2IMDSEndpointMode.
	EC2IMDSEndpoint string

	// Specifies that SDK clients must resolve a dual-stack endpoint for
	// services.
	UseDualStackEndpoint aws.DualStackEndpointState

	// Specifies that SDK clients must resolve a FIPS endpoint for
	// services.
	UseFIPSEndpoint aws.FIPSEndpointState

	// Specifies the SDK configuration mode for defaults.
	DefaultsModeOptions DefaultsModeOptions

	// The sdk app ID retrieved from env var or shared config to be added to request user agent header
	AppID string

	// Specifies whether an operation request could be compressed
	DisableRequestCompression *bool

	// The inclusive min bytes of a request body that could be compressed
	RequestMinCompressSizeBytes *int64

	// Whether S3 Express auth is disabled.
	S3DisableExpressAuth *bool

	// Whether account id should be built into endpoint resolution
	AccountIDEndpointMode aws.AccountIDEndpointMode

	// Specify if request checksum should be calculated
	RequestChecksumCalculation aws.RequestChecksumCalculation

	// Specifies if response checksum should be validated
	ResponseChecksumValidation aws.ResponseChecksumValidation

	// Service endpoint override. This value is not necessarily final and is
	// passed to the service's EndpointResolverV2 for further delegation.
	BaseEndpoint string

	// Registry of operation interceptors.
	Interceptors smithyhttp.InterceptorRegistry

	// Priority list of preferred auth scheme names (e.g. sigv4a).
	AuthSchemePreference []string

	// ServiceOptions provides service specific configuration options that will be applied
	// when constructing clients for specific services. Each callback function receives the service ID
	// and the service's Options struct, allowing for dynamic configuration based on the service.
	ServiceOptions []func(string, any)
}
LoadOptions are discrete set of options that are valid for loading the configuration

func (LoadOptions) GetEC2IMDSClientEnableState ¶
added in v1.5.0
func (o LoadOptions) GetEC2IMDSClientEnableState() (imds.ClientEnableState, bool, error)
GetEC2IMDSClientEnableState implements a EC2IMDSClientEnableState options resolver interface.

func (LoadOptions) GetEC2IMDSEndpoint ¶
added in v1.5.0
func (o LoadOptions) GetEC2IMDSEndpoint() (string, bool, error)
GetEC2IMDSEndpoint implements a EC2IMDSEndpoint option resolver interface.

func (LoadOptions) GetEC2IMDSEndpointMode ¶
added in v1.5.0
func (o LoadOptions) GetEC2IMDSEndpointMode() (imds.EndpointModeState, bool, error)
GetEC2IMDSEndpointMode implements a EC2IMDSEndpointMode option resolver interface.

func (LoadOptions) GetEnableEndpointDiscovery ¶
added in v1.4.0
func (o LoadOptions) GetEnableEndpointDiscovery(ctx context.Context) (value aws.EndpointDiscoveryEnableState, ok bool, err error)
GetEnableEndpointDiscovery returns if the EnableEndpointDiscovery flag is set.

func (LoadOptions) GetRetryMaxAttempts ¶
added in v1.14.0
func (o LoadOptions) GetRetryMaxAttempts(ctx context.Context) (int, bool, error)
GetRetryMaxAttempts returns the RetryMaxAttempts if specified in the LoadOptions and not 0.

func (LoadOptions) GetRetryMode ¶
added in v1.14.0
func (o LoadOptions) GetRetryMode(ctx context.Context) (aws.RetryMode, bool, error)
GetRetryMode returns the RetryMode specified in the LoadOptions.

func (LoadOptions) GetS3DisableExpressAuth ¶
added in v1.25.8
func (o LoadOptions) GetS3DisableExpressAuth() (value, ok bool)
GetS3DisableExpressAuth returns the configured value for [EnvConfig.S3DisableExpressAuth].

func (LoadOptions) GetS3DisableMultiRegionAccessPoints ¶
added in v1.18.34
func (o LoadOptions) GetS3DisableMultiRegionAccessPoints(ctx context.Context) (v bool, found bool, err error)
GetS3DisableMultiRegionAccessPoints returns whether to disable the S3 multi-region access points feature.

func (LoadOptions) GetS3UseARNRegion ¶
added in v0.4.0
func (o LoadOptions) GetS3UseARNRegion(ctx context.Context) (v bool, found bool, err error)
GetS3UseARNRegion returns whether to allow ARNs to direct the region the S3 client's requests are sent to.

func (LoadOptions) GetServiceBaseEndpoint ¶
added in v1.28.0
func (o LoadOptions) GetServiceBaseEndpoint(context.Context, string) (string, bool, error)
GetServiceBaseEndpoint satisfies (internal/configsources).ServiceBaseEndpointProvider.

The sdkID value is unused because LoadOptions only supports setting a GLOBAL endpoint override. In-code, per-service endpoint overrides are performed via functional options in service client space.

func (LoadOptions) GetUseDualStackEndpoint ¶
added in v1.10.0
func (o LoadOptions) GetUseDualStackEndpoint(ctx context.Context) (value aws.DualStackEndpointState, found bool, err error)
GetUseDualStackEndpoint returns whether the service's dual-stack endpoint should be used for requests.

func (LoadOptions) GetUseFIPSEndpoint ¶
added in v1.10.0
func (o LoadOptions) GetUseFIPSEndpoint(ctx context.Context) (value aws.FIPSEndpointState, found bool, err error)
GetUseFIPSEndpoint returns whether the service's FIPS endpoint should be used for requests.

type LoadOptionsFunc ¶
added in v0.4.0
type LoadOptionsFunc func(*LoadOptions) error
LoadOptionsFunc is a type alias for LoadOptions functional option

func WithAPIOptions ¶
func WithAPIOptions(v []func(*middleware.Stack) error) LoadOptionsFunc
WithAPIOptions is a helper function to construct functional options that sets APIOptions on LoadOptions. If APIOptions is set to nil, the APIOptions value is ignored. If multiple WithAPIOptions calls are made, the last call overrides the previous call values.

Example ¶
func WithAccountIDEndpointMode ¶
added in v1.27.19
func WithAccountIDEndpointMode(m aws.AccountIDEndpointMode) LoadOptionsFunc
WithAccountIDEndpointMode is a helper function to construct functional options that sets AccountIDEndpointMode on config's LoadOptions

func WithAfterAttempt ¶
added in v1.30.0
func WithAfterAttempt(i smithyhttp.AfterAttemptInterceptor) LoadOptionsFunc
WithAfterAttempt adds the AfterAttemptInterceptor to config.

func WithAfterDeserialization ¶
added in v1.30.0
func WithAfterDeserialization(i smithyhttp.AfterDeserializationInterceptor) LoadOptionsFunc
WithAfterDeserialization adds the AfterDeserializationInterceptor to config.

func WithAfterExecution ¶
added in v1.30.0
func WithAfterExecution(i smithyhttp.AfterExecutionInterceptor) LoadOptionsFunc
WithAfterExecution adds the AfterExecutionInterceptor to config.

func WithAfterSerialization ¶
added in v1.30.0
func WithAfterSerialization(i smithyhttp.AfterSerializationInterceptor) LoadOptionsFunc
WithAfterSerialization adds the AfterSerializationInterceptor to config.

func WithAfterSigning ¶
added in v1.30.0
func WithAfterSigning(i smithyhttp.AfterSigningInterceptor) LoadOptionsFunc
WithAfterSigning adds the AfterSigningInterceptor to config.

func WithAfterTransmit ¶
added in v1.30.0
func WithAfterTransmit(i smithyhttp.AfterTransmitInterceptor) LoadOptionsFunc
WithAfterTransmit adds the AfterTransmitInterceptor to config.

func WithAppID ¶
added in v1.18.28
func WithAppID(ID string) LoadOptionsFunc
WithAppID is a helper function to construct functional options that sets AppID on config's LoadOptions.

func WithAssumeRoleCredentialOptions ¶
added in v0.2.0
func WithAssumeRoleCredentialOptions(v func(*stscreds.AssumeRoleOptions)) LoadOptionsFunc
WithAssumeRoleCredentialOptions is a helper function to construct functional options that sets a function to use stscreds.AssumeRoleOptions on config's LoadOptions. If assume role credentials options is set to nil, the assume role credentials value will be ignored. If multiple WithAssumeRoleCredentialOptions calls are made, the last call overrides the previous call values.

Example ¶
func WithAuthSchemePreference ¶
added in v1.30.3
func WithAuthSchemePreference(schemeIDs ...string) LoadOptionsFunc
WithAuthSchemePreference sets the priority order of auth schemes on config.

Schemes are expressed as names e.g. sigv4a or sigv4.

func WithBaseEndpoint ¶
added in v1.28.0
func WithBaseEndpoint(v string) LoadOptionsFunc
WithBaseEndpoint is a helper function to construct functional options that sets BaseEndpoint on config's LoadOptions. Empty values have no effect, and subsequent calls to this API override previous ones.

This is an in-code setting, therefore, any value set using this hook takes precedence over and will override ALL environment and shared config directives that set endpoint URLs. Functional options on service clients have higher specificity, and functional options that modify the value of BaseEndpoint on a client will take precedence over this setting.

func WithBearerAuthTokenCacheOptions ¶
added in v1.17.2
func WithBearerAuthTokenCacheOptions(v func(*smithybearer.TokenCacheOptions)) LoadOptionsFunc
WithBearerAuthTokenCacheOptions is a helper function to construct functional options that sets a function to modify the TokenCacheOptions the smithy-go auth/bearer#TokenCache will be configured with, if the TokenCache is used by the configuration loader.

If multiple WithBearerAuthTokenCacheOptions calls are made, the last call overrides the previous call values.

func WithBearerAuthTokenProvider ¶
added in v1.17.2
func WithBearerAuthTokenProvider(v smithybearer.TokenProvider) LoadOptionsFunc
WithBearerAuthTokenProvider is a helper function to construct functional options that sets Credential provider value on config's LoadOptions. If credentials provider is set to nil, the credentials provider value will be ignored. If multiple WithBearerAuthTokenProvider calls are made, the last call overrides the previous call values.

func WithBeforeAttempt ¶
added in v1.30.0
func WithBeforeAttempt(i smithyhttp.BeforeAttemptInterceptor) LoadOptionsFunc
WithBeforeAttempt adds the BeforeAttemptInterceptor to config.

func WithBeforeDeserialization ¶
added in v1.30.0
func WithBeforeDeserialization(i smithyhttp.BeforeDeserializationInterceptor) LoadOptionsFunc
WithBeforeDeserialization adds the BeforeDeserializationInterceptor to config.

func WithBeforeExecution ¶
added in v1.30.0
func WithBeforeExecution(i smithyhttp.BeforeExecutionInterceptor) LoadOptionsFunc
WithBeforeExecution adds the BeforeExecutionInterceptor to config.

func WithBeforeRetryLoop ¶
added in v1.30.0
func WithBeforeRetryLoop(i smithyhttp.BeforeRetryLoopInterceptor) LoadOptionsFunc
WithBeforeRetryLoop adds the BeforeRetryLoopInterceptor to config.

func WithBeforeSerialization ¶
added in v1.30.0
func WithBeforeSerialization(i smithyhttp.BeforeSerializationInterceptor) LoadOptionsFunc
WithBeforeSerialization adds the BeforeSerializationInterceptor to config.

func WithBeforeSigning ¶
added in v1.30.0
func WithBeforeSigning(i smithyhttp.BeforeSigningInterceptor) LoadOptionsFunc
WithBeforeSigning adds the BeforeSigningInterceptor to config.

func WithBeforeTransmit ¶
added in v1.30.0
func WithBeforeTransmit(i smithyhttp.BeforeTransmitInterceptor) LoadOptionsFunc
WithBeforeTransmit adds the BeforeTransmitInterceptor to config.

func WithClientLogMode ¶
added in v0.3.0
func WithClientLogMode(v aws.ClientLogMode) LoadOptionsFunc
WithClientLogMode is a helper function to construct functional options that sets client log mode on LoadOptions. If client log mode is set to nil, the client log mode value will be ignored. If multiple WithClientLogMode calls are made, the last call overrides the previous call values.

func WithCredentialsCacheOptions ¶
added in v1.12.0
func WithCredentialsCacheOptions(v func(*aws.CredentialsCacheOptions)) LoadOptionsFunc
WithCredentialsCacheOptions is a helper function to construct functional options that sets a function to modify the aws.CredentialsCacheOptions the aws.CredentialsCache will be configured with, if the CredentialsCache is used by the configuration loader.

If multiple WithCredentialsCacheOptions calls are made, the last call overrides the previous call values.

Example ¶
func WithCredentialsProvider ¶
func WithCredentialsProvider(v aws.CredentialsProvider) LoadOptionsFunc
WithCredentialsProvider is a helper function to construct functional options that sets Credential provider value on config's LoadOptions. If credentials provider is set to nil, the credentials provider value will be ignored. If multiple WithCredentialsProvider calls are made, the last call overrides the previous call values.

Example ¶
func WithCustomCABundle ¶
func WithCustomCABundle(v io.Reader) LoadOptionsFunc
WithCustomCABundle is a helper function to construct functional options that sets CustomCABundle on config's LoadOptions. Setting the custom CA Bundle to nil will result in custom CA Bundle value being ignored. If multiple WithCustomCABundle calls are made, the last call overrides the previous call values.

func WithDefaultRegion ¶
func WithDefaultRegion(v string) LoadOptionsFunc
WithDefaultRegion is a helper function to construct functional options that sets a DefaultRegion on config's LoadOptions. Setting the default region to an empty string, will result in the default region value being ignored. If multiple WithDefaultRegion calls are made, the last call overrides the previous call values. Note that both WithRegion and WithEC2IMDSRegion call takes precedence over WithDefaultRegion call when resolving region.

func WithDefaultsMode ¶
added in v1.13.0
func WithDefaultsMode(mode aws.DefaultsMode, optFns ...func(options *DefaultsModeOptions)) LoadOptionsFunc
WithDefaultsMode sets the SDK defaults configuration mode to the value provided.

Zero or more functional options can be provided to provide configuration options for performing environment discovery when using aws.DefaultsModeAuto.

func WithDisableRequestCompression ¶
added in v1.26.0
func WithDisableRequestCompression(DisableRequestCompression *bool) LoadOptionsFunc
WithDisableRequestCompression is a helper function to construct functional options that sets DisableRequestCompression on config's LoadOptions.

func WithEC2IMDSClientEnableState ¶
added in v1.5.0
func WithEC2IMDSClientEnableState(v imds.ClientEnableState) LoadOptionsFunc
WithEC2IMDSClientEnableState is a helper function to construct functional options that sets the EC2IMDSClientEnableState.

func WithEC2IMDSEndpoint ¶
added in v1.5.0
func WithEC2IMDSEndpoint(v string) LoadOptionsFunc
WithEC2IMDSEndpoint is a helper function to construct functional options that sets the EC2IMDSEndpoint.

func WithEC2IMDSEndpointMode ¶
added in v1.5.0
func WithEC2IMDSEndpointMode(v imds.EndpointModeState) LoadOptionsFunc
WithEC2IMDSEndpointMode is a helper function to construct functional options that sets the EC2IMDSEndpointMode.

func WithEC2IMDSRegion ¶
func WithEC2IMDSRegion(fnOpts ...func(o *UseEC2IMDSRegion)) LoadOptionsFunc
WithEC2IMDSRegion is a helper function to construct functional options that enables resolving EC2IMDS region. The function takes in a UseEC2IMDSRegion functional option, and can be used to set the EC2IMDS client which will be used to resolve EC2IMDSRegion. If no functional option is provided, an EC2IMDS client is built and used by the resolver. If multiple WithEC2IMDSRegion calls are made, the last call overrides the previous call values. Note that the WithRegion calls takes precedence over WithEC2IMDSRegion when resolving region.

Example ¶
func WithEC2RoleCredentialOptions ¶
added in v0.2.0
func WithEC2RoleCredentialOptions(v func(*ec2rolecreds.Options)) LoadOptionsFunc
WithEC2RoleCredentialOptions is a helper function to construct functional options that sets a function to use ec2rolecreds.Options on config's LoadOptions. If EC2 role credential options is set to nil, the EC2 role credential options value will be ignored. If multiple WithEC2RoleCredentialOptions calls are made, the last call overrides the previous call values.

func WithEndpointCredentialOptions ¶
added in v0.2.0
func WithEndpointCredentialOptions(v func(*endpointcreds.Options)) LoadOptionsFunc
WithEndpointCredentialOptions is a helper function to construct functional options that sets a function to use endpointcreds.Options on config's LoadOptions. If endpoint credential options is set to nil, the endpoint credential options value will be ignored. If multiple WithEndpointCredentialOptions calls are made, the last call overrides the previous call values.

func WithEndpointDiscovery ¶
added in v1.4.0
func WithEndpointDiscovery(v aws.EndpointDiscoveryEnableState) LoadOptionsFunc
WithEndpointDiscovery is a helper function to construct functional options that can be used to enable endpoint discovery on LoadOptions for supported clients. If multiple WithEndpointDiscovery calls are made, the last call overrides the previous call values.

func
WithEndpointResolver
deprecated
func
WithEndpointResolverWithOptions
deprecated
added in v1.11.0
func WithHTTPClient ¶
func WithHTTPClient(v HTTPClient) LoadOptionsFunc
WithHTTPClient is a helper function to construct functional options that sets HTTPClient on LoadOptions. If HTTPClient is set to nil, the HTTPClient value will be ignored. If multiple WithHTTPClient calls are made, the last call overrides the previous call values.

Example ¶
func WithLogConfigurationWarnings ¶
added in v0.3.0
func WithLogConfigurationWarnings(v bool) LoadOptionsFunc
WithLogConfigurationWarnings is a helper function to construct functional options that can be used to set LogConfigurationWarnings on LoadOptions.

If multiple WithLogConfigurationWarnings calls are made, the last call overrides the previous call values.

func WithLogger ¶
added in v0.3.0
func WithLogger(v logging.Logger) LoadOptionsFunc
WithLogger is a helper function to construct functional options that sets Logger on LoadOptions. If Logger is set to nil, the Logger value will be ignored. If multiple WithLogger calls are made, the last call overrides the previous call values.

func WithProcessCredentialOptions ¶
func WithProcessCredentialOptions(v func(*processcreds.Options)) LoadOptionsFunc
WithProcessCredentialOptions is a helper function to construct functional options that sets a function to use processcreds.Options on config's LoadOptions. If process credential options is set to nil, the process credential value will be ignored. If multiple WithProcessCredentialOptions calls are made, the last call overrides the previous call values.

func WithRegion ¶
func WithRegion(v string) LoadOptionsFunc
WithRegion is a helper function to construct functional options that sets Region on config's LoadOptions. Setting the region to an empty string, will result in the region value being ignored. If multiple WithRegion calls are made, the last call overrides the previous call values.

Example ¶
func WithRequestChecksumCalculation ¶
added in v1.29.0
func WithRequestChecksumCalculation(c aws.RequestChecksumCalculation) LoadOptionsFunc
WithRequestChecksumCalculation is a helper function to construct functional options that sets RequestChecksumCalculation on config's LoadOptions

func WithRequestMinCompressSizeBytes ¶
added in v1.26.0
func WithRequestMinCompressSizeBytes(RequestMinCompressSizeBytes *int64) LoadOptionsFunc
WithRequestMinCompressSizeBytes is a helper function to construct functional options that sets RequestMinCompressSizeBytes on config's LoadOptions.

func WithResponseChecksumValidation ¶
added in v1.29.0
func WithResponseChecksumValidation(v aws.ResponseChecksumValidation) LoadOptionsFunc
WithResponseChecksumValidation is a helper function to construct functional options that sets ResponseChecksumValidation on config's LoadOptions

func WithRetryMaxAttempts ¶
added in v1.14.0
func WithRetryMaxAttempts(v int) LoadOptionsFunc
WithRetryMaxAttempts is a helper function to construct functional options that sets RetryMaxAttempts on LoadOptions. If RetryMaxAttempts is unset, the RetryMaxAttempts value is ignored. If multiple WithRetryMaxAttempts calls are made, the last call overrides the previous call values.

Will be ignored of LoadOptions.Retryer or WithRetryer are used.

func WithRetryMode ¶
added in v1.14.0
func WithRetryMode(v aws.RetryMode) LoadOptionsFunc
WithRetryMode is a helper function to construct functional options that sets RetryMode on LoadOptions. If RetryMode is unset, the RetryMode value is ignored. If multiple WithRetryMode calls are made, the last call overrides the previous call values.

Will be ignored of LoadOptions.Retryer or WithRetryer are used.

func WithRetryer ¶
added in v0.3.0
func WithRetryer(v func() aws.Retryer) LoadOptionsFunc
WithRetryer is a helper function to construct functional options that sets Retryer on LoadOptions. If Retryer is set to nil, the Retryer value is ignored. If multiple WithRetryer calls are made, the last call overrides the previous call values.

func WithS3DisableExpressAuth ¶
added in v1.25.8
func WithS3DisableExpressAuth(v bool) LoadOptionsFunc
WithS3DisableExpressAuth sets [LoadOptions.S3DisableExpressAuth] to the value provided.

func WithS3DisableMultiRegionAccessPoints ¶
added in v1.18.34
func WithS3DisableMultiRegionAccessPoints(v bool) LoadOptionsFunc
WithS3DisableMultiRegionAccessPoints is a helper function to construct functional options that can be used to set S3DisableMultiRegionAccessPoints on LoadOptions. If multiple WithS3DisableMultiRegionAccessPoints calls are made, the last call overrides the previous call values.

func WithS3UseARNRegion ¶
added in v0.4.0
func WithS3UseARNRegion(v bool) LoadOptionsFunc
WithS3UseARNRegion is a helper function to construct functional options that can be used to set S3UseARNRegion on LoadOptions. If multiple WithS3UseARNRegion calls are made, the last call overrides the previous call values.

func WithSSOProviderOptions ¶
added in v1.1.0
func WithSSOProviderOptions(v func(*ssocreds.Options)) LoadOptionsFunc
WithSSOProviderOptions is a helper function to construct functional options that sets a function to use ssocreds.Options on config's LoadOptions. If the SSO credential provider options is set to nil, the sso provider options value will be ignored. If multiple WithSSOProviderOptions calls are made, the last call overrides the previous call values.

func WithSSOTokenProviderOptions ¶
added in v1.17.2
func WithSSOTokenProviderOptions(v func(*ssocreds.SSOTokenProviderOptions)) LoadOptionsFunc
WithSSOTokenProviderOptions is a helper function to construct functional options that sets a function to modify the SSOtokenProviderOptions the SDK's credentials/ssocreds#SSOProvider will be configured with, if the SSOTokenProvider is used by the configuration loader.

If multiple WithSSOTokenProviderOptions calls are made, the last call overrides the previous call values.

func WithServiceOptions ¶
added in v1.31.0
func WithServiceOptions(callbacks ...func(string, any)) LoadOptionsFunc
WithServiceOptions is a helper function to construct functional options that sets ServiceOptions on config's LoadOptions.

func WithSharedConfigFiles ¶
func WithSharedConfigFiles(v []string) LoadOptionsFunc
WithSharedConfigFiles is a helper function to construct functional options that sets slice of SharedConfigFiles on config's LoadOptions. Setting the shared config files to an nil string slice, will result in the shared config files value being ignored. If multiple WithSharedConfigFiles calls are made, the last call overrides the previous call values.

func WithSharedConfigProfile ¶
func WithSharedConfigProfile(v string) LoadOptionsFunc
WithSharedConfigProfile is a helper function to construct functional options that sets SharedConfigProfile on config's LoadOptions. Setting the shared config profile to an empty string, will result in the shared config profile value being ignored. If multiple WithSharedConfigProfile calls are made, the last call overrides the previous call values.

Example ¶
func WithSharedCredentialsFiles ¶
added in v0.4.0
func WithSharedCredentialsFiles(v []string) LoadOptionsFunc
WithSharedCredentialsFiles is a helper function to construct functional options that sets slice of SharedCredentialsFiles on config's LoadOptions. Setting the shared credentials files to an nil string slice, will result in the shared credentials files value being ignored. If multiple WithSharedCredentialsFiles calls are made, the last call overrides the previous call values.

func WithUseDualStackEndpoint ¶
added in v1.10.0
func WithUseDualStackEndpoint(v aws.DualStackEndpointState) LoadOptionsFunc
WithUseDualStackEndpoint is a helper function to construct functional options that can be used to set UseDualStackEndpoint on LoadOptions.

func WithUseFIPSEndpoint ¶
added in v1.10.0
func WithUseFIPSEndpoint(v aws.FIPSEndpointState) LoadOptionsFunc
WithUseFIPSEndpoint is a helper function to construct functional options that can be used to set UseFIPSEndpoint on LoadOptions.

func WithWebIdentityRoleCredentialOptions ¶
added in v0.2.0
func WithWebIdentityRoleCredentialOptions(v func(*stscreds.WebIdentityRoleOptions)) LoadOptionsFunc
WithWebIdentityRoleCredentialOptions is a helper function to construct functional options that sets a function to use stscreds.WebIdentityRoleOptions on config's LoadOptions. If web identity role credentials options is set to nil, the web identity role credentials value will be ignored. If multiple WithWebIdentityRoleCredentialOptions calls are made, the last call overrides the previous call values.

Example ¶
type LoadSharedConfigOptions ¶
added in v0.4.0
type LoadSharedConfigOptions struct {

	// CredentialsFiles are the shared credentials files
	CredentialsFiles []string

	// ConfigFiles are the shared config files
	ConfigFiles []string

	// Logger is the logger used to log shared config behavior
	Logger logging.Logger
}
LoadSharedConfigOptions struct contains optional values that can be used to load the config.

type SSOSession ¶
added in v1.17.2
type SSOSession struct {
	Name        string
	SSORegion   string
	SSOStartURL string
}
SSOSession provides the shared configuration parameters of the sso-session section.

type Services ¶
added in v1.21.0
type Services struct {
	// Services section values
	// {"serviceId": {"key": "value"}}
	// e.g. {"s3": {"endpoint_url": "example.com"}}
	ServiceValues map[string]map[string]string
}
Services contains values configured in the services section of the AWS configuration file.

type SharedConfig ¶
type SharedConfig struct {
	Profile string

	// Credentials values from the config file. Both aws_access_key_id
	// and aws_secret_access_key must be provided together in the same file
	// to be considered valid. The values will be ignored if not a complete group.
	// aws_session_token is an optional field that can be provided if both of the
	// other two fields are also provided.
	//
	//	aws_access_key_id
	//	aws_secret_access_key
	//	aws_session_token
	Credentials aws.Credentials

	CredentialSource     string
	CredentialProcess    string
	WebIdentityTokenFile string

	// SSO session options
	SSOSessionName string
	SSOSession     *SSOSession

	// Legacy SSO session options
	SSORegion   string
	SSOStartURL string

	// SSO fields not used
	SSOAccountID string
	SSORoleName  string

	RoleARN             string
	ExternalID          string
	MFASerial           string
	RoleSessionName     string
	RoleDurationSeconds *time.Duration

	SourceProfileName string
	Source            *SharedConfig

	// Region is the region the SDK should use for looking up AWS service endpoints
	// and signing requests.
	//
	//	region = us-west-2
	Region string

	// EnableEndpointDiscovery can be enabled or disabled in the shared config
	// by setting endpoint_discovery_enabled to true, or false respectively.
	//
	//	endpoint_discovery_enabled = true
	EnableEndpointDiscovery aws.EndpointDiscoveryEnableState

	// Specifies if the S3 service should allow ARNs to direct the region
	// the client's requests are sent to.
	//
	// s3_use_arn_region=true
	S3UseARNRegion *bool

	// Specifies the EC2 Instance Metadata Service default endpoint selection
	// mode (IPv4 or IPv6)
	//
	// ec2_metadata_service_endpoint_mode=IPv6
	EC2IMDSEndpointMode imds.EndpointModeState

	// Specifies the EC2 Instance Metadata Service endpoint to use. If
	// specified it overrides EC2IMDSEndpointMode.
	//
	// ec2_metadata_service_endpoint=http://fd00:ec2::254
	EC2IMDSEndpoint string

	// Specifies that IMDS clients should not fallback to IMDSv1 if token
	// requests fail.
	//
	// ec2_metadata_v1_disabled=true
	EC2IMDSv1Disabled *bool

	// Specifies if the S3 service should disable support for Multi-Region
	// access-points
	//
	// s3_disable_multiregion_access_points=true
	S3DisableMultiRegionAccessPoints *bool

	// Specifies that SDK clients must resolve a dual-stack endpoint for
	// services.
	//
	// use_dualstack_endpoint=true
	UseDualStackEndpoint aws.DualStackEndpointState

	// Specifies that SDK clients must resolve a FIPS endpoint for
	// services.
	//
	// use_fips_endpoint=true
	UseFIPSEndpoint aws.FIPSEndpointState

	// Specifies which defaults mode should be used by services.
	//
	// defaults_mode=standard
	DefaultsMode aws.DefaultsMode

	// Specifies the maximum number attempts an API client will call an
	// operation that fails with a retryable error.
	//
	// max_attempts=3
	RetryMaxAttempts int

	// Specifies the retry model the API client will be created with.
	//
	// retry_mode=standard
	RetryMode aws.RetryMode

	// Sets the path to a custom Credentials Authority (CA) Bundle PEM file
	// that the SDK will use instead of the system's root CA bundle. Only use
	// this if you want to configure the SDK to use a custom set of CAs.
	//
	// Enabling this option will attempt to merge the Transport into the SDK's
	// HTTP client. If the client's Transport is not a http.Transport an error
	// will be returned. If the Transport's TLS config is set this option will
	// cause the SDK to overwrite the Transport's TLS config's  RootCAs value.
	//
	// Setting a custom HTTPClient in the aws.Config options will override this
	// setting. To use this option and custom HTTP client, the HTTP client
	// needs to be provided when creating the config. Not the service client.
	//
	//  ca_bundle=$HOME/my_custom_ca_bundle
	CustomCABundle string

	// aws sdk app ID that can be added to user agent header string
	AppID string

	// Flag used to disable configured endpoints.
	IgnoreConfiguredEndpoints *bool

	// Value to contain configured endpoints to be propagated to
	// corresponding endpoint resolution field.
	BaseEndpoint string

	// Services section config.
	ServicesSectionName string
	Services            Services

	// determine if request compression is allowed, default to false
	// retrieved from config file's profile field disable_request_compression
	DisableRequestCompression *bool

	// inclusive threshold request body size to trigger compression,
	// default to 10240 and must be within 0 and 10485760 bytes inclusive
	// retrieved from config file's profile field request_min_compression_size_bytes
	RequestMinCompressSizeBytes *int64

	// Whether S3Express auth is disabled.
	//
	// This will NOT prevent requests from being made to S3Express buckets, it
	// will only bypass the modified endpoint routing and signing behaviors
	// associated with the feature.
	S3DisableExpressAuth *bool

	AccountIDEndpointMode aws.AccountIDEndpointMode

	// RequestChecksumCalculation indicates if the request checksum should be calculated
	RequestChecksumCalculation aws.RequestChecksumCalculation

	// ResponseChecksumValidation indicates if the response checksum should be validated
	ResponseChecksumValidation aws.ResponseChecksumValidation

	// Priority list of preferred auth scheme names (e.g. sigv4a).
	AuthSchemePreference []string
}
SharedConfig represents the configuration fields of the SDK config files.

func LoadSharedConfigProfile ¶
added in v0.4.0
func LoadSharedConfigProfile(ctx context.Context, profile string, optFns ...func(*LoadSharedConfigOptions)) (SharedConfig, error)
LoadSharedConfigProfile retrieves the configuration from the list of files using the profile provided. The order the files are listed will determine precedence. Values in subsequent files will overwrite values defined in earlier files.

For example, given two files A and B. Both define credentials. If the order of the files are A then B, B's credential values will be used instead of A's.

If config files are not set, SDK will default to using a file at location `.aws/config` if present. If credentials files are not set, SDK will default to using a file at location `.aws/credentials` if present. No default files are set, if files set to an empty slice.

You can read more about shared config and credentials file location at https://docs.aws.amazon.com/credref/latest/refdocs/file-location.html#file-location

func (SharedConfig) GetEC2IMDSEndpoint ¶
added in v1.5.0
func (c SharedConfig) GetEC2IMDSEndpoint() (string, bool, error)
GetEC2IMDSEndpoint implements a EC2IMDSEndpoint option resolver interface.

func (SharedConfig) GetEC2IMDSEndpointMode ¶
added in v1.5.0
func (c SharedConfig) GetEC2IMDSEndpointMode() (imds.EndpointModeState, bool, error)
GetEC2IMDSEndpointMode implements a EC2IMDSEndpointMode option resolver interface.

func (SharedConfig) GetEC2IMDSV1FallbackDisabled ¶
added in v1.22.0
func (c SharedConfig) GetEC2IMDSV1FallbackDisabled() (bool, bool)
GetEC2IMDSV1FallbackDisabled implements an EC2IMDSV1FallbackDisabled option resolver interface.

func (SharedConfig) GetEnableEndpointDiscovery ¶
func (c SharedConfig) GetEnableEndpointDiscovery(ctx context.Context) (value aws.EndpointDiscoveryEnableState, ok bool, err error)
GetEnableEndpointDiscovery returns if the enable_endpoint_discovery is set.

func (SharedConfig) GetIgnoreConfiguredEndpoints ¶
added in v1.21.0
func (c SharedConfig) GetIgnoreConfiguredEndpoints(context.Context) (bool, bool, error)
GetIgnoreConfiguredEndpoints is used in knowing when to disable configured endpoints feature.

func (SharedConfig) GetRetryMaxAttempts ¶
added in v1.14.0
func (c SharedConfig) GetRetryMaxAttempts(ctx context.Context) (value int, ok bool, err error)
GetRetryMaxAttempts returns the maximum number of attempts an API client created Retryer should attempt an operation call before failing.

func (SharedConfig) GetRetryMode ¶
added in v1.14.0
func (c SharedConfig) GetRetryMode(ctx context.Context) (value aws.RetryMode, ok bool, err error)
GetRetryMode returns the model the API client should create its Retryer in.

func (SharedConfig) GetS3DisableExpressAuth ¶
added in v1.25.7
func (c SharedConfig) GetS3DisableExpressAuth() (value, ok bool)
GetS3DisableExpressAuth returns the configured value for [SharedConfig.S3DisableExpressAuth].

func (SharedConfig) GetS3DisableMultiRegionAccessPoints ¶
added in v1.8.0
func (c SharedConfig) GetS3DisableMultiRegionAccessPoints(ctx context.Context) (value, ok bool, err error)
GetS3DisableMultiRegionAccessPoints returns if the S3 service should disable support for Multi-Region access-points.

func (SharedConfig) GetS3UseARNRegion ¶
func (c SharedConfig) GetS3UseARNRegion(ctx context.Context) (value, ok bool, err error)
GetS3UseARNRegion returns if the S3 service should allow ARNs to direct the region the client's requests are sent to.

func (SharedConfig) GetServiceBaseEndpoint ¶
added in v1.21.0
func (c SharedConfig) GetServiceBaseEndpoint(ctx context.Context, sdkID string) (string, bool, error)
GetServiceBaseEndpoint is used to retrieve a normalized SDK ID for use with configured endpoints.

func (SharedConfig) GetUseDualStackEndpoint ¶
added in v1.10.0
func (c SharedConfig) GetUseDualStackEndpoint(ctx context.Context) (value aws.DualStackEndpointState, found bool, err error)
GetUseDualStackEndpoint returns whether the service's dual-stack endpoint should be used for requests.

func (SharedConfig) GetUseFIPSEndpoint ¶
added in v1.10.0
func (c SharedConfig) GetUseFIPSEndpoint(ctx context.Context) (value aws.FIPSEndpointState, found bool, err error)
GetUseFIPSEndpoint returns whether the service's FIPS endpoint should be used for requests.

type SharedConfigAssumeRoleError ¶
type SharedConfigAssumeRoleError struct {
	Profile string
	RoleARN string
	Err     error
}
SharedConfigAssumeRoleError is an error for the shared config when the profile contains assume role information, but that information is invalid or not complete.

func (SharedConfigAssumeRoleError) Error ¶
func (e SharedConfigAssumeRoleError) Error() string
func (SharedConfigAssumeRoleError) Unwrap ¶
func (e SharedConfigAssumeRoleError) Unwrap() error
Unwrap returns the underlying error that caused the failure.

type SharedConfigLoadError ¶
type SharedConfigLoadError struct {
	Filename string
	Err      error
}
SharedConfigLoadError is an error for the shared config file failed to load.

func (SharedConfigLoadError) Error ¶
func (e SharedConfigLoadError) Error() string
func (SharedConfigLoadError) Unwrap ¶
func (e SharedConfigLoadError) Unwrap() error
Unwrap returns the underlying error that caused the failure.

type SharedConfigProfileNotExistError ¶
type SharedConfigProfileNotExistError struct {
	Filename []string
	Profile  string
	Err      error
}
SharedConfigProfileNotExistError is an error for the shared config when the profile was not find in the config file.

func (SharedConfigProfileNotExistError) Error ¶
func (e SharedConfigProfileNotExistError) Error() string
func (SharedConfigProfileNotExistError) Unwrap ¶
func (e SharedConfigProfileNotExistError) Unwrap() error
Unwrap returns the underlying error that caused the failure.

type UseEC2IMDSRegion ¶
added in v0.4.0
type UseEC2IMDSRegion struct {
	// If unset will default to generic EC2 IMDS client.
	Client *imds.Client
}
UseEC2IMDSRegion provides a regionProvider that retrieves the region from the EC2 Metadata service.

 Source Files ¶
View all Source files
auth_scheme_preference.go
config.go
defaultsmode.go
doc.go
env_config.go
generate.go
go_module_metadata.go
load_options.go
local.go
provider.go
resolve.go
resolve_bearer_token.go
resolve_credentials.go
shared_config.go
Why Go
Use Cases
Case Studies
Get Started
Playground
Tour
Stack Overflow
Help
Packages
Standard Library
Sub-repositories
About Go Packages
About
Download
Blog
Issue Tracker
Release Notes
Brand Guidelines
Code of Conduct
Connect
Twitter
GitHub
Slack
r/golang
Meetup
Golang Weekly
Gopher in flight goggles
Copyright
Terms of Service
Privacy Policy
Report an Issue
System theme
Theme Toggle


Shortcuts Modal

Google logo
go.dev uses cookies from Google to deliver and enhance the quality of its services and to analyze traffic. Learn more.
Okay
