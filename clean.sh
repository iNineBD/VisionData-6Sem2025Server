#!/bin/bash

# Script para limpar Docker completamente

echo "Parando todos os containers..."
docker stop $(docker ps -aq) 2>/dev/null || echo "Nenhum container em execução."

echo "Removendo todos os containers..."
docker rm -f $(docker ps -aq) 2>/dev/null || echo "Nenhum container para remover."

echo "Removendo todas as imagens..."
docker rmi -f $(docker images -q) 2>/dev/null || echo "Nenhuma imagem para remover."

echo "Removendo todos os volumes..."
docker volume rm $(docker volume ls -q) 2>/dev/null || echo "Nenhum volume para remover."

echo "Executando prune do sistema..."
docker system prune -a --volumes -f

echo "Limpeza completa do Docker!"
