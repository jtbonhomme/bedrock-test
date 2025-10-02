#!/bin/bash

# Script pour utiliser go-mcp-postgres au lieu du serveur de test
# Usage: ./use-real-mcp.sh [PORT]

set -euo pipefail

MCP_PORT=${1:-8080}
MCP_URL="http://localhost:${MCP_PORT}"

echo "🔄 Configuration pour utiliser go-mcp-postgres réel"
echo "=================================================="

# Fonction pour vérifier si le serveur MCP est en cours d'exécution
check_mcp_server() {
    local port=$1
    if curl -s "http://localhost:${port}/health" >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Fonction pour tester la connectivité MCP
test_mcp_connectivity() {
    local url=$1
    
    echo "🧪 Test de connectivité MCP vers $url"
    
    # Test 1: Health check
    if curl -s "${url}/health" >/dev/null; then
        echo "✅ Health check OK"
    else
        echo "❌ Health check échoué"
        return 1
    fi
    
    # Test 2: Liste des outils
    echo "📋 Outils disponibles:"
    if tools=$(curl -s "${url}/tools" 2>/dev/null); then
        echo "$tools" | jq -r '.[].name' | sed 's/^/   - /' || echo "   Erreur lors du parsing JSON"
    else
        echo "   ❌ Impossible de récupérer les outils"
        return 1
    fi
    
    # Test 3: Test d'un outil simple
    echo "🔧 Test de l'outil list_databases:"
    if result=$(curl -s -X POST "${url}/tools/execute" \
        -H "Content-Type: application/json" \
        -d '{"name":"mcp_postgres_list_database","arguments":{}}' 2>/dev/null); then
        
        if echo "$result" | jq -e '.content[0].text' >/dev/null 2>&1; then
            echo "✅ Outil fonctionnel"
            echo "   Réponse: $(echo "$result" | jq -r '.content[0].text' | head -1)"
        else
            echo "❌ Réponse invalide de l'outil"
            echo "   Debug: $result"
        fi
    else
        echo "❌ Impossible d'exécuter l'outil"
        return 1
    fi
    
    return 0
}

# Vérifier si go-mcp-postgres est installé
if ! command -v go-mcp-postgres >/dev/null 2>&1; then
    echo "❌ go-mcp-postgres n'est pas installé"
    echo
    echo "📦 Installation:"
    echo "go install github.com/guoling2008/go-mcp-postgres@latest"
    echo
    echo "📋 Ou depuis les sources:"
    echo "git clone https://github.com/guoling2008/go-mcp-postgres.git"
    echo "cd go-mcp-postgres"  
    echo "go build -o go-mcp-postgres ."
    exit 1
fi

# Vérifier si le serveur MCP est déjà en cours d'exécution
if check_mcp_server $MCP_PORT; then
    echo "✅ Serveur MCP détecté sur le port $MCP_PORT"
    
    # Tester la connectivité
    if test_mcp_connectivity "$MCP_URL"; then
        echo
        echo "🎉 Serveur go-mcp-postgres opérationnel !"
        echo
        echo "📝 Pour utiliser avec votre client Bedrock:"
        echo "   go run . -d -mcp $MCP_URL -q \"Liste les tables de facturation\""
        echo
        echo "📊 Exemple de requête SQL via MCP:"
        echo "   go run . -mcp $MCP_URL -q \"Montre-moi les coûts du module eks-zidane\""
        
    else
        echo "❌ Le serveur sur le port $MCP_PORT ne semble pas être go-mcp-postgres"
        echo "   Vérifiez que c'est bien le bon serveur MCP"
    fi
else
    echo "❌ Aucun serveur MCP détecté sur le port $MCP_PORT"
    echo
    echo "🚀 Pour démarrer go-mcp-postgres:"
    echo
    echo "1. Configurez PostgreSQL:"
    echo "   cp .env.mcp.example .env.mcp"
    echo "   # Éditez .env.mcp avec vos vraies informations de connexion"
    echo "   source .env.mcp"
    echo
    echo "2. Démarrez le serveur:"
    echo "   MCP_PORT=$MCP_PORT go-mcp-postgres &"
    echo
    echo "3. Relancez ce script pour vérifier:"
    echo "   ./use-real-mcp.sh $MCP_PORT"
fi
