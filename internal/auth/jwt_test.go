package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtTestSecret = "test-secret"

func TestGenerateAndVerifyJWTTokenSuccess(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", jwtTestSecret)

	token, err := GenerateJWTToken(42, "alice", "admin")
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateJWTToken() returned empty token")
	}

	claims, err := VerifyJWTToken(token)
	if err != nil {
		t.Fatalf("VerifyJWTToken() error = %v", err)
	}

	if claims.ID != 42 {
		t.Fatalf("claims.ID = %d, want 42", claims.ID)
	}
	if claims.Username != "alice" {
		t.Fatalf("claims.Username = %q, want alice", claims.Username)
	}
	if claims.Role != "admin" {
		t.Fatalf("claims.Role = %q, want admin", claims.Role)
	}
	if claims.ExpiresAt == nil {
		t.Fatal("claims.ExpiresAt is nil")
	}
}

func TestVerifyJWTTokenExpiredToken(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", jwtTestSecret)

	expiredClaims := &JWTTokenClaims{
		ID:       1,
		Username: "bob",
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte(jwtTestSecret))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	if _, err := VerifyJWTToken(token); err == nil {
		t.Fatal("VerifyJWTToken() expected error for expired token, got nil")
	}
}

func TestVerifyJWTTokenWrongSigningKey(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "signing-key-a")

	claims := &JWTTokenClaims{
		ID:       7,
		Username: "eve",
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("signing-key-b"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if _, err := VerifyJWTToken(token); err == nil {
		t.Fatal("VerifyJWTToken() expected signature error, got nil")
	}
}

func TestVerifyJWTTokenRejectsNonHMACMethod(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", jwtTestSecret)

	noneToken := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"id":       1,
		"username": "mallory",
		"role":     "user",
		"exp":      time.Now().Add(10 * time.Minute).Unix(),
	})
	token, err := noneToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to sign none token: %v", err)
	}

	if _, err := VerifyJWTToken(token); err == nil {
		t.Fatal("VerifyJWTToken() expected error for non-HMAC method, got nil")
	}
}

func TestRefreshJWTokenUpdatesExpiry(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", jwtTestSecret)

	base := time.Now().Add(1 * time.Minute)
	claims := &JWTTokenClaims{
		ID:       9,
		Username: "refresher",
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(base),
		},
	}

	refreshed, err := RefreshJWToken(claims)
	if err != nil {
		t.Fatalf("RefreshJWToken() error = %v", err)
	}

	newClaims, err := VerifyJWTToken(refreshed)
	if err != nil {
		t.Fatalf("VerifyJWTToken(refreshed) error = %v", err)
	}
	if newClaims.ExpiresAt == nil {
		t.Fatal("newClaims.ExpiresAt is nil")
	}
	if !newClaims.ExpiresAt.Time.After(base) {
		t.Fatalf("new expiry %v is not after original %v", newClaims.ExpiresAt.Time, base)
	}
}
