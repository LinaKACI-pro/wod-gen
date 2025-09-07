package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LinaKACI-pro/wod-gen/internal/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

//  fix Les defaults des tags (default:"20") ne s’appliquent que
// si ton parseur les gère (caarlos0/env oui, mais uniquement si le tag est env:"..." default:"..." sur la bonne struct).
//
// Si ta RateLimiterConfig n’a pas ces tags, ou si la variable n’est pas lue (mauvais nom d’env), tu te retrouves avec 0.

const (
	Global = "global"
	Token  = "token"
	IP     = "ip"
)

type bucket struct {
	lim     *rate.Limiter
	expires time.Time
}

type RateLimiter struct {
	cfg     config.RateLimiterConfig
	global  *rate.Limiter
	mu      sync.Mutex
	buckets map[string]*bucket
	ticker  *time.Ticker
	stopCh  chan struct{}
}

func NewLimiter(cfg *config.RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		cfg:     *cfg,
		buckets: make(map[string]*bucket),
		stopCh:  make(chan struct{}),
	}

	if cfg.Strategy == Global {
		rl.global = rate.NewLimiter(rate.Limit(cfg.RequestsPerS), cfg.Burst)
	}
	// GC: toutes les max(1m, TTL)
	intv := cfg.TTL
	if intv < time.Minute {
		intv = time.Minute
	}
	rl.ticker = time.NewTicker(intv)
	go rl.gc()
	return rl
}

func (r *RateLimiter) Stop() {
	close(r.stopCh)
	if r.ticker != nil {
		r.ticker.Stop()
	}
}

func (r *RateLimiter) gc() {
	for {
		select {
		case <-r.stopCh:
			return
		case now := <-r.ticker.C:
			r.mu.Lock()
			for k, b := range r.buckets {
				if now.After(b.expires) {
					delete(r.buckets, k)
				}
			}
			r.mu.Unlock()
		}
	}
}

func (r *RateLimiter) keyFromRequest(c *gin.Context) string {
	switch r.cfg.Strategy {
	case Token:
		// Authorization: Bearer <token>
		auth := c.Request.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tok := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			if tok != "" {
				sum := sha256.Sum256([]byte(tok))
				return "tok_" + hex.EncodeToString(sum[:8]) // id court, jamais le token
			}
		}
		// fallback IP si pas de token
		fallthrough
	case IP:
		// Gin sait extraire le bon client IP si TrustProxy est configuré
		ip := c.ClientIP()
		// normalise: on ne garde que l’IP (sans port)
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		if ip == "" {
			ip = "unknown"
		}
		return "ip_" + ip
	default: // "global"
		return Global
	}
}

func (r *RateLimiter) allowForKey(key string) (bool, time.Duration) {
	now := time.Now()

	if r.cfg.Strategy == "global" {
		res := r.global.ReserveN(now, 1)
		if !res.OK() {
			// trop de backlog → refuse avec retry ~ 1/RPS
			return false, time.Duration(float64(time.Second) / r.cfg.RequestsPerS)
		}
		if d := res.DelayFrom(now); d > 0 {
			res.CancelAt(now)
			return false, d
		}

		return true, 0
	}

	// Par clé (ip/token)
	r.mu.Lock()
	b, ok := r.buckets[key]
	if !ok || now.After(b.expires) {
		b = &bucket{
			lim:     rate.NewLimiter(rate.Limit(r.cfg.RequestsPerS), r.cfg.Burst),
			expires: now.Add(r.cfg.TTL),
		}
		r.buckets[key] = b
	}
	b.expires = now.Add(r.cfg.TTL)
	res := b.lim.ReserveN(now, 1)
	r.mu.Unlock()

	if !res.OK() {
		return false, time.Duration(float64(time.Second) / r.cfg.RequestsPerS)
	}
	if d := res.DelayFrom(now); d > 0 {
		res.CancelAt(now)
		return false, d
	}
	return true, 0
}

func (r *RateLimiter) Middleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.cfg.Enabled {
			c.Next()
			return
		}
		key := r.keyFromRequest(c)
		ok, delay := r.allowForKey(key)
		if ok {
			c.Next()
			return
		}
		if delay > 0 {
			// seconds arrondis
			secs := int(delay.Round(time.Second) / time.Second)
			if secs < 1 {
				secs = 1
			}
			c.Header("Retry-After", strconv.Itoa(secs))
		}
		if logger != nil {
			logger.Warn("rate_limit_exceeded",
				slog.String("key", key),
				slog.String("path", c.FullPath()),
				slog.String("method", c.Request.Method),
				slog.Int("status", http.StatusTooManyRequests),
			)
		}
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"code":    http.StatusTooManyRequests,
			"message": "rate limit exceeded",
		})
	}
}
