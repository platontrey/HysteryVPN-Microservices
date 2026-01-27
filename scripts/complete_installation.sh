# This script is called at the end of a successful installation
# Run post-installation tasks

# Source the main installation functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/install-hysteriavpn.sh"

# Load configuration variables
if [ -f "/opt/hysteriavpn/install-config.sh" ]; then
    source "/opt/hysteriavpn/install-config.sh"
else
    echo "❌ Install config not found. Run main installation first."
    exit 1
fi

# Run post-installation tasks
post_installation_tasks

echo ""
echo "🎯 HysteriaVPN is now fully operational!"
echo "Visit https://$MASTER_DOMAIN to access your VPN management dashboard."