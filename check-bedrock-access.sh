#!/bin/bash

echo "=== Vérification de l'accès AWS Bedrock ==="

# Vérifier la configuration AWS
echo "1. Configuration AWS actuelle:"
aws configure list

echo -e "\n2. Région AWS:"
aws configure get region

echo -e "\n3. Test de connectivité AWS:"
aws sts get-caller-identity

echo -e "\n4. Vérification des modèles Bedrock disponibles dans votre région:"
echo "   (Ceci peut échouer si vous n'avez pas accès à Bedrock)"
aws bedrock list-foundation-models --region us-east-1 2>/dev/null | jq '.modelSummaries[] | select(.modelId | contains("claude")) | {modelId, modelName}' || echo "   ❌ Pas d'accès à Bedrock ou région incorrecte"

echo -e "\n5. Modèles suggérés à tester (par ordre de disponibilité):"
echo "   - anthropic.claude-instant-v1"
echo "   - anthropic.claude-v2"
echo "   - anthropic.claude-3-haiku-20240307-v1:0"
echo "   - anthropic.claude-3-sonnet-20240229-v1:0"

echo -e "\n6. Régions recommandées pour Bedrock:"
echo "   - us-east-1 (Virginie du Nord) - Meilleure disponibilité"
echo "   - us-west-2 (Oregon)"
echo "   - eu-west-1 (Irlande) - Pour l'Europe"

echo -e "\n=== Actions recommandées ==="
echo "1. Vérifiez votre région AWS (doit être compatible Bedrock)"
echo "2. Activez l'accès aux modèles Claude dans la console AWS Bedrock"
echo "3. Configurez votre région avec: aws configure set region us-east-1"
echo "4. Testez avec un modèle simple comme claude-instant-v1 d'abord"

echo -e "\n=== Variables d'environnement utiles ==="
echo "export AWS_REGION=us-east-1"
echo "export AWS_PROFILE=your-profile  # si vous utilisez des profils"
