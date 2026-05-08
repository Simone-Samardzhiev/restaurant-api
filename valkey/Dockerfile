FROM valkey/valkey:latest

CMD ["sh", "-c", "exec valkey-server --port ${PORT:-6379} --requirepass ${VALKEY_PASSWORD} --maxmemory ${MAX_MEMORY:-512mb} --maxmemory-policy allkeys-lru --protected-mode no"]