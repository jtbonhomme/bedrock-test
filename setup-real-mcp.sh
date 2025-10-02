#!/bin/bash

# Configuration pour le serveur MCP PostgreSQL réel (go-mcp-postgres)
# Source: https://github.com/guoling2008/go-mcp-postgres

echo "🔧 Configuration du serveur MCP PostgreSQL réel"
echo "=============================================="

# Variables d'environnement pour PostgreSQL
cat << 'EOF'
# ===== Configuration PostgreSQL =====
# Remplacez ces valeurs par vos vraies informations de connexion

export POSTGRES_HOST="your-postgres-host.com"
export POSTGRES_PORT="5432"
export POSTGRES_DB="your_database_name"
export POSTGRES_USER="your_username" 
export POSTGRES_PASSWORD="your_password"
export POSTGRES_SSLMODE="require"  # ou "disable" pour les tests locaux

# Port du serveur MCP (défaut souvent 8080)
export MCP_PORT="8080"

# URL complète (alternative à la configuration séparée)
export DATABASE_URL="postgresql://user:password@host:port/database?sslmode=require"

# ===== Configuration pour votre cas d'usage =====
# Si vous utilisez les tables de facturation mentionnées précédemment:

# Exemple pour une base locale de test:
# export POSTGRES_HOST="localhost"
# export POSTGRES_PORT="5432"
# export POSTGRES_DB="billing_db"
# export POSTGRES_USER="postgres"
# export POSTGRES_PASSWORD="postgres"
# export POSTGRES_SSLMODE="disable"

# Exemple pour une base distante:
# export POSTGRES_HOST="your-cloud-postgres.amazonaws.com"
# export POSTGRES_DB="production_billing"
# export POSTGRES_USER="readonly_user"
# export POSTGRES_PASSWORD="your_secure_password"
# export POSTGRES_SSLMODE="require"
EOF

echo
echo "📋 Instructions:"
echo "1. Copiez les variables ci-dessus dans un fichier .env.mcp"
echo "2. Adaptez les valeurs à votre configuration PostgreSQL"  
echo "3. Chargez les variables: source .env.mcp"
echo "4. Démarrez go-mcp-postgres"

echo
echo "🚀 Exemples de démarrage de go-mcp-postgres:"
echo
echo "# Option 1: Via variables d'environnement"
echo "source .env.mcp"
echo "go-mcp-postgres"
echo
echo "# Option 2: Via DATABASE_URL"
echo "DATABASE_URL='postgresql://user:pass@host:5432/db' go-mcp-postgres"
echo
echo "# Option 3: Avec port personnalisé"
echo "MCP_PORT=8081 go-mcp-postgres"

echo
echo "📡 Test de connectivité:"
echo "curl http://localhost:8080/health"
