// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"

	jwt "github.com/golang-jwt/jwt"
	"github.com/hashicorp/go-azure-sdk/sdk/auth"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
)

type NewSDKAuthOptions struct {
	AuthType       string
	ClientID       string
	ClientSecret   string
	ClientJWT      string
	ClientCertPath string
	TenantID       string
	SubscriptionID string
}

func BuildResourceManagerAuthorizer(ctx context.Context, authOpts NewSDKAuthOptions, env environments.Environment) (auth.Authorizer, error) {
	authorizer, err := buildAuthorizer(ctx, authOpts, env, env.ResourceManager)
	if err != nil {
		return nil, fmt.Errorf("building Resource Manager authorizer from credentials: %+v", err)
	}
	return authorizer, nil
}

func BuildStorageAuthorizer(ctx context.Context, authOpts NewSDKAuthOptions, env environments.Environment) (auth.Authorizer, error) {
	authorizer, err := buildAuthorizer(ctx, authOpts, env, env.Storage)
	if err != nil {
		return nil, fmt.Errorf("building Storage authorizer from credentials: %+v", err)
	}
	return authorizer, nil
}

func buildAuthorizer(ctx context.Context, authOpts NewSDKAuthOptions, env environments.Environment, api environments.Api) (auth.Authorizer, error) {
	var authConfig auth.Credentials
	switch authOpts.AuthType {
	case AuthTypeDeviceLogin:
		return nil, fmt.Errorf("DeviceLogin is not supported in v2 of the Azure Packer Plugin, however you can use the Azure CLI `az login --use-device-code` to use a device code, and then use CLI authentication")
	case AuthTypeAzureCLI:
		authConfig = auth.Credentials{
			Environment:                       env,
			EnableAuthenticatingUsingAzureCLI: true,
		}
	case AuthTypeMSI:
		authConfig = auth.Credentials{
			Environment:                              env,
			EnableAuthenticatingUsingManagedIdentity: true,
		}
	case AuthTypeClientSecret:
		authConfig = auth.Credentials{
			Environment:                           env,
			EnableAuthenticatingUsingClientSecret: true,
			ClientID:                              authOpts.ClientID,
			ClientSecret:                          authOpts.ClientSecret,
			TenantID:                              authOpts.TenantID,
		}
	case AuthTypeClientCert:
		authConfig = auth.Credentials{
			Environment: env,
			EnableAuthenticatingUsingClientCertificate: true,
			ClientID:                  authOpts.ClientID,
			ClientCertificatePath:     authOpts.ClientCertPath,
			ClientCertificatePassword: "",
		}
	case AuthTypeClientBearerJWT:
		authConfig = auth.Credentials{
			Environment:                   env,
			EnableAuthenticationUsingOIDC: true,
			ClientID:                      authOpts.ClientID,
			TenantID:                      authOpts.TenantID,
			OIDCAssertionToken:            authOpts.ClientJWT,
		}
	default:
		panic("AuthType not set")
	}
	authorizer, err := auth.NewAuthorizerFromCredentials(ctx, authConfig, api)
	if err != nil {
		return nil, err
	}
	return authorizer, nil
}

func GetObjectIdFromToken(token string) (string, error) {
	claims := jwt.MapClaims{}
	var p jwt.Parser

	var err error

	_, _, err = p.ParseUnverified(token, claims)

	if err != nil {
		return "", err
	}
	return claims["oid"].(string), nil
}
