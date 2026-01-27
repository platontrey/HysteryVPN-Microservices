# Hysteria2 SNI Configuration Guide

## Important: Obfuscation vs Masquerade

**⚠️ CRITICAL:** Obfuscation (`obfs`) and Masquerade (`masquerade`) are **mutually exclusive** options. They serve similar purposes of traffic disguise but use different mechanisms and cannot be used together.

## Configuration Options

### Option 1: Masquerade (Default)
Traffic appears as normal HTTPS traffic to a legitimate website.

**When to use:**
- When you want traffic to blend in with normal HTTPS
- When using domains that resolve to legitimate websites
- Default and recommended option for most cases

**Configuration:**
```yaml
# Enable masquerade (default)
masquerade:
  type: proxy
  proxy:
    url: https://www.google.com
    rewriteHost: true

# OBFS MUST BE DISABLED
# obfs: # <-- This section should not exist
```

### Option 2: Obfuscation (Salamander)
Traffic is obfuscated using the Salamander protocol.

**When to use:**
- When proxy masquerade is detected/blocked
- When you need stronger traffic obfuscation
- In highly censored environments

**Configuration:**
```yaml
# Enable obfuscation
obfs:
  type: salamander
  password: "your_strong_password_here"

# MASQUERADE MUST BE DISABLED
# masquerade: # <-- This section should not exist
```

## Web Interface Behavior

The web interface automatically handles mutual exclusion:

1. **Enable Obfuscation Checkbox:**
   - ✅ Enables `obfs` section
   - ❌ Automatically removes `masquerade` section
   - 📝 Shows obfs password field

2. **Disable Obfuscation:**
   - ✅ Restores default `masquerade` configuration
   - ❌ Removes `obfs` section
   - 📝 Shows masquerade settings

## API Behavior

When updating configuration via API:

```json
// Enable obfuscation (masquerade will be automatically removed)
PUT /api/v1/nodes/{id}/config
{
  "obfs": {
    "type": "salamander",
    "password": "secret123"
  }
}

// Enable masquerade (obfs will be automatically removed)
PUT /api/v1/nodes/{id}/config
{
  "masquerade": {
    "type": "proxy",
    "proxy": {
      "url": "https://example.com",
      "rewriteHost": true
    }
  }
}
```

## SNI + Obfuscation/Masquerade

SNI works independently of obfuscation/masquerade:

```yaml
# SNI can be used with EITHER obfs OR masquerade
sni:
  enabled: true
  default: "vpn.example.com"
  domains:
    - domain: "vpn.example.com"
      cert: "/etc/hysteria/sni/vpn.example.com.crt"
      key: "/etc/hysteria/sni/vpn.example.com.key"

# Choose ONLY ONE of the following:

# Option A: Masquerade (recommended)
masquerade:
  type: proxy
  proxy:
    url: https://www.google.com
    rewriteHost: true

# Option B: Obfuscation (alternative)
# obfs:
#   type: salamander
#   password: "your_password"
```

## Validation

The system automatically validates configurations and prevents:

- ❌ Both `obfs` and `masquerade` in the same configuration
- ❌ Invalid obfuscation types (only "salamander" supported)
- ❌ Empty masquerade configurations
- ✅ Automatic correction of conflicting settings

## Troubleshooting

### "Obfuscation and masquerade cannot be enabled simultaneously"
**Cause:** Both sections exist in configuration
**Solution:** Remove one of the sections

### Configuration keeps reverting
**Cause:** Both sections present, system auto-corrects
**Solution:** Choose only one option and remove the other

### Traffic not working after enabling obfuscation
**Cause:** Client configuration not updated with obfs settings
**Solution:** Update client config with matching obfs password

## Best Practices

1. **Start with masquerade** - It's more stable and less detectable
2. **Use obfuscation only when needed** - For highly restricted networks
3. **Always test after switching** - Ensure connectivity works
4. **Keep passwords strong** - Use random strings for obfs passwords
5. **Document your choice** - Know which option is active on each node

## Migration Guide

### From Masquerade to Obfuscation:
1. Disable masquerade in web interface
2. Enable obfuscation checkbox
3. Set strong password
4. Update all client configurations
5. Test connectivity

### From Obfuscation to Masquerade:
1. Disable obfuscation in web interface
2. Masquerade will be automatically restored
3. Update client configurations (remove obfs section)
4. Test connectivity

Remember: **Never enable both simultaneously!**