import React, { useState, useEffect } from 'react';
import axios from 'axios';

const NodeConfig = ({ nodeId }) => {
  const [config, setConfig] = useState({});
  const [xrayConfig, setXrayConfig] = useState({
    enabled: false,
    protocol: 'vless',
    listenPort: 443,
    vlessUUID: '',
    vlessFlow: 'xtls-rprx-vision',
    realityDest: 'www.example.com:443',
    realityServerNames: ['www.example.com'],
    realityPrivateKey: '',
    realityPublicKey: '',
    realityShortIds: ['abc123']
  });
  const [sniConfig, setSniConfig] = useState({
    enabled: false,
    primaryDomain: '',
    domains: [],
    autoRenew: true,
    email: '',
    letsEncryptEnabled: false,
    preferredChallenge: 'http-01',
    validateDNS: true
  });
  const [certificates, setCertificates] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [activeTab, setActiveTab] = useState('basic');

  useEffect(() => {
    loadConfig();
    loadSNIConfig();
    loadCertificates();
  }, [nodeId]);

  const loadConfig = async () => {
    try {
      const response = await axios.get(`/api/v1/nodes/${nodeId}/config`);
      setConfig(response.data.data);
    } catch (error) {
      setMessage('Failed to load configuration');
    }
  };

  const loadSNIConfig = async () => {
    try {
      const response = await axios.get(`/api/v1/nodes/${nodeId}/sni`);
      setSniConfig(response.data.data);
    } catch (error) {
      console.error('Failed to load SNI configuration:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadCertificates = async () => {
    try {
      const response = await axios.get(`/api/v1/nodes/${nodeId}/sni/certificates`);
      setCertificates(response.data.data || []);
    } catch (error) {
      console.error('Failed to load certificates:', error);
    }
  };

  const saveConfig = async () => {
    setSaving(true);
    try {
      await axios.put(`/api/v1/nodes/${nodeId}/config`, {
        hysteriaConfig: config
      });
      setMessage('Configuration saved successfully');
    } catch (error) {
      setMessage('Failed to save configuration');
    } finally {
      setSaving(false);
    }
  };

  const saveSNIConfig = async () => {
    setSaving(true);
    try {
      await axios.put(`/api/v1/nodes/${nodeId}/sni`, sniConfig);
      setMessage('SNI configuration saved successfully');
      await loadCertificates(); // Reload certificates
    } catch (error) {
      setMessage('Failed to save SNI configuration');
    } finally {
      setSaving(false);
    }
  };

  const updateConfig = (key, value) => {
    setConfig(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const updateSNIConfig = (key, value) => {
    setSniConfig(prev => ({
      ...prev,
      [key]: value
    }));
    
    // Special handling for Let's Encrypt
    if (key === 'letsEncryptEnabled' && value === true) {
      // Enable auto-renewal when Let's Encrypt is enabled
      setSniConfig(prev => ({
        ...prev,
        autoRenew: true
      }));
    }
  };

  const addDomain = () => {
    const newDomain = prompt('Enter domain name:');
    if (newDomain && isValidDomain(newDomain)) {
      setSniConfig(prev => ({
        ...prev,
        domains: [...prev.domains, newDomain.trim()]
      }));
    } else if (newDomain) {
      setMessage('Invalid domain format');
    }
  };

  const removeDomain = (domainToRemove) => {
    setSniConfig(prev => ({
      ...prev,
      domains: prev.domains.filter(domain => domain !== domainToRemove)
    }));
  };

  const setPrimaryDomain = (domain) => {
    setSniConfig(prev => ({
      ...prev,
      primaryDomain: domain
    }));
  };

  const generateCertificates = async () => {
    try {
      setSaving(true);
      
      if (sniConfig.letsEncryptEnabled && !sniConfig.email) {
        setMessage('Email is required for Let\'s Encrypt certificates');
        return;
      }
      
      const endpoint = sniConfig.letsEncryptEnabled 
        ? '/api/v1/nodes/${nodeId}/sni/certificates/generate-letsencrypt'
        : '/api/v1/nodes/${nodeId}/sni/certificates/generate';
        
      await axios.post(endpoint, {
        domains: sniConfig.domains,
        email: sniConfig.email,
        preferredChallenge: sniConfig.preferredChallenge,
        validateDNS: sniConfig.validateDNS
      });
      
      setMessage(sniConfig.letsEncryptEnabled 
        ? 'Let\'s Encrypt certificates generated successfully' 
        : 'Self-signed certificates generated successfully');
      await loadCertificates();
    } catch (error) {
      setMessage(`Failed to generate certificates: ${error.response?.data?.error || error.message}`);
    } finally {
      setSaving(false);
    }
  };

  const isValidDomain = (domain) => {
    const domainRegex = /^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9](?:\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9])*$/;
    return domainRegex.test(domain) && domain.length <= 253;
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString();
  };

  const isExpiringSoon = (expiryDate) => {
    const thirtyDaysFromNow = new Date();
    thirtyDaysFromNow.setDate(thirtyDaysFromNow.getDate() + 30);
    return new Date(expiryDate) < thirtyDaysFromNow;
  };

  if (loading) return <div>Loading configuration...</div>;

  return (
    <div className="node-config">
      <h3>Node Configuration</h3>

      {message && <div className="alert">{message}</div>}

      {/* Tab Navigation */}
      <div className="tab-navigation">
        <button
          className={`tab-btn ${activeTab === 'basic' ? 'active' : ''}`}
          onClick={() => setActiveTab('basic')}
        >
          Basic Configuration
        </button>
        <button
          className={`tab-btn ${activeTab === 'sni' ? 'active' : ''}`}
          onClick={() => setActiveTab('sni')}
        >
          SNI Management
        </button>
        <button
          className={`tab-btn ${activeTab === 'certificates' ? 'active' : ''}`}
          onClick={() => setActiveTab('certificates')}
        >
          Certificates
        </button>
        <button
          className={`tab-btn ${activeTab === 'xray' ? 'active' : ''}`}
          onClick={() => setActiveTab('xray')}
        >
          Xray (VLESS/Reality)
        </button>
        <button
          className={`tab-btn ${activeTab === 'advanced' ? 'active' : ''}`}
          onClick={() => setActiveTab('advanced')}
        >
          Advanced
        </button>
      </div>

      {/* Basic Configuration Tab */}
      {activeTab === 'basic' && (
        <div className="config-form">
          <div className="form-group">
            <label>Listen Port:</label>
            <input
              type="number"
              value={config.listen || ''}
              onChange={(e) => updateConfig('listen', parseInt(e.target.value))}
            />
          </div>

        <div className="form-group">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={!!config.obfs}
              onChange={(e) => {
                if (e.target.checked) {
                  // Enable obfs (salamander) - remove masquerade
                  updateConfig('obfs', {
                    type: 'salamander',
                    password: config['obfs-password'] || ''
                  });
                  delete config.masquerade;
                  setConfig({...config});
                } else {
                  // Disable obfs - restore default masquerade
                  delete config.obfs;
                  updateConfig('masquerade', {
                    type: 'proxy',
                    proxy: {
                      url: 'https://www.google.com',
                      rewriteHost: true
                    }
                  });
                }
              }}
            />
            Enable Obfuscation (Salamander)
          </label>
          <small className="form-help">
            Note: Obfuscation and masquerade are mutually exclusive. Enabling obfuscation will disable masquerade.
          </small>
        </div>

        {config.obfs && (
          <div className="form-group">
            <label>Obfs Password:</label>
            <input
              type="password"
              value={config.obfs.password || ''}
              onChange={(e) => updateConfig('obfs', {
                ...config.obfs,
                password: e.target.value
              })}
              placeholder="Enter obfuscation password"
            />
          </div>
        )}

        {!config.obfs && (
          <div className="form-group">
            <label>Masquerade Settings:</label>
            <div className="masquerade-config">
              <div className="form-group">
                <label>Masquerade Type:</label>
                <select
                  value={config.masquerade?.type || 'proxy'}
                  onChange={(e) => updateConfig('masquerade', {
                    ...config.masquerade,
                    type: e.target.value
                  })}
                >
                  <option value="proxy">Proxy</option>
                  <option value="file">File</option>
                  <option value="string">String</option>
                </select>
              </div>
              
              {config.masquerade?.type === 'proxy' && (
                <div className="form-group">
                  <label>Proxy URL:</label>
                  <input
                    type="url"
                    value={config.masquerade?.proxy?.url || 'https://www.google.com'}
                    onChange={(e) => updateConfig('masquerade', {
                      ...config.masquerade,
                      proxy: {
                        ...config.masquerade.proxy,
                        url: e.target.value
                      }
                    })}
                  />
                  <label className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={config.masquerade?.proxy?.rewriteHost !== false}
                      onChange={(e) => updateConfig('masquerade', {
                        ...config.masquerade,
                        proxy: {
                          ...config.masquerade.proxy,
                          rewriteHost: e.target.checked
                        }
                      })}
                    />
                    Rewrite Host Header
                  </label>
                </div>
              )}
            </div>
          </div>
        )}

          <div className="form-group">
            <label>Auth:</label>
            <input
              type="text"
              value={config.auth || ''}
              onChange={(e) => updateConfig('auth', e.target.value)}
            />
          </div>

          <div className="form-group">
            <label>Auth Password:</label>
            <input
              type="password"
              value={config['auth-password'] || ''}
              onChange={(e) => updateConfig('auth-password', e.target.value)}
            />
          </div>

          <button
            onClick={saveConfig}
            disabled={saving}
            className="save-btn"
          >
            {saving ? 'Saving...' : 'Save Basic Configuration'}
          </button>
        </div>
      )}

      {/* SNI Management Tab */}
      {activeTab === 'sni' && (
        <div className="sni-config">
          <div className="form-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={sniConfig.enabled}
                onChange={(e) => updateSNIConfig('enabled', e.target.checked)}
              />
              Enable SNI Support
            </label>
          </div>

          {sniConfig.enabled && (
            <>
              <div className="form-group">
                <label>Primary Domain:</label>
                <input
                  type="text"
                  value={sniConfig.primaryDomain}
                  onChange={(e) => updateSNIConfig('primaryDomain', e.target.value)}
                  placeholder="example.com"
                />
              </div>

              <div className="form-group">
                <label>SNI Domains:</label>
                <div className="domains-list">
                  {sniConfig.domains.map((domain, index) => (
                    <div key={index} className="domain-item">
                      <span className={domain === sniConfig.primaryDomain ? 'primary-domain' : ''}>
                        {domain}
                        {domain === sniConfig.primaryDomain && ' (Primary)'}
                      </span>
                      <div className="domain-actions">
                        <button
                          onClick={() => setPrimaryDomain(domain)}
                          className="set-primary-btn"
                          title="Set as primary domain"
                        >
                          ⭐
                        </button>
                        <button
                          onClick={() => removeDomain(domain)}
                          className="remove-domain-btn"
                          title="Remove domain"
                        >
                          ❌
                        </button>
                      </div>
                    </div>
                  ))}
                  <button onClick={addDomain} className="add-domain-btn">
                    + Add Domain
                  </button>
                </div>
              </div>

              <div className="form-group">
                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={sniConfig.autoRenew}
                    onChange={(e) => updateSNIConfig('autoRenew', e.target.checked)}
                  />
                  Auto-renew certificates
                </label>
              </div>

              <div className="form-group">
                <label>Email for certificate notifications:</label>
                <input
                  type="email"
                  value={sniConfig.email}
                  onChange={(e) => updateSNIConfig('email', e.target.value)}
                  placeholder="admin@example.com"
                />
              </div>

              <div className="form-group">
                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={sniConfig.letsEncryptEnabled}
                    onChange={(e) => updateSNIConfig('letsEncryptEnabled', e.target.checked)}
                  />
                  Use Let's Encrypt (Auto-SSL certificates)
                </label>
                <small className="form-help">
                  Let's Encrypt provides free, trusted SSL certificates. Requires valid email and DNS configuration.
                </small>
              </div>

              {sniConfig.letsEncryptEnabled && (
                <>
                  <div className="form-group">
                    <label>Preferred Challenge Type:</label>
                    <select
                      value={sniConfig.preferredChallenge}
                      onChange={(e) => updateSNIConfig('preferredChallenge', e.target.value)}
                    >
                      <option value="http-01">HTTP-01 (Requires port 80)</option>
                      <option value="dns-01">DNS-01 (Requires DNS TXT records)</option>
                      <option value="tls-alpn-01">TLS-ALPN-01 (Advanced)</option>
                    </select>
                    <small className="form-help">
                      HTTP-01 is recommended for most cases. DNS-01 provides wildcard support but requires manual DNS setup.
                    </small>
                  </div>

                  <div className="form-group">
                    <label className="checkbox-label">
                      <input
                        type="checkbox"
                        checked={sniConfig.validateDNS}
                        onChange={(e) => updateSNIConfig('validateDNS', e.target.checked)}
                      />
                      Validate DNS before certificate generation
                    </label>
                    <small className="form-help">
                      Ensures domains resolve correctly before requesting certificates. Recommended for production.
                    </small>
                  </div>

                  <div className="letsencrypt-info">
                    <h5>🔒 Let's Encrypt Information</h5>
                    <ul>
                      <li>Free, trusted SSL certificates valid for 90 days</li>
                      <li>Automatic renewal when auto-renew is enabled</li>
                      <li>Requires domain to resolve to this server</li>
                      <li>Port 80 must be accessible for HTTP-01 challenge</li>
                      <li>Certificate generation may take 1-2 minutes</li>
                    </ul>
                  </div>
                </>
              )}

              <div className="sni-actions">
                <button
                  onClick={saveSNIConfig}
                  disabled={saving}
                  className="save-btn"
                >
                  {saving ? 'Saving...' : 'Save SNI Configuration'}
                </button>
                <button
                  onClick={generateCertificates}
                  disabled={saving || sniConfig.domains.length === 0}
                  className={`generate-btn ${sniConfig.letsEncryptEnabled ? 'letsencrypt' : 'self-signed'}`}
                >
                  {sniConfig.letsEncryptEnabled ? '🔒 Generate Let\'s Encrypt Certificates' : '🔐 Generate Self-Signed Certificates'}
                </button>
              </div>
            </>
          )}
        </div>
      )}

      {/* Certificates Tab */}
      {activeTab === 'certificates' && (
        <div className="certificates-section">
          <h4>SSL/TLS Certificates</h4>
          {certificates.length === 0 ? (
            <div className="no-certificates">
              <p>No certificates found. Configure SNI and generate certificates to see them here.</p>
            </div>
          ) : (
            <div className="certificates-list">
              {certificates.map((cert, index) => (
                <div key={index} className={`certificate-item ${isExpiringSoon(cert.not_after) ? 'expiring-soon' : ''}`}>
                  <div className="cert-info">
                    <h5>{cert.domain}</h5>
                    <div className="cert-details">
                      <p><strong>Valid from:</strong> {formatDate(cert.not_before)}</p>
                      <p><strong>Valid until:</strong> {formatDate(cert.not_after)}</p>
                      <p><strong>Type:</strong> {cert.is_self_signed ? 'Self-signed' : 'CA-signed'}</p>
                      {cert.issuer && <p><strong>Issuer:</strong> {cert.issuer}</p>}
                    </div>
                  </div>
                  <div className="cert-status">
                    {isExpiringSoon(cert.not_after) && (
                      <span className="status-warning">⚠️ Expiring Soon</span>
                    )}
                    <span className="status-valid">✅ Valid</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Xray Configuration Tab */}
      {activeTab === 'xray' && (
        <div className="xray-config">
          <div className="form-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={xrayConfig.enabled}
                onChange={(e) => setXrayConfig(prev => ({
                  ...prev,
                  enabled: e.target.checked
                }))}
              />
              Enable Xray (VLESS/Reality support)
            </label>
            <small className="form-help">
              Enable Xray-core for advanced protocols like VLESS and Reality. This provides better censorship resistance.
            </small>
          </div>

          {xrayConfig.enabled && (
            <>
              <div className="form-group">
                <label>Protocol:</label>
                <select
                  value={xrayConfig.protocol}
                  onChange={(e) => setXrayConfig(prev => ({
                    ...prev,
                    protocol: e.target.value
                  }))}
                >
                  <option value="vless">VLESS</option>
                  <option value="vless-reality">VLESS + Reality</option>
                  <option value="vmess">VMess</option>
                  <option value="trojan">Trojan</option>
                  <option value="shadowsocks">Shadowsocks</option>
                </select>
              </div>

              <div className="form-group">
                <label>Listen Port:</label>
                <input
                  type="number"
                  value={xrayConfig.listenPort}
                  onChange={(e) => setXrayConfig(prev => ({
                    ...prev,
                    listenPort: parseInt(e.target.value)
                  }))}
                  min="1"
                  max="65535"
                />
              </div>

              {(xrayConfig.protocol === 'vless' || xrayConfig.protocol === 'reality') && (
                <div className="protocol-config">
                  <h4>VLESS Configuration</h4>
                  <div className="form-group">
                    <label>UUID:</label>
                    <input
                      type="text"
                      value={xrayConfig.vlessUUID}
                      onChange={(e) => setXrayConfig(prev => ({
                        ...prev,
                        vlessUUID: e.target.value
                      }))}
                      placeholder="550e8400-e29b-41d4-a716-446655440000"
                    />
                    <small className="form-help">
                      Unique user identifier. Generate a new UUID for each user.
                    </small>
                  </div>

                  <div className="form-group">
                    <label>Flow:</label>
                    <select
                      value={xrayConfig.vlessFlow}
                      onChange={(e) => setXrayConfig(prev => ({
                        ...prev,
                        vlessFlow: e.target.value
                      }))}
                    >
                      <option value="">None</option>
                      <option value="xtls-rprx-vision">XTLS Vision</option>
                      <option value="xtls-rprx-vision-udp443">XTLS Vision UDP443</option>
                    </select>
                  </div>
                </div>
              )}

              {xrayConfig.protocol === 'vless-reality' && (
                <div className="protocol-config">
                  <h4>Reality Configuration</h4>
                  <div className="warning-box">
                    Reality provides advanced anti-detection by mimicking legitimate HTTPS traffic.
                  </div>

                  <div className="form-group">
                    <label>Destination:</label>
                    <input
                      type="text"
                      value={xrayConfig.realityDest}
                      onChange={(e) => setXrayConfig(prev => ({
                        ...prev,
                        realityDest: e.target.value
                      }))}
                      placeholder="www.example.com:443"
                    />
                    <small className="form-help">
                      Target domain and port to mimic (must be a real, accessible HTTPS site).
                    </small>
                  </div>

                  <div className="form-group">
                    <label>Server Names:</label>
                    <input
                      type="text"
                      value={xrayConfig.realityServerNames.join(', ')}
                      onChange={(e) => setXrayConfig(prev => ({
                        ...prev,
                        realityServerNames: e.target.value.split(',').map(s => s.trim()).filter(s => s)
                      }))}
                      placeholder="www.example.com, example.com"
                    />
                    <small className="form-help">
                      Comma-separated list of domain names to spoof in TLS handshake.
                    </small>
                  </div>

                  <div className="form-group">
                    <label>Private Key:</label>
                    <input
                      type="text"
                      value={xrayConfig.realityPrivateKey}
                      onChange={(e) => setXrayConfig(prev => ({
                        ...prev,
                        realityPrivateKey: e.target.value
                      }))}
                      placeholder="X25519 private key (base64)"
                    />
                    <button
                      type="button"
                      onClick={() => {
                        // Generate new keys (would call API)
                        setMessage('Key generation not implemented in UI yet');
                      }}
                      className="generate-keys-btn"
                    >
                      🔑 Generate Keys
                    </button>
                  </div>

                  <div className="form-group">
                    <label>Short IDs:</label>
                    <input
                      type="text"
                      value={xrayConfig.realityShortIds.join(', ')}
                      onChange={(e) => setXrayConfig(prev => ({
                        ...prev,
                        realityShortIds: e.target.value.split(',').map(s => s.trim()).filter(s => s)
                      }))}
                      placeholder="abc123, def456"
                    />
                    <small className="form-help">
                      Comma-separated list of short IDs for additional obfuscation.
                    </small>
                  </div>
                </div>
              )}

              <div className="xray-actions">
                <button
                  onClick={() => setMessage('Xray configuration not implemented in UI yet')}
                  disabled={saving}
                  className="save-btn"
                >
                  {saving ? 'Saving...' : 'Save Xray Configuration'}
                </button>
                <button
                  onClick={() => setMessage('Xray config generation not implemented in UI yet')}
                  className="generate-btn"
                >
                  📄 Generate Config File
                </button>
              </div>
            </>
          )}
        </div>
      )}

      {/* Advanced Tab */}
      {activeTab === 'advanced' && (
        <div className="advanced-config">
          <div className="form-group">
            <label>QUIC Configuration:</label>
            <textarea
              value={JSON.stringify(config.quic || {}, null, 2)}
              onChange={(e) => {
                try {
                  updateConfig('quic', JSON.parse(e.target.value));
                } catch {}
              }}
              rows="6"
            />
          </div>

          <div className="form-group">
            <label>Bandwidth Configuration:</label>
            <textarea
              value={JSON.stringify(config.bandwidth || {}, null, 2)}
              onChange={(e) => {
                try {
                  updateConfig('bandwidth', JSON.parse(e.target.value));
                } catch {}
              }}
              rows="6"
            />
          </div>

          <div className="raw-config">
            <h4>Raw JSON Config:</h4>
            <textarea
              value={JSON.stringify(config, null, 2)}
              onChange={(e) => {
                try {
                  setConfig(JSON.parse(e.target.value));
                } catch {}
              }}
              rows="20"
            />
          </div>

          <button
            onClick={saveConfig}
            disabled={saving}
            className="save-btn"
          >
            {saving ? 'Saving...' : 'Save Advanced Configuration'}
          </button>
        </div>
      )}
    </div>
  );
};

export default NodeConfig;

// CSS styles (you can move this to a separate CSS file)
const styles = `
.node-config {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.alert {
  background: #f8d7da;
  color: #721c24;
  padding: 12px;
  border-radius: 4px;
  margin-bottom: 20px;
  border: 1px solid #f5c6cb;
}

.tab-navigation {
  display: flex;
  border-bottom: 1px solid #ddd;
  margin-bottom: 20px;
}

.tab-btn {
  background: none;
  border: none;
  padding: 12px 20px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  font-size: 14px;
  transition: all 0.3s ease;
}

.tab-btn:hover {
  background: #f8f9fa;
}

.tab-btn.active {
  border-bottom-color: #007bff;
  color: #007bff;
  font-weight: bold;
}

.config-form, .sni-config, .certificates-section, .advanced-config {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
  color: #333;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-group textarea {
  font-family: 'Courier New', monospace;
  resize: vertical;
}

.checkbox-label {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: auto;
  margin-right: 8px;
}

.domains-list {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 10px;
  max-height: 200px;
  overflow-y: auto;
}

.domain-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

.domain-item:last-child {
  border-bottom: none;
}

.primary-domain {
  color: #007bff;
  font-weight: bold;
}

.domain-actions {
  display: flex;
  gap: 5px;
}

.set-primary-btn,
.remove-domain-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 3px;
  font-size: 12px;
}

.set-primary-btn:hover {
  background: #e3f2fd;
}

.remove-domain-btn:hover {
  background: #ffebee;
}

.add-domain-btn {
  width: 100%;
  padding: 10px;
  background: #f8f9fa;
  border: 2px dashed #ddd;
  border-radius: 4px;
  cursor: pointer;
  color: #666;
  transition: all 0.3s ease;
}

.add-domain-btn:hover {
  background: #e9ecef;
  border-color: #007bff;
  color: #007bff;
}

.sni-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.save-btn,
.generate-btn {
  padding: 12px 24px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
  transition: all 0.3s ease;
}

.save-btn {
  background: #007bff;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background: #0056b3;
}

.generate-btn {
  background: #28a745;
  color: white;
}

.generate-btn:hover:not(:disabled) {
  background: #1e7e34;
}

.save-btn:disabled,
.generate-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.no-certificates {
  text-align: center;
  padding: 40px;
  color: #666;
  background: #f8f9fa;
  border-radius: 8px;
}

.certificates-list {
  display: grid;
  gap: 15px;
}

.certificate-item {
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
}

.certificate-item.expiring-soon {
  border-color: #ffc107;
  background: #fff3cd;
}

.cert-info h5 {
  margin: 0 0 10px 0;
  color: #333;
  font-size: 16px;
}

.cert-details p {
  margin: 5px 0;
  font-size: 14px;
  color: #666;
}

.cert-status {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 5px;
}

.status-warning {
  color: #856404;
  font-weight: bold;
  font-size: 12px;
}

.status-valid {
  color: #155724;
  font-weight: bold;
}

.raw-config {
  margin-top: 20px;
}

.raw-config h4 {
  margin-bottom: 10px;
  color: #333;
}

.raw-config textarea {
  width: 100%;
  min-height: 300px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  background: #f8f9fa;
}

.form-help {
  display: block;
  margin-top: 5px;
  font-size: 12px;
  color: #666;
  font-style: italic;
}

.masquerade-config {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 15px;
  background: #f8f9fa;
  margin-top: 10px;
}

.masquerade-config .form-group {
  margin-bottom: 15px;
}

.masquerade-config .form-group:last-child {
  margin-bottom: 0;
}

.masquerade-config select,
.masquerade-config input {
  background: white;
}

.warning-box {
  background: #fff3cd;
  border: 1px solid #ffeaa7;
  border-radius: 4px;
  padding: 10px;
  margin: 10px 0;
  color: #856404;
  font-size: 13px;
}

.warning-box::before {
  content: "⚠️ ";
  font-weight: bold;
}

.letsencrypt-info {
  background: #e8f5e8;
  border: 1px solid #c3e6cb;
  border-radius: 4px;
  padding: 15px;
  margin: 15px 0;
}

.letsencrypt-info h5 {
  margin: 0 0 10px 0;
  color: #155724;
  font-size: 14px;
}

.letsencrypt-info ul {
  margin: 0;
  padding-left: 20px;
  color: #155724;
}

.letsencrypt-info li {
  margin-bottom: 5px;
  font-size: 13px;
}

.generate-btn.letsencrypt {
  background: #28a745;
  border-color: #28a745;
}

.generate-btn.letsencrypt:hover:not(:disabled) {
  background: #218838;
  border-color: #1e7e34;
}

.generate-btn.self-signed {
  background: #6c757d;
  border-color: #6c757d;
}

.generate-btn.self-signed:hover:not(:disabled) {
  background: #5a6268;
  border-color: #545b62;
}

.domain-status {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  margin-left: 10px;
}

.domain-status.valid {
  color: #28a745;
}

.domain-status.invalid {
  color: #dc3545;
}

.domain-status.pending {
  color: #ffc107;
}

.cert-type-badge {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: bold;
  margin-left: 8px;
}

.cert-type-badge.letsencrypt {
  background: #d4edda;
  color: #155724;
}

.cert-type-badge.self-signed {
  background: #f8d7da;
  color: #721c24;
}

.cert-type-badge.expiring {
  background: #fff3cd;
  color: #856404;
}
`;