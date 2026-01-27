package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"hysteria2_microservices/agent-service/internal/config"
)

// CertificateManager handles SSL/TLS certificate management for SNI domains
type CertificateManager interface {
	// GenerateSelfSignedCert creates a self-signed certificate for a domain
	GenerateSelfSignedCert(domain string) (certPath, keyPath string, err error)

	// InstallCertificate saves certificate and key files
	InstallCertificate(domain, certContent, keyContent string) (certPath, keyPath string, err error)

	// ValidateCertificate checks if a certificate is valid and not expired
	ValidateCertificate(domain string) (bool, error)

	// GetCertificatePaths returns the paths to certificate files for a domain
	GetCertificatePaths(domain string) (certPath, keyPath string, err error)

	// DeleteCertificate removes certificate files for a domain
	DeleteCertificate(domain string) error

	// ListCertificates returns all available certificates
	ListCertificates() ([]CertificateInfo, error)

	// IsCertificateExpiringSoon checks if certificate expires within specified days
	IsCertificateExpiringSoon(domain string, days int) (bool, error)

	// SetupCertDirectory creates the certificate directory structure
	SetupCertDirectory() error

	// Let's Encrypt methods
	GenerateLetsEncryptCert(domain, email string, preferredChallenge string) (certPath, keyPath string, err error)
	ValidateDomainOwnership(domain string) (bool, error)
	AutoRenewCertificates() error
	CheckDNSResolution(domain string) (bool, error)
	InstallCertbot() error
	IsLetsEncryptEnabled() bool
}

