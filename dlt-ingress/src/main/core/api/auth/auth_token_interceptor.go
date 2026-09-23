package auth

import (
	"context"
	"dlt-ingress/src/main/config"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

var keySetMap map[string]jwk.Set

func TokenInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.NewErrorResponse("Token not received", api.ErrorCodeInvalidToken, http.StatusUnauthorized))
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		unverified, err := validateToken(tokenStr)

		if err != nil {
			fields := []any{
				"error", err.Error(),
			}
			if unverified != nil {
				fields = append(fields, "iss", unverified.Issuer())
			}
			logger.WarnWithCtx(c.Request.Context(), "token validation failed", fields...)
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.NewErrorResponse("Invalid Token", api.ErrorCodeInvalidToken, http.StatusUnauthorized))
			return
		}

		c.Next()
	}
}

func validateToken(tokenStr string) (jwt.Token, error) {
	unverified, err := jwt.ParseInsecure([]byte(tokenStr))
	if err != nil {
		return nil, err
	}
	keySet, found := keySetMap[unverified.Issuer()]
	if !found {
		return unverified, fmt.Errorf("issuer not found in sources")
	}
	_, err = jwt.ParseString(
		tokenStr,
		jwt.WithRequiredClaim("iss"),
		jwt.WithKeySet(
			keySet,
			jws.WithRequireKid(false),
			jws.WithInferAlgorithmFromKey(true),
		),
	)
	if err != nil {
		return unverified, err
	}
	return unverified, nil
}

func SetupAuth(ctx context.Context) error {
	keySetMap = make(map[string]jwk.Set, len(config.AppConfig.Auth.Jwt.Sources))
	for _, jwksSource := range config.AppConfig.Auth.Jwt.Sources {
		if jwksSource.Issuer == "" {
			logger.WarnWithCtx(ctx, "Skipping empty issuer")
			continue
		}
		var sourceKeySet jwk.Set
		if jwksSource.Uri != "" {
			cache := jwk.NewCache(ctx) // Cache works using HTTP response headers "Cache-Control: max-age" and "Expires" from IdPs to know when to refresh cache
			jwksUriErrorMsg := "failed to read JWT JWKS from uri: %w"
			if err := cache.Register(jwksSource.Uri); err != nil {
				return fmt.Errorf(jwksUriErrorMsg, err)
			}
			if _, err := cache.Refresh(ctx, jwksSource.Uri); err != nil {
				return fmt.Errorf(jwksUriErrorMsg, err)
			}
			sourceKeySet = jwk.NewCachedSet(cache, jwksSource.Uri)
		} else if jwksSource.PublicKey != "" {
			var err error
			publicKey := strings.ReplaceAll(jwksSource.PublicKey, `\n`, "\n")
			if sourceKeySet, err = jwk.ParseString(publicKey, jwk.WithPEM(true)); err != nil {
				return fmt.Errorf("failed to parse public key: %w", err)
			}
		} else {
			return fmt.Errorf("invalid config: uri or publicKey should be filled for JWT sources")
		}
		keySetMap[jwksSource.Issuer] = sourceKeySet
	}
	if len(keySetMap) == 0 {
		return fmt.Errorf("invalid config: at least a JWT source should be configured")
	}
	return nil
}
