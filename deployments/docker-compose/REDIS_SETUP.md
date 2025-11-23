# Docker Compose - Redis Integration

## ✅ Changes Made

Successfully added Redis to the docker-compose.yml configuration for the portfolio-insights application.

---

## 📋 What Was Added

### **1. Redis Service**

```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
  volumes:
    - redis_data:/data
  command: redis-server --appendonly yes
```

**Features:**
- ✅ Redis 7 (latest stable)
- ✅ Alpine Linux (lightweight)
- ✅ Persistent storage with AOF (Append-Only File)
- ✅ Exposed on port 6379
- ✅ Named volume for data persistence

### **2. Updated Portfolio Service**

**Added Environment Variables:**
```yaml
portfolio-service:
  environment:
    - DB_HOST=postgres
    - DB_USER=garcios
    - DB_PASSWORD=Password123
    - DB_NAME=portfolio
    - NATS_URL=nats://nats:4222
    - REDIS_ADDR=redis:6379                      # NEW
    - PRICE_CACHE_TTL_SECONDS=60                 # NEW
    - MARKETDATA_SERVICE_ADDR=marketdata-service:50054  # NEW
```

**Updated Dependencies:**
```yaml
depends_on:
  - postgres      # NEW - for holdings storage
  - nats          # existing
  - redis         # NEW - for price caching
  - marketdata-service  # NEW - for price fetching
```

### **3. Redis Data Volume**

```yaml
volumes:
  postgres_data:
  minio_data:
  redis_data:     # NEW
```

---

## 🔧 Configuration Details

### Redis Configuration

| Setting | Value | Description |
|---------|-------|-------------|
| **Image** | `redis:7-alpine` | Latest Redis 7 on Alpine Linux |
| **Port** | `6379` | Standard Redis port |
| **Persistence** | AOF enabled | Append-Only File for durability |
| **Volume** | `redis_data` | Named volume for data persistence |

### Portfolio Service Configuration

| Variable | Value | Purpose |
|----------|-------|---------|
| `DB_HOST` | `postgres` | PostgreSQL host for holdings |
| `DB_NAME` | `portfolio` | Database name |
| `REDIS_ADDR` | `redis:6379` | Redis server address |
| `PRICE_CACHE_TTL_SECONDS` | `60` | Cache TTL (1 minute) |
| `MARKETDATA_SERVICE_ADDR` | `marketdata-service:50054` | Market data gRPC endpoint |

---

## 🚀 Usage

### Start All Services

```bash
# Using podman-compose
make podman-up

# Or directly
podman-compose -f deployments/docker-compose/docker-compose.yml up -d

# Using docker-compose
docker-compose -f deployments/docker-compose/docker-compose.yml up -d
```

### Check Redis Status

```bash
# Check if Redis is running
podman ps | grep redis

# Connect to Redis CLI
podman exec -it portfolio-insights-redis-1 redis-cli

# Test Redis
redis> PING
PONG

# Check cached prices
redis> KEYS price:*
redis> GET price:AAPL
```

### View Logs

```bash
# Redis logs
podman logs -f portfolio-insights-redis-1

# Portfolio service logs
podman logs -f portfolio-insights-portfolio-service-1
```

### Stop Services

```bash
make podman-down

# Or
podman-compose -f deployments/docker-compose/docker-compose.yml down
```

---

## 📊 Service Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Docker Compose                       │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│  │PostgreSQL│  │   NATS   │  │  Redis   │  │ MinIO  │ │
│  │  :5432   │  │  :4222   │  │  :6379   │  │ :9000  │ │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬───┘ │
│       │             │              │              │     │
│  ┌────┴──────────────┴──────────────┴──────────────┴──┐ │
│  │                                                     │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────┐ │ │
│  │  │ User Service │  │Portfolio Svc │  │Market Svc│ │ │
│  │  │   :50051     │  │   :50052     │  │  :50054  │ │ │
│  │  └──────────────┘  └──────┬───────┘  └────┬─────┘ │ │
│  │                           │                │       │ │
│  │                    ┌──────┴────────────────┘       │ │
│  │                    │                               │ │
│  │              ┌─────▼─────┐                         │ │
│  │              │  Gateway  │                         │ │
│  │              │   :8080   │                         │ │
│  │              └───────────┘                         │ │
│  └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

---

## 🔄 Data Flow with Redis

### Portfolio Holdings Request

```
1. Client → Gateway → Portfolio Service
                          ↓
2. Portfolio Service → PostgreSQL (get holdings)
                          ↓
3. Portfolio Service → Redis (check cached prices)
                          ↓
4. Cache Hit? → Return cached prices
                          ↓
5. Cache Miss → Market Service (fetch prices)
                          ↓
6. Portfolio Service → Redis (cache prices)
                          ↓
7. Portfolio Service → Client (return enriched holdings)
```

---

## 🧪 Testing

### 1. Verify Redis is Running

```bash
# Check container status
podman ps | grep redis

# Expected output:
# portfolio-insights-redis-1   redis:7-alpine   redis-server ...   Up   0.0.0.0:6379->6379/tcp
```

### 2. Test Redis Connection

```bash
# Connect to Redis CLI
podman exec -it portfolio-insights-redis-1 redis-cli

# Test commands
redis> PING
PONG

redis> SET test "Hello Redis"
OK

redis> GET test
"Hello Redis"

redis> DEL test
(integer) 1
```

### 3. Monitor Cache Activity

```bash
# Monitor all Redis commands
podman exec -it portfolio-insights-redis-1 redis-cli MONITOR

# In another terminal, make a portfolio request
grpcurl -plaintext \
  -d '{"user_id": "user-123"}' \
  localhost:50052 \
  portfolio.PortfolioService/GetHoldings

# You should see cache operations in MONITOR output
```

