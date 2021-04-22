package cluster

import (
	"github.com/openshift/assisted-service/internal/cluster/validations"
	"github.com/openshift/assisted-service/internal/dbc"
	"github.com/openshift/assisted-service/internal/gencrypto"
	"github.com/openshift/assisted-service/pkg/auth"
	"github.com/pkg/errors"
)

func AgentToken(c *dbc.Cluster, authType auth.AuthType) (token string, err error) {
	switch authType {
	case auth.TypeRHSSO:
		token, err = cloudPullSecretToken(c.PullSecret)
	case auth.TypeLocal:
		token, err = gencrypto.LocalJWT(c.ID.String())
	case auth.TypeNone:
		token = ""
	default:
		err = errors.Errorf("invalid authentication type %v", authType)
	}
	return
}

func cloudPullSecretToken(pullSecret string) (string, error) {
	creds, err := validations.ParsePullSecret(pullSecret)
	if err != nil {
		return "", err
	}
	r, ok := creds["cloud.openshift.com"]
	if !ok {
		return "", errors.Errorf("Pull secret does not contain auth for cloud.openshift.com")
	}
	return r.AuthRaw, nil
}
