#!/usr/bin/env fish
# Fisher uninstall script

echo "🗑️  Uninstalling fuzz.fish..."

set -l install_dir "$HOME/.local/share/fuzz.fish"

# Remove the installation directory
if test -d "$install_dir"
    rm -rf "$install_dir"
    echo "   Removed directory: $install_dir"
end

echo "✅ fuzz.fish uninstalled"