// CertificateInfo holds information about a certificate
type CertificateInfo struct {
	Domain       string    `json:"domain"`
	CertPath     string    `json:"cert_path"`
	KeyPath      string    `json:"key_path"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	IsSelfSigned bool      `json:"is_self_signed"`
	Issuer       string    `json:"issuer"`
}

type CertificateManagerImpl struct {
	logger      *logrus.Logger
	config      *config.Config
	certDir     string
	acmeClient  *acme.Client
	certbotPath string
}

// NewCertificateManager creates a new CertificateManager
func NewCertificateManager(logger *logrus.Logger, cfg *config.Config) CertificateManager {
	certDir := cfg.Hysteria2.SNICertPath
	if certDir == "" {
		certDir = "/etc/hysteria/sni"
	}

	certbotPath := "/usr/bin/certbot"
	if _, err := os.Stat(certbotPath); os.IsNotExist(err) {
		certbotPath = "/usr/local/bin/certbot"
	}

	return &CertificateManagerImpl{
		logger:      logger,
		config:      cfg,
		certDir:     certDir,
		certbotPath: certbotPath,
	}
}

// SetupCertDirectory creates the certificate directory structure
func (cm *CertificateManagerImpl) SetupCertDirectory() error {
	if err := os.MkdirAll(cm.certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}
	cm.logger.Infof("Certificate directory created/verified: %s", cm.certDir)
	return nil
}

// GenerateSelfSignedCert creates a self-signed certificate for a domain
func (cm *CertificateManagerImpl) GenerateSelfSignedCert(domain string) (certPath, keyPath string, err error) {
	cm.logger.Infof("Generating self-signed certificate for domain: %s", domain)

	// Ensure certificate directory exists
	if err := cm.SetupCertDirectory(); err != nil {
		return "", "", err
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Prepare certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   domain,
			Organization: []string{"HysteryVPN SNI"},
		},
		DNSNames:              []string{domain},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	// Save certificate file
	certPath = filepath.Join(cm.certDir, fmt.Sprintf("%s.crt", domain))
	certFile, err := os.Create(certPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return "", "", fmt.Errorf("failed to encode certificate: %w", err)
	}

	// Save private key file
	keyPath = filepath.Join(cm.config.Hysteria2.SNIKeyPath, fmt.Sprintf("%s.key", domain))
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyFile.Close()

	keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return "", "", fmt.Errorf("failed to encode private key: %w", err)
	}

	cm.logger.Infof("Certificate generated successfully: %s", domain)
	return certPath, keyPath, nil
}

// InstallCertificate saves certificate and key files
func (cm *CertificateManagerImpl) InstallCertificate(domain, certContent, keyContent string) (certPath, keyPath string, err error) {
	cm.logger.Infof("Installing certificate for domain: %s", domain)

	// Ensure certificate directory exists
	if err := cm.SetupCertDirectory(); err != nil {
		return "", "", err
	}

	// Save certificate
	certPath = filepath.Join(cm.certDir, fmt.Sprintf("%s.crt", domain))
	if err := os.WriteFile(certPath, []byte(certContent), 0644); err != nil {
		return "", "", fmt.Errorf("failed to save certificate: %w", err)
	}

	// Save private key
	keyPath = filepath.Join(cm.config.Hysteria2.SNIKeyPath, fmt.Sprintf("%s.key", domain))
	if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
		return "", "", fmt.Errorf("failed to save private key: %w", err)
	}

	// Validate the installed certificate
	if valid, err := cm.ValidateCertificate(domain); !valid || err != nil {
		if err != nil {
			return "", "", fmt.Errorf("certificate validation failed: %w", err)
		}
		return "", "", fmt.Errorf("installed certificate is invalid")
	}

	cm.logger.Infof("Certificate installed successfully: %s", domain)
	return certPath, keyPath, nil
}

// ValidateCertificate checks if a certificate is valid and not expired
func (cm *CertificateManagerImpl) ValidateCertificate(domain string) (bool, error) {
	certPath, keyPath, err := cm.GetCertificatePaths(domain)
	if err != nil {
		return false, fmt.Errorf("failed to get certificate paths: %w", err)
	}

	// Check if files exist
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return false, fmt.Errorf("certificate file does not exist: %s", certPath)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return false, fmt.Errorf("private key file does not exist: %s", keyPath)
	}

	// Parse certificate
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return false, fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return false, fmt.Errorf("failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check if certificate is expired
	if time.Now().After(cert.NotAfter) {
		return false, fmt.Errorf("certificate has expired")
	}

	// Check if certificate is for the correct domain
	if cert.Subject.CommonName != domain {
		domainMatch := false
		for _, dnsName := range cert.DNSNames {
			if dnsName == domain {
				domainMatch = true
				break
			}
		}
		if !domainMatch {
			return false, fmt.Errorf("certificate is not for domain %s", domain)
		}
	}

	return true, nil
}

// GetCertificatePaths returns the paths to certificate files for a domain
func (cm *CertificateManagerImpl) GetCertificatePaths(domain string) (certPath, keyPath string, err error) {
	certPath = filepath.Join(cm.certDir, fmt.Sprintf("%s.crt", domain))
	keyPath = filepath.Join(cm.config.Hysteria2.SNIKeyPath, fmt.Sprintf("%s.key", domain))
	return certPath, keyPath, nil
}

// DeleteCertificate removes certificate files for a domain
func (cm *CertificateManagerImpl) DeleteCertificate(domain string) error {
	cm.logger.Infof("Deleting certificate for domain: %s", domain)

	certPath, keyPath, err := cm.GetCertificatePaths(domain)
	if err != nil {
		return err
	}

	// Remove certificate file
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove certificate file: %w", err)
	}

	// Remove private key file
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove private key file: %w", err)
	}

	cm.logger.Infof("Certificate deleted successfully: %s", domain)
	return nil
}

// ListCertificates returns all available certificates
func (cm *CertificateManagerImpl) ListCertificates() ([]CertificateInfo, error) {
	var certificates []CertificateInfo

	files, err := os.ReadDir(cm.certDir)
	if err != nil {
		if os.IsNotExist(err) {
			return certificates, nil
		}
		return nil, fmt.Errorf("failed to read certificate directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".crt") {
			continue
		}

		domain := strings.TrimSuffix(file.Name(), ".crt")
		certPath, _, err := cm.GetCertificatePaths(domain)
		if err != nil {
			continue
		}

		// Parse certificate to get details
		certData, err := os.ReadFile(certPath)
		if err != nil {
			continue
		}

		block, _ := pem.Decode(certData)
		if block == nil {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		_, keyPath, err := cm.GetCertificatePaths(domain)
		if err != nil {
			continue
		}

		certificates = append(certificates, CertificateInfo{
			Domain:       domain,
			CertPath:     certPath,
			KeyPath:      keyPath,
			NotBefore:    cert.NotBefore,
			NotAfter:     cert.NotAfter,
			IsSelfSigned: cert.Issuer.CommonName == cert.Subject.CommonName,
			Issuer:       cert.Issuer.CommonName,
		})
	}

	return certificates, nil
}

// IsCertificateExpiringSoon checks if certificate expires within specified days
func (cm *CertificateManagerImpl) IsCertificateExpiringSoon(domain string, days int) (bool, error) {
	certPath, _, err := cm.GetCertificatePaths(domain)
	if err != nil {
		return false, err
	}

	certData, err := os.ReadFile(certPath)
	if err != nil {
		return false, err
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return false, fmt.Errorf("failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, err
	}

	expiryThreshold := time.Now().AddDate(0, 0, days)
	return cert.NotAfter.Before(expiryThreshold), nil
}

// ===== LET'S ENCRYPT METHODS =====

// IsLetsEncryptEnabled checks if Let's Encrypt tools are available
func (cm *CertificateManagerImpl) IsLetsEncryptEnabled() bool {
	_, err := os.Stat(cm.certbotPath)
	return err == nil
}

// InstallCertbot installs certbot if not available
func (cm *CertificateManagerImpl) InstallCertbot() error {
	if cm.IsLetsEncryptEnabled() {
		cm.logger.Info("Certbot already installed")
		return nil
	}

	cm.logger.Info("Installing certbot...")

	// Try different installation methods based on OS
	commands := [][]string{
		{"apt-get", "update", "-y"},
		{"apt-get", "install", "certbot", "-y"},
		{"yum", "install", "-y", "certbot"},
		{"dnf", "install", "-y", "certbot"},
	}

	for i := 0; i < len(commands); i += 2 {
		cmd := exec.Command(commands[i][0], commands[i][1:]...)
		if err := cmd.Run(); err != nil {
			cm.logger.Warnf("Failed to run %s: %v", commands[i][0], err)
			continue
		}
	}

	// Check if certbot was installed
	if !cm.IsLetsEncryptEnabled() {
		return fmt.Errorf("failed to install certbot")
	}

	cm.logger.Info("Certbot installed successfully")
	return nil
}

// CheckDNSResolution validates that domain resolves to this server
func (cm *CertificateManagerImpl) CheckDNSResolution(domain string) (bool, error) {
	cm.logger.Infof("Checking DNS resolution for domain: %s", domain)

	// Get server's public IP
	publicIP, err := cm.getPublicIP()
	if err != nil {
		cm.logger.Warnf("Failed to get public IP: %v", err)
		return false, err
	}

	// Resolve domain to IPs
	ips, err := net.LookupIP(domain)
	if err != nil {
		cm.logger.Warnf("DNS resolution failed for %s: %v", domain, err)
		return false, err
	}

	// Check if any IP matches our public IP
	for _, ip := range ips {
		if ip.String() == publicIP {
			cm.logger.Infof("DNS resolution successful: %s -> %s", domain, publicIP)
			return true, nil
		}
	}

	cm.logger.Warnf("Domain %s does not resolve to server IP %s", domain, publicIP)
	return false, fmt.Errorf("domain does not resolve to server IP")
}

// getPublicIP gets the server's public IP address
func (cm *CertificateManagerImpl) getPublicIP() (string, error) {
	// Try multiple services to get public IP
	services := []string{
		"https://api.ipify.org",
		"https://ipinfo.io/ip",
		"https://icanhazip.com",
	}

	for _, service := range services {
		resp, err := http.Get(service)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			return strings.TrimSpace(string(body)), nil
		}
	}

	return "", fmt.Errorf("failed to get public IP")
}

// ValidateDomainOwnership performs domain validation before certificate generation
func (cm *CertificateManagerImpl) ValidateDomainOwnership(domain string) (bool, error) {
	cm.logger.Infof("Validating domain ownership for: %s", domain)

	// Check DNS resolution first
	if valid, err := cm.CheckDNSResolution(domain); !valid {
		return false, err
	}

	// Additional validation could include:
	// - HTTP file challenge verification
	// - DNS TXT record verification
	// - Port 80/443 accessibility check

	return true, nil
}

// GenerateLetsEncryptCert generates a Let's Encrypt certificate for the domain
func (cm *CertificateManagerImpl) GenerateLetsEncryptCert(domain, email string, preferredChallenge string) (certPath, keyPath string, err error) {
	cm.logger.Infof("Generating Let's Encrypt certificate for domain: %s", domain)

	// Ensure certbot is installed
	if err := cm.InstallCertbot(); err != nil {
		return "", "", fmt.Errorf("failed to install certbot: %w", err)
	}

	// Validate domain ownership
	if valid, err := cm.ValidateDomainOwnership(domain); !valid {
		return "", "", fmt.Errorf("domain validation failed: %w", err)
	}

	// Ensure certificate directory exists
	if err := cm.SetupCertDirectory(); err != nil {
		return "", "", err
	}

	// Prepare certbot command
	certPath, keyPath = cm.getCertificatePaths(domain)

	args := []string{
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"--email", email,
		"--domains", domain,
		"--standalone",
		"--cert-name", domain,
		"--key-path", keyPath,
		"--fullchain-path", certPath,
	}

	// Add preferred challenge if specified
	if preferredChallenge != "" {
		args = append(args, "--preferred-challenges", preferredChallenge)
	}

	// Run certbot
	cmd := exec.Command(cm.certbotPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		cm.logger.Errorf("Certbot failed: %v, output: %s", err, string(output))
		return "", "", fmt.Errorf("certbot failed: %w", err)
	}

	cm.logger.Infof("Let's Encrypt certificate generated successfully for: %s", domain)
	return certPath, keyPath, nil
}

// AutoRenewCertificates checks and renews all expiring certificates
func (cm *CertificateManagerImpl) AutoRenewCertificates() error {
	cm.logger.Info("Checking for certificates to renew...")

	if !cm.IsLetsEncryptEnabled() {
		cm.logger.Info("Let's Encrypt not available, skipping auto-renewal")
		return nil
	}

	// Get list of certificates
	certificates, err := cm.ListCertificates()
	if err != nil {
		return fmt.Errorf("failed to list certificates: %w", err)
	}

	renewedCount := 0
	for _, cert := range certificates {
		// Check if certificate is expiring within 30 days
		if cert.IsSelfSigned {
			continue // Skip self-signed certificates
		}

		if expiring, err := cm.IsCertificateExpiringSoon(cert.Domain, 30); err == nil && expiring {
			cm.logger.Infof("Renewing certificate for domain: %s", cert.Domain)

			// Use certbot to renew
			cmd := exec.Command(cm.certbotPath, "renew", "--cert-name", cert.Domain, "--non-interactive")
			output, err := cmd.CombinedOutput()
			if err != nil {
				cm.logger.Errorf("Failed to renew certificate for %s: %v, output: %s", cert.Domain, err, string(output))
				continue
			}

			renewedCount++
			cm.logger.Infof("Certificate renewed successfully for: %s", cert.Domain)
		}
	}

	cm.logger.Infof("Auto-renewal completed. Renewed %d certificates", renewedCount)
	return nil
}

// getCertificatePaths returns full paths for certificate files
func (cm *CertificateManagerImpl) getCertificatePaths(domain string) (certPath, keyPath string) {
	certPath = filepath.Join(cm.certDir, fmt.Sprintf("%s.fullchain.pem", domain))
	keyPath = filepath.Join(cm.config.Hysteria2.SNIKeyPath, fmt.Sprintf("%s.privkey.pem", domain))
	return certPath, keyPath
}
