package jwt

import (
	"errors"
	"fmt"
	"go-todo/util/config"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type GtClaims struct {
	IsAdmin  bool   `json:"is_admin"`
	Username string `json:"username"`
	Family   string `json:"family"`
	jwt.RegisteredClaims
}

type JwtErrorReason int

const (
	JwtErrorReasonExpired JwtErrorReason = iota
	JwtErrorReasonInvalidSignature
	JwtErrorReasonTokenMalformed
	JwtErrorReasonUnhandled
)

func (t JwtErrorReason) String() string {
	switch t {
	case JwtErrorReasonExpired:
		return "token-expired"
	case JwtErrorReasonInvalidSignature:
		return "token-invalid-signature"
	case JwtErrorReasonTokenMalformed:
		return "token-malformed"
	case JwtErrorReasonUnhandled:
		return "token-unhandled-error"
	}
	return "unknown"
}

type JwtDecodeError struct {
	Claims *GtClaims
	Reason JwtErrorReason
	Err    error
}

func (e *JwtDecodeError) Error() string {
	return e.Err.Error()
}

func NewJwtDecodeError(claims *GtClaims, reason JwtErrorReason, err error) *JwtDecodeError {
	wrappedErr := fmt.Errorf("DecodeJwtError: %w", err)
	return &JwtDecodeError{
		Claims: claims,
		Reason: reason,
		Err:    wrappedErr,
	}
}

func generateJwt(username string, userID string, isAdmin bool, isRefreshToken bool, family string) (string, *GtClaims, error) {
	generateError := func(err error) error {
		return fmt.Errorf("GenerateJwtError: %w", err)
	}

	config, err := config.Get()
	if err != nil {
		return "", nil, generateError(err)
	}

	authLifeSpan := config.AccessTokenLifeSpan
	if isRefreshToken {
		authLifeSpan = config.RefreshTokenLifeSpan
	}
	lifeSpanDuration := time.Minute * time.Duration(authLifeSpan)
	timeNow := time.Now().UTC()
	expiry := jwt.NewNumericDate(timeNow.Add(lifeSpanDuration))
	if family == "" && isRefreshToken {
		family = uuid.New().String()
	} else if !isRefreshToken {
		family = "access"
	}
	claims := GtClaims{
		isAdmin,
		username,
		family,
		jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userID,
			ExpiresAt: expiry,
			Issuer:    "GO-TODO",
			IssuedAt:  jwt.NewNumericDate(timeNow),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	secret := config.JwtAccessSecret
	if isRefreshToken {
		secret = config.JwtRefreshSecret
	}
	encodedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", nil, generateError(err)
	}
	return encodedToken, &claims, nil
}

func GenerateAccessJwt(username string, userID string, isAdmin bool) (string, *GtClaims, error) {
	return generateJwt(username, userID, isAdmin, false, "")
}

func GenerateRefreshJwt(username string, userID string, isAdmin bool, tokenFamily string) (string, *GtClaims, error) {
	return generateJwt(username, userID, isAdmin, true, tokenFamily)
}

// Takes a jwt as a string and boolean isRefreshToken telling should it be
// decoded with refresh secret. If all goes well, returns claims and if not,
// returns JwtValidationError or normal error.
func decodeJwt(tokenString string, isRefreshToken bool) (*GtClaims, error) {
	config, err := config.Get()
	if err != nil {
		return nil, err
	}

	secret := config.JwtAccessSecret
	if isRefreshToken {
		secret = config.JwtRefreshSecret
	}
	decodedToken, err := jwt.ParseWithClaims(tokenString, &GtClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if decodedToken == nil {
			reason := JwtErrorReasonUnhandled
			switch {
			case errors.Is(err, jwt.ErrTokenMalformed):
				reason = JwtErrorReasonTokenMalformed
			}
			return nil, NewJwtDecodeError(nil, reason, err)
		}
		if claims, ok := decodedToken.Claims.(*GtClaims); ok {
			reason := JwtErrorReasonUnhandled
			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				reason = JwtErrorReasonExpired
			case errors.Is(err, jwt.ErrSignatureInvalid):
				reason = JwtErrorReasonInvalidSignature
			default:
				return nil, NewJwtDecodeError(claims, reason, err)
			}
			return nil, NewJwtDecodeError(claims, reason, err)
		}
		return nil, NewJwtDecodeError(nil, JwtErrorReasonUnhandled, err)
	} else if claims, ok := decodedToken.Claims.(*GtClaims); ok {
		return claims, nil
	} else {
		return nil, NewJwtDecodeError(nil, JwtErrorReasonUnhandled, err)
	}
}

func DecodeAccessToken(tokenString string) (*GtClaims, error) {
	return decodeJwt(tokenString, false)
}

func DecodeRefreshToken(tokenString string) (*GtClaims, error) {
	return decodeJwt(tokenString, true)
}

func getTokenFromHeader(c *gin.Context) string {
	bearerToken := c.Request.Header.Get("Authorization")

	splitToken := strings.Split(bearerToken, " ")
	if len(splitToken) == 2 {
		return splitToken[1]
	}
	return ""
}

func DecodeTokenFromHeader(c *gin.Context) (*GtClaims, error) {
	tokenString := getTokenFromHeader(c)
	return DecodeAccessToken(tokenString)
}
