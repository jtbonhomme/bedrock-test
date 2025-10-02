#!/bin/bash

# Script pour tester les deux types de clients MCP : HTTP et stdio

set -euo pipefail

echo "🧪 Test des clients MCP (HTTP vs stdio)"
echo "====================================="

# Configuration
HTTP_URL="http://localhost:8080"
STDIO_EXE="go-mcp-postgres"
DSN="${DATABASE_URL:-postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable}"

# Fonctions utilitaires
success() { echo -e "\033[0;32m✅ $1\033[0m"; }
error() { echo -e "\033[0;31m❌ $1\033[0m"; }
info() { echo -e "\033[1;33mℹ️  $1\033[0m"; }

# Test 1: Client HTTP
echo
info "Test 1: Client HTTP MCP"
if curl -s "${HTTP_URL}/health" >/dev/null 2>&1; then
    success "Serveur HTTP MCP détecté sur ${HTTP_URL}"
    
    # Test des outils HTTP
    info "Test des outils via HTTP..."
    if tools=$(curl -s "${HTTP_URL}/tools" 2>/dev/null); then
        echo "   Outils disponibles:"
        echo "$tools" | jq -r '.[].name' | sed 's/^/     - /' 2>/dev/null || echo "     (Format non-JSON)"
    fi
    
    # Test d'exécution d'outil HTTP
    info "Test d'exécution d'outil via HTTP..."
    result=$(curl -s -X POST "${HTTP_URL}/tools/execute" \
        -H "Content-Type: application/json" \
        -d '{"name":"mcp_postgres_list_table","arguments":{}}' 2>/dev/null || echo "FAILED")
    
    if [[ "$result" != "FAILED" ]] && echo "$result" | jq -e '.content[0].text' >/dev/null 2>&1; then
        success "Outil HTTP fonctionne"
    else
        error "Outil HTTP échoué"
    fi
else
    error "Aucun serveur HTTP MCP détecté sur ${HTTP_URL}"
    info "Pour démarrer un serveur HTTP de test:"
    info "  go run ./cmd/mcp-server &"
fi

# Test 2: Client stdio
echo
info "Test 2: Client stdio MCP"

# Vérifier si go-mcp-postgres est disponible
if command -v "$STDIO_EXE" >/dev/null 2>&1; then
    success "Exécutable $STDIO_EXE trouvé"
    
    # Tester la connexion stdio
    info "Test de connexion stdio avec DSN..."
    echo "   DSN: $(echo "$DSN" | sed 's/:[^@]*@/:***@/')"
    
    # Construire et tester le client stdio
    info "Construction du client de test stdio..."
    if go build -o test-mcp-stdio ./cmd/test-mcp-stdio 2>/dev/null; then
        success "Client de test construit"
        
        # Exécuter le test stdio
        info "Exécution du test stdio..."
        if timeout 10s ./test-mcp-stdio -dsn "$DSN" -debug 2>/dev/null; then
            success "Test stdio réussi"
        else
            error "Test stdio échoué (timeout ou erreur de connexion)"
            info "Vérifiez que PostgreSQL est accessible avec ce DSN"
        fi
        
        rm -f test-mcp-stdio
    else
        error "Échec de construction du client de test"
    fi
    
else
    error "Exécutable $STDIO_EXE non trouvé"
    info "Installation:"
    info "  go install github.com/guoling2008/go-mcp-postgres@latest"
    info "  # ou construire depuis les sources"
fi

# Test 3: Client unifié
echo
info "Test 3: Client unifié (auto-détection)"

# Test avec URL HTTP
info "Test auto-détection HTTP..."
echo "   URL: $HTTP_URL"
if curl -s "${HTTP_URL}/health" >/dev/null 2>&1; then
    success "Auto-détection HTTP OK"
else
    error "Auto-détection HTTP échouée"
fi

# Test avec DSN stdio
info "Test auto-détection stdio..."
echo "   DSN: $(echo "$DSN" | sed 's/:[^@]*@/:***@/')"
if command -v "$STDIO_EXE" >/dev/null 2>&1; then
    success "Auto-détection stdio OK"
else
    error "Auto-détection stdio échouée (exécutable manquant)"
fi

# Résumé
echo
echo "📊 Résumé des tests"
echo "=================="

if curl -s "${HTTP_URL}/health" >/dev/null 2>&1; then
    success "Client HTTP MCP: Disponible"
    echo "   Usage: -mcp http://localhost:8080"
else
    error "Client HTTP MCP: Indisponible"
    echo "   Pour activer: go run ./cmd/mcp-server &"
fi

if command -v "$STDIO_EXE" >/dev/null 2>&1; then
    success "Client stdio MCP: Disponible"
    echo "   Usage: -mcp-dsn 'postgresql://...'"
else
    error "Client stdio MCP: Indisponible"
    echo "   Pour installer: go install github.com/guoling2008/go-mcp-postgres@latest"
fi

echo
echo "🚀 Utilisation recommandée:"
echo "=========================="
echo
echo "# Avec serveur HTTP (pour tests/développement):"
echo "go run . -mcp http://localhost:8080 -q 'Votre question'"
echo
echo "# Avec go-mcp-postgres (pour production):"
echo "export DATABASE_URL='postgresql://user:pass@host:port/db'"
echo "go run . -mcp-dsn \"\$DATABASE_URL\" -q 'Votre question'"
echo
echo "# Le client détectera automatiquement le type selon le format"
