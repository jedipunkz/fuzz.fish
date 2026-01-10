#!/usr/bin/env fish
# Fisher uninstall script

echo "🗑️  Uninstalling fuzz.fish..."

if set -q __fish_config_dir
    set -g _fuzz_fish_functions_dir "$__fish_config_dir/functions"
else
    set -g _fuzz_fish_functions_dir "$HOME/.config/fish/functions"
end

set -l bin_path "$_fuzz_fish_functions_dir/fhv"

# Remove the binary
if test -f "$bin_path"
    rm -f "$bin_path"
    echo "   Removed binary: $bin_path"
end

echo "✅ fuzz.fish uninstalled"
