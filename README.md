# Bedrock + MCP Integration

Ce projet intègre AWS Bedrock Claude avec un serveur MCP local pour accéder à une base de données PostgreSQL.

## 🏗️ Architecture

```
[Votre App Go] → [AWS Bedrock Claude] → [Outils MCP] → [Serveur MCP Local] → [PostgreSQL]
```

## 🚀 Démarrage rapide

### 1. Construire le projet
```bash
make build
```

### 2. Démarrer le serveur MCP (simulation)
```bash
make dev-server
# ou en arrière-plan:
make run-server
```

### 3. Tester l'intégration
```bash
make dev-client
# ou avec des paramètres personnalisés:
go run . -d -mcp http://localhost:8080 -q "Analyse les coûts du module eks-zidane"
```

## 📋 Options de ligne de commande

- `-d` : Mode debug
- `-mcp URL` : URL du serveur MCP (défaut: http://localhost:8080)  
- `-q "query"` : Question à poser à Claude

## 🔧 Configuration

### Variables d'environnement AWS
```bash
export AWS_REGION=eu-west-1
export AWS_PROFILE=your-profile
```

### Serveur MCP réel
Pour connecter à votre vrai serveur MCP PostgreSQL, modifiez l'URL :
```bash
go run . -mcp http://your-mcp-server:port -q "Votre question"
```

## 🛠️ Outils disponibles

Le système expose ces outils à Claude :

1. **postgres_read_query** - Exécuter des requêtes SQL lecture seule
2. **postgres_list_tables** - Lister toutes les tables  
3. **postgres_desc_table** - Décrire la structure d'une table

## 📝 Exemples de requêtes

```bash
# Analyse des coûts
go run . -q "Montre-moi les coûts du module eks-zidane pour les 6 derniers mois"

# Structure de base
go run . -q "Liste toutes les tables et décris leur structure"

# Analyse personnalisée  
go run . -q "Trouve les modules les plus coûteux et prédis leurs coûts futurs"
```

## 🔍 Debugging

Activer les logs debug :
```bash
go run . -d -q "Votre question"
```

## 🏃‍♂️ Tests

```bash
# Test complet avec MCP
make test

# Test Bedrock seulement
make test-bedrock-only
```
