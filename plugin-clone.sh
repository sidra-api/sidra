#!/bin/bash
mkdir -p ./plugins || true
rm -rf ./plugins/* || true
# Array of repository URLs
repos=(
    "https://github.com/sidra-api/plugin-basic-auth.git"
    "https://github.com/sidra-api/plugin-jwt.git"
    "https://github.com/sidra-api/plugin-rate-limit.git"
    "https://github.com/sidra-api/plugin-whitelist.git"
    "https://github.com/sidra-api/plugin-cache.git"
    "https://github.com/sidra-api/plugin-rsa.git"
    "https://github.com/sidra-api/plugin-azure-jwt.git"
)

for repo in "${repos[@]}"; do
    repo_name=$(basename "$repo" .git)
    git clone "$repo" "./plugins/$repo_name"
    cd "./plugins/$repo_name" && go build -o plugin_$repo_name && cd ../../
done