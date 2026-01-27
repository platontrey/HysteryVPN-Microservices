# Directory for configuration files
mkdir -p generated-configs/certs/ca
mkdir -p generated-configs/certs/services
mkdir -p generated-configs/envs
mkdir -p /opt/hysteriavpn/client-configs
mkdir -p /opt/hysteriavpn/log
mkdir -p /opt/hysteriavpn/backup

# Set proper permissions
chmod 755 generated-configs
chmod 700 generated-configs/certs
chmod 755 generated-configs/envs
chmod 755 /opt/hysteriavpn