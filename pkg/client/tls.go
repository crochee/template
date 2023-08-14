package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"template/pkg/utils"
)

type Config struct {
	Ca   utils.FileOrContent `json:"ca" yaml:"ca"`
	Cert utils.FileOrContent `json:"cert" yaml:"cert"`
	Key  utils.FileOrContent `json:"key" yaml:"key"`
}

// TlsConfig output tls
func TlsConfig(clientAuth tls.ClientAuthType, cfg Config) (*tls.Config, error) {
	caPEMBlock, err := cfg.Ca.Read()
	if err != nil {
		return nil, err
	}
	var certPEMBlock []byte
	if certPEMBlock, err = cfg.Cert.Read(); err != nil {
		return nil, err
	}
	var keyPEMBlock []byte
	if keyPEMBlock, err = cfg.Key.Read(); err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEMBlock) {
		return nil, errors.New("failed to parse root certificate")
	}
	var certificate tls.Certificate
	if certificate, err = tls.X509KeyPair(certPEMBlock, keyPEMBlock); err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   clientAuth, // 服务端认证客户端
		ClientCAs:    pool,       // 服务端认证客户端
		CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		MinVersion:   tls.VersionTLS12,
		RootCAs:      pool, // 客户端认证服务端
	}, nil
}
