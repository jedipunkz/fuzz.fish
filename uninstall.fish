#!/usr/bin/env fish
# Fisher uninstall script

echo "🗑️  Uninstalling fuzz.fish..."

set -l plugin_dir (dirname (status -f))
set -l bin_path "$plugin_dir/bin/fhv"

# Remove the binary
if test -f "$bin_path"
    rm -f "$bin_path"
    echo "   Removed binary: $bin_path"
end

# Remove bin directory if empty
if test -d "$plugin_dir/bin"; and not test (count "$plugin_dir/bin"/*) -gt 0
    rmdir "$plugin_dir/bin" 2>/dev/null
end

echo "✅ fuzz.fish uninstalled"