### 4. Check Cached Prices

```bash
podman exec -it portfolio-insights-redis-1 redis-cli

redis> KEYS price:*
1) "price:AAPL"
2) "price:GOOGL"
3) "price:MSFT"

redis> GET price:AAPL
"{\"symbol\":\"AAPL\",\"price\":175.50,\"timestamp\":\"2025-11-24T09:42:00Z\",\"cached_at\":\"2025-11-24T09:42:05Z\"}"

redis> TTL price:AAPL
(integer) 45  # Seconds remaining
```

---

## 📈 Monitoring

### Redis Stats

```bash
# Get Redis info
podman exec -it portfolio-insights-redis-1 redis-cli INFO

# Key sections:
# - Server: Redis version, uptime
# - Clients: Connected clients
# - Memory: Memory usage
# - Stats: Commands processed, hits/misses
# - Persistence: AOF status
```

### Cache Metrics

```bash
# Get cache hit/miss stats
podman exec -it portfolio-insights-redis-1 redis-cli INFO stats

# Look for:
# keyspace_hits: Number of successful lookups
# keyspace_misses: Number of failed lookups
# Hit rate = hits / (hits + misses)
```

---

## 🔧 Configuration Options

### Adjust Cache TTL

Edit `docker-compose.yml`:

```yaml
portfolio-service:
  environment:
    - PRICE_CACHE_TTL_SECONDS=30   # 30 seconds
    # or
    - PRICE_CACHE_TTL_SECONDS=300  # 5 minutes
```

Then restart:
```bash
podman-compose restart portfolio-service
```

### Redis Memory Limit

Add to Redis service:

```yaml
redis:
  image: redis:7-alpine
  command: redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
```

### Redis Password (Production)

```yaml
redis:
  image: redis:7-alpine
  command: redis-server --appendonly yes --requirepass your-secure-password
  environment:
    - REDIS_PASSWORD=your-secure-password

portfolio-service:
  environment:
    - REDIS_ADDR=redis:6379
    - REDIS_PASSWORD=your-secure-password
```

---

## 🐛 Troubleshooting

### Redis Not Starting

**Check logs:**
```bash
podman logs portfolio-insights-redis-1
```

**Common issues:**
- Port 6379 already in use
- Volume permission issues
- Insufficient memory

### Portfolio Service Can't Connect to Redis

**Check:**
```bash
# 1. Redis is running
podman ps | grep redis

# 2. Network connectivity
podman exec portfolio-insights-portfolio-service-1 ping redis

# 3. Environment variables
podman exec portfolio-insights-portfolio-service-1 env | grep REDIS
```

### Cache Not Working

**Verify:**
```bash
# 1. Check portfolio service logs
podman logs portfolio-insights-portfolio-service-1 | grep -i redis

# Expected:
# Successfully connected to Redis
# Price caching enabled

# 2. Monitor Redis
podman exec -it portfolio-insights-redis-1 redis-cli MONITOR

# 3. Make a request and watch for cache operations
```

---

## 💾 Data Persistence

### Redis AOF (Append-Only File)

Redis is configured with AOF for data persistence:

```yaml
command: redis-server --appendonly yes
```

**Benefits:**
- Every write operation is logged
- Data survives container restarts
- Automatic recovery on startup

**Location:**
- Data stored in `redis_data` volume
- Persists across container recreations

### Backup Redis Data

```bash
# Create backup
podman exec portfolio-insights-redis-1 redis-cli BGSAVE

# Copy backup file
podman cp portfolio-insights-redis-1:/data/dump.rdb ./redis-backup.rdb
```

### Restore Redis Data

```bash
# Stop Redis
podman-compose stop redis

# Copy backup to volume
podman cp ./redis-backup.rdb portfolio-insights-redis-1:/data/dump.rdb

# Start Redis
podman-compose start redis
```

---

## 🔐 Security Best Practices

### Production Recommendations

1. **Enable Authentication:**
   ```yaml
   redis:
     command: redis-server --requirepass ${REDIS_PASSWORD}
   ```

2. **Disable Dangerous Commands:**
   ```yaml
   redis:
     command: redis-server --rename-command FLUSHDB "" --rename-command FLUSHALL ""
   ```

3. **Bind to Specific Interface:**
   ```yaml
   redis:
     command: redis-server --bind 0.0.0.0 --protected-mode yes
   ```

4. **Use TLS (Production):**
   ```yaml
   redis:
     command: redis-server --tls-port 6379 --port 0 --tls-cert-file /certs/redis.crt --tls-key-file /certs/redis.key
   ```

---

## 📚 Files Modified

1. ✅ `deployments/docker-compose/docker-compose.yml`
   - Added Redis service
   - Updated portfolio-service configuration
   - Added redis_data volume

---

## 🎯 Next Steps

### Immediate
1. ✅ Redis added to docker-compose
2. ⏳ Restart services with `make podman-down && make podman-up`
3. ⏳ Verify Redis connectivity
4. ⏳ Test price caching

### Short-term
5. ⏳ Monitor cache hit rates
6. ⏳ Adjust TTL based on usage patterns
7. ⏳ Add Redis monitoring to Grafana
8. ⏳ Set up Redis backup automation

### Long-term
9. ⏳ Consider Redis Cluster for HA
10. ⏳ Implement Redis Sentinel for failover
11. ⏳ Add Redis metrics to Prometheus
12. ⏳ Implement cache warming strategies

---

**Status**: ✅ **Redis successfully added to docker-compose!**

The portfolio-service now has Redis available for price caching, improving performance and reducing load on the market-service.

