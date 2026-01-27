// Node Configuration API Routes with SNI Support
// Add these to your API service router

const express = require('express');
const router = express.Router();

// GET /api/v1/nodes/{id}/config
router.get('/nodes/:id/config', async (req, res) => {
  try {
    const { id } = req.params;

    // Get node config via orchestrator gRPC
    const config = await orchestratorClient.getNodeConfig(id);

    res.json({
      success: true,
      data: config
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// PUT /api/v1/nodes/{id}/config
router.put('/nodes/:id/config', async (req, res) => {
  try {
    const { id } = req.params;
    const { hysteriaConfig } = req.body;

    // Validate config (basic validation)
    if (!hysteriaConfig || typeof hysteriaConfig !== 'object') {
      return res.status(400).json({
        success: false,
        error: 'Invalid Hysteria2 configuration'
      });
    }

    // Send config update via orchestrator gRPC
    await orchestratorClient.updateNodeConfig(id, hysteriaConfig);

    res.json({
      success: true,
      message: 'Configuration updated successfully'
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// ===== SNI Configuration Endpoints =====

// GET /api/v1/nodes/{id}/sni
router.get('/nodes/:id/sni', async (req, res) => {
  try {
    const { id } = req.params;

    // Get SNI configuration via orchestrator gRPC
    const sniConfig = await orchestratorClient.getSNIConfig(id);

    res.json({
      success: true,
      data: sniConfig
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// PUT /api/v1/nodes/{id}/sni
router.put('/nodes/:id/sni', async (req, res) => {
  try {
    const { id } = req.params;
    const {
      enabled,
      primaryDomain,
      domains = [],
      autoRenew = true,
      email = ''
    } = req.body;

    // Validate request body
    if (enabled && (!primaryDomain || domains.length === 0)) {
      return res.status(400).json({
        success: false,
        error: 'Primary domain and at least one domain are required when SNI is enabled'
      });
    }

    // Validate domains
    for (const domain of domains) {
      if (!isValidDomain(domain)) {
        return res.status(400).json({
          success: false,
          error: `Invalid domain: ${domain}`
        });
      }
    }

    // Update SNI configuration via orchestrator gRPC
    const result = await orchestratorClient.updateSNIConfig(id, {
      enabled,
      primaryDomain,
      domains,
      autoRenew,
      email
    });

    res.json({
      success: true,
      message: 'SNI configuration updated successfully',
      data: result
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// POST /api/v1/nodes/{id}/sni/domains
router.post('/nodes/:id/sni/domains', async (req, res) => {
  try {
    const { id } = req.params;
    const { domain } = req.body;

    if (!domain) {
      return res.status(400).json({
        success: false,
        error: 'Domain is required'
      });
    }

    if (!isValidDomain(domain)) {
      return res.status(400).json({
        success: false,
        error: `Invalid domain: ${domain}`
      });
    }

    // Add domain via orchestrator gRPC
    await orchestratorClient.addSNIDomain(id, domain);

    res.json({
      success: true,
      message: `Domain ${domain} added successfully`
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// DELETE /api/v1/nodes/{id}/sni/domains/{domain}
router.delete('/nodes/:id/sni/domains/:domain', async (req, res) => {
  try {
    const { id, domain } = req.params;

    // Remove domain via orchestrator gRPC
    await orchestratorClient.removeSNIDomain(id, domain);

    res.json({
      success: true,
      message: `Domain ${domain} removed successfully`
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// GET /api/v1/nodes/{id}/sni/status
router.get('/nodes/:id/sni/status', async (req, res) => {
  try {
    const { id } = req.params;

    // Get SNI status via orchestrator gRPC
    const status = await orchestratorClient.getSNIStatus(id);

    res.json({
      success: true,
      data: status
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// GET /api/v1/nodes/{id}/sni/certificates
router.get('/nodes/:id/sni/certificates', async (req, res) => {
  try {
    const { id } = req.params;

    // Get certificate list via orchestrator gRPC
    const certificates = await orchestratorClient.getCertificates(id);

    res.json({
      success: true,
      data: certificates
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// POST /api/v1/nodes/{id}/sni/certificates/generate
router.post('/nodes/:id/sni/certificates/generate', async (req, res) => {
  try {
    const { id } = req.params;
    const { domains } = req.body;

    if (!domains || !Array.isArray(domains) || domains.length === 0) {
      return res.status(400).json({
        success: false,
        error: 'Domains array is required'
      });
    }

    // Validate domains
    for (const domain of domains) {
      if (!isValidDomain(domain)) {
        return res.status(400).json({
          success: false,
          error: `Invalid domain: ${domain}`
        });
      }
    }

    // Generate self-signed certificates via orchestrator gRPC
    const result = await orchestratorClient.generateCertificates(id, domains);

    res.json({
      success: true,
      message: 'Self-signed certificates generated successfully',
      data: result
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// POST /api/v1/nodes/{id}/sni/certificates/generate-letsencrypt
router.post('/nodes/:id/sni/certificates/generate-letsencrypt', async (req, res) => {
  try {
    const { id } = req.params;
    const { domains, email, preferredChallenge = 'http-01', validateDNS = true } = req.body;

    if (!domains || !Array.isArray(domains) || domains.length === 0) {
      return res.status(400).json({
        success: false,
        error: 'Domains array is required'
      });
    }

    if (!email) {
      return res.status(400).json({
        success: false,
        error: 'Email is required for Let\'s Encrypt certificates'
      });
    }

    // Validate email
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return res.status(400).json({
        success: false,
        error: 'Invalid email address'
      });
    }

    // Validate domains
    for (const domain of domains) {
      if (!isValidDomain(domain)) {
        return res.status(400).json({
          success: false,
          error: `Invalid domain: ${domain}`
        });
      }
    }

    // Generate Let's Encrypt certificates via orchestrator gRPC
    const result = await orchestratorClient.generateLetsEncryptCertificates(id, {
      domains,
      email,
      preferredChallenge,
      validateDNS
    });

    res.json({
      success: true,
      message: 'Let\'s Encrypt certificates generated successfully',
      data: result
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// POST /api/v1/nodes/{id}/sni/certificates/renew
router.post('/nodes/:id/sni/certificates/renew', async (req, res) => {
  try {
    const { id } = req.params;

    // Renew certificates via orchestrator gRPC
    const result = await orchestratorClient.renewCertificates(id);

    res.json({
      success: true,
      message: 'Certificates renewed successfully',
      data: result
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// POST /api/v1/nodes/{id}/sni/certificates/validate
router.post('/nodes/:id/sni/certificates/validate', async (req, res) => {
  try {
    const { id } = req.params;
    const { domains } = req.body;

    if (!domains || !Array.isArray(domains) || domains.length === 0) {
      return res.status(400).json({
        success: false,
        error: 'Domains array is required'
      });
    }

    // Validate domains via orchestrator gRPC
    const result = await orchestratorClient.validateDomains(id, domains);

    res.json({
      success: true,
      message: 'Domain validation completed',
      data: result
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// GET /api/v1/nodes/{id}/sni/letsencrypt/status
router.get('/nodes/:id/sni/letsencrypt/status', async (req, res) => {
  try {
    const { id } = req.params;

    // Get Let's Encrypt status via orchestrator gRPC
    const status = await orchestratorClient.getLetsEncryptStatus(id);

    res.json({
      success: true,
      data: status
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// POST /api/v1/nodes/{id}/sni/certificates/upload
router.post('/nodes/:id/sni/certificates/upload', async (req, res) => {
  try {
    const { id } = req.params;
    const { domain, certContent, keyContent } = req.body;

    if (!domain || !certContent || !keyContent) {
      return res.status(400).json({
        success: false,
        error: 'Domain, certificate content, and key content are required'
      });
    }

    if (!isValidDomain(domain)) {
      return res.status(400).json({
        success: false,
        error: `Invalid domain: ${domain}`
      });
    }

    // Upload certificate via orchestrator gRPC
    const result = await orchestratorClient.uploadCertificate(id, {
      domain,
      certContent,
      keyContent
    });

    res.json({
      success: true,
      message: 'Certificate uploaded successfully',
      data: result
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// Utility function to validate domain format
function isValidDomain(domain) {
  const domainRegex = /^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9](?:\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9])*$/;
  return domainRegex.test(domain) && domain.length <= 253;
}

module.exports = router;