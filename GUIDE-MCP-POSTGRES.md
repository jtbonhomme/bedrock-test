# Guide d'utilisation de go-mcp-postgres avec Bedrock

Ce guide explique comment remplacer le serveur MCP de test par le vrai serveur `go-mcp-postgres` pour se connecter à votre base PostgreSQL.

## 🎯 Objectif

Connecter votre application Bedrock Claude à une vraie base de données PostgreSQL via le serveur MCP `go-mcp-postgres` au lieu d'utiliser des données simulées.

## 📋 Prérequis

1. **go-mcp-postgres installé**
2. **Accès à une base PostgreSQL** avec les tables de facturation
3. **AWS Bedrock configuré** (pour Claude)

## 🔧 Étapes d'installation

### Étape 1: Installer go-mcp-postgres

```bash
# Option 1: Via go install
go install github.com/guoling2008/go-mcp-postgres@latest

# Option 2: Depuis les sources
git clone https://github.com/guoling2008/go-mcp-postgres.git
cd go-mcp-postgres
go build -o go-mcp-postgres .
sudo mv go-mcp-postgres /usr/local/bin/  # ou dans votre PATH
```

### Étape 2: Configuration PostgreSQL

```bash
# Copiez le fichier d'exemple
cp .env.mcp.example .env.mcp

# Éditez avec vos vraies informations
nano .env.mcp
```

**Exemple de configuration `.env.mcp`:**
```bash
# Base de données de production
POSTGRES_HOST=your-postgres.amazonaws.com
POSTGRES_PORT=5432
POSTGRES_DB=billing_database
POSTGRES_USER=readonly_user
POSTGRES_PASSWORD=your_secure_password
POSTGRES_SSLMODE=require
MCP_PORT=8080
```

### Étape 3: Démarrer go-mcp-postgres

```bash
# Charger la configuration
source .env.mcp

# Démarrer le serveur MCP
go-mcp-postgres &

# Vérifier qu'il fonctionne
curl http://localhost:8080/health
```

### Étape 4: Tester la connectivité

```bash
# Utiliser le script de vérification
./use-real-mcp.sh

# Ou tester manuellement
curl -s http://localhost:8080/tools | jq .
```

## 🚀 Utilisation avec Bedrock

### Lancement basique

```bash
# Dans votre projet bedrock-test
go run . -d -mcp http://localhost:8080 -q "Liste toutes les tables de la base"
```

### Requêtes d'analyse de coûts

```bash
# Analyse d'un module spécifique
go run . -mcp http://localhost:8080 -q "Montre-moi les coûts mensuels du module eks-zidane"

# Prédictions de coûts
go run . -mcp http://localhost:8080 -q "Prédis les coûts d'octobre, novembre et décembre pour eks-zidane"

# Analyse comparative
go run . -mcp http://localhost:8080 -q "Compare les coûts de tous les modules de l'équipe core-cloud"
```

### Requêtes SQL complexes

```bash
# Tendances annuelles
go run . -mcp http://localhost:8080 -q "Analyse la croissance des coûts year-over-year par équipe"

# Top modules les plus coûteux
go run . -mcp http://localhost:8080 -q "Trouve les 10 modules les plus coûteux et leur évolution"
```

## 🔍 Debugging

### Vérifier la connexion PostgreSQL

```bash
# Test direct de la base
psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT COUNT(*) FROM billing_report_unified_provider;"
```

### Vérifier les outils MCP

```bash
# Liste des outils disponibles
curl -s http://localhost:8080/tools | jq -r '.[].name'

# Test d'un outil spécifique
curl -X POST http://localhost:8080/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"name":"mcp_postgres_list_table","arguments":{}}'
```

### Logs du serveur MCP

```bash
# Démarrer en mode verbose pour voir les logs
RUST_LOG=debug go-mcp-postgres  # si c'est en Rust
# ou
go-mcp-postgres -verbose       # selon l'implémentation
```

## ⚠️ Sécurité

1. **Utilisateur en lecture seule** : Créez un utilisateur PostgreSQL dédié avec des droits en lecture seule
2. **Connexions SSL** : Utilisez toujours `sslmode=require` en production
3. **Firewall** : Le serveur MCP ne devrait être accessible que localement
4. **Credentials** : Ne commitez jamais le fichier `.env.mcp`

### Exemple d'utilisateur PostgreSQL en lecture seule

```sql
-- Créer un utilisateur dédié pour MCP
CREATE USER mcp_readonly WITH PASSWORD 'secure_password';

-- Accorder les droits de lecture sur les tables nécessaires
GRANT CONNECT ON DATABASE billing_database TO mcp_readonly;
GRANT USAGE ON SCHEMA public TO mcp_readonly;
GRANT SELECT ON billing_report_unified_provider TO mcp_readonly;
GRANT SELECT ON module_mapping TO mcp_readonly;

-- Vérifier les droits
\dp billing_report_unified_provider
```

## 🆘 Résolution de problèmes

### Erreur de connexion PostgreSQL
```
Error: failed to connect to database
```
**Solution** : Vérifiez les credentials et la connectivité réseau

### Erreur "Tool not found"
```
Error: Tool 'mcp_postgres_read_query' not implemented
```
**Solution** : Vérifiez que go-mcp-postgres expose les bons outils

### Timeout Bedrock
```
Error: context deadline exceeded
```
**Solution** : Les requêtes SQL sont peut-être trop complexes, simplifiez ou ajoutez des index

## 📞 Support

- **Projet go-mcp-postgres** : https://github.com/guoling2008/go-mcp-postgres
- **Documentation MCP** : https://spec.modelcontextprotocol.io/
- **AWS Bedrock** : https://docs.aws.amazon.com/bedrock/
