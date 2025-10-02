# Guide d'utilisation : MCP stdio avec go-mcp-postgres

## 🎯 Problème résolu

Le serveur `go-mcp-postgres` n'utilise pas HTTP mais le protocole MCP standard via stdio (stdin/stdout). Votre programme doit maintenant communiquer directement avec l'exécutable via des appels système.

## 🔧 Solution implémentée

### Architecture mise à jour :
```
[Votre App Go] → [Appel système] → [go-mcp-postgres --dsn ...] → [PostgreSQL]
                     ↕️ stdio
```

### Fichiers créés :

1. **`internal/mcp/stdio_client.go`** - Client MCP via stdio
2. **`internal/mcp/http_client.go`** - Client MCP via HTTP (pour compatibilité)
3. **`internal/mcp/auto_client.go`** - Auto-détection du type de client
4. **`cmd/test-mcp-stdio/main.go`** - Utilitaire de test stdio
5. **`test-mcp-clients.sh`** - Script de test complet

## 🚀 Utilisation

### Option 1: Client stdio (Production - Recommandé)

```bash
# Installer go-mcp-postgres
go install github.com/guoling2008/go-mcp-postgres@latest

# Utiliser avec DSN
export DATABASE_URL="postgresql://user:password@host:port/database?sslmode=require"

# Lancer votre application Bedrock
go run . -mcp-dsn "$DATABASE_URL" -q "Analyse les coûts du module eks-zidane"
```

### Option 2: Client HTTP (Tests/Développement)

```bash
# Démarrer le serveur HTTP de test
go run ./cmd/mcp-server &

# Utiliser avec URL HTTP
go run . -mcp http://localhost:8080 -q "Votre question"
```

## 📋 Exemples d'utilisation

### Analyse de coûts avec vraies données PostgreSQL :

```bash
# Configuration
export DATABASE_URL="postgresql://readonly_user:password@your-db.amazonaws.com:5432/billing_db?sslmode=require"

# Requêtes d'analyse
go run . -mcp-dsn "$DATABASE_URL" -q "Liste toutes les tables de facturation"

go run . -mcp-dsn "$DATABASE_URL" -q "Montre les coûts mensuels du module eks-zidane"

go run . -mcp-dsn "$DATABASE_URL" -q "Prédis les coûts d'octobre, novembre et décembre pour eks-zidane"

go run . -mcp-dsn "$DATABASE_URL" -q "Compare les modules les plus coûteux de l'équipe core-cloud"
```

### Test de connectivité :

```bash
# Test du client stdio uniquement
go run ./cmd/test-mcp-stdio -dsn "$DATABASE_URL" -debug

# Test complet (HTTP + stdio)
./test-mcp-clients.sh
```

## 🔍 Débogage

### Vérifier la connexion PostgreSQL :
```bash
# Test direct
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM billing_report_unified_provider;"
```

### Vérifier go-mcp-postgres :
```bash
# Test manuel
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}' | go-mcp-postgres --dsn "$DATABASE_URL"
```

### Logs de débogage :
```bash
# Mode debug
go run . -d -mcp-dsn "$DATABASE_URL" -q "Votre question"
```

## ⚡ Avantages du client stdio

- ✅ **Accès direct** aux vraies données PostgreSQL
- ✅ **Pas de serveur HTTP** intermédiaire requis
- ✅ **Sécurité renforcée** (pas de port réseau ouvert)
- ✅ **Compatible** avec le protocole MCP standard
- ✅ **Performance** optimale (pas de sérialisation HTTP)

## 🆘 Résolution de problèmes

### Erreur "executable file not found"
```bash
# Installer go-mcp-postgres
go install github.com/guoling2008/go-mcp-postgres@latest

# Ou ajouter au PATH
export PATH=$PATH:~/go/bin
```

### Erreur de connexion PostgreSQL
```bash
# Vérifier les credentials
psql "$DATABASE_URL" -c "SELECT 1;"

# Vérifier les permissions
psql "$DATABASE_URL" -c "SELECT * FROM information_schema.tables LIMIT 1;"
```

### Timeout ou blocage
```bash
# Réduire la complexité de la requête
go run . -mcp-dsn "$DATABASE_URL" -q "SELECT COUNT(*) FROM billing_report_unified_provider LIMIT 10"
```

## 🔄 Migration depuis HTTP

Si vous utilisiez déjà le client HTTP, il suffit de changer le flag :

```bash
# Ancien (HTTP)
go run . -mcp http://localhost:8080 -q "Votre question"

# Nouveau (stdio)  
go run . -mcp-dsn "$DATABASE_URL" -q "Votre question"
```

Le code détecte automatiquement le type de client à utiliser !

## 📞 Support

- **MCP Protocol** : https://spec.modelcontextprotocol.io/
- **go-mcp-postgres** : https://github.com/guoling2008/go-mcp-postgres
- **AWS Bedrock** : https://docs.aws.amazon.com/bedrock/
