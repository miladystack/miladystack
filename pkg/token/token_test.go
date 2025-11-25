package token

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/metadata"
)

// TestInitAndReset 测试配置初始化和重置功能
func TestInitAndReset(t *testing.T) {
	// 保存原始配置以在测试后恢复
	originalConfig := config
	// once 是 sync.Once 类型，包含 noCopy 字段，不能直接赋值
	// 这里仅保存其状态，后续通过 Reset() 恢复
	defer func() {
		config = originalConfig
	}()

	// 测试初始化
	key := "test-secret-key"
	identityKey := "user_id"
	expiration := 1 * time.Hour
	refreshTokenExpiration := 24 * time.Hour
	skipPaths := []string{"/test", "/skip/*"}

	Init(key,
		WithIdentityKey(identityKey),
		WithExpiration(expiration),
		WithRefreshTokenExpiration(refreshTokenExpiration),
		WithSkipPaths(skipPaths...),
	)

	cfg := GetConfig()
	if cfg.key != key {
		t.Errorf("Expected key %s, got %s", key, cfg.key)
	}
	if cfg.identityKey != identityKey {
		t.Errorf("Expected identityKey %s, got %s", identityKey, cfg.identityKey)
	}
	if cfg.expiration != expiration {
		t.Errorf("Expected expiration %v, got %v", expiration, cfg.expiration)
	}
	if cfg.refreshTokenExpiration != refreshTokenExpiration {
		t.Errorf("Expected refreshTokenExpiration %v, got %v", refreshTokenExpiration, cfg.refreshTokenExpiration)
	}
	if len(cfg.skipPaths) != len(skipPaths) {
		t.Errorf("Expected skipPaths length %d, got %d", len(skipPaths), len(cfg.skipPaths))
	}

	// 测试重置
	Reset()
	resetCfg := GetConfig()
	if resetCfg.key != "Rtg8BPKNEf2mB4mgvKONGPZZQSaJWNLijxR42qRgq0iBb5" {
		t.Errorf("Expected default key after reset")
	}
	if resetCfg.refreshTokenExpiration != 7*24*time.Hour {
		t.Errorf("Expected default refreshTokenExpiration 7 days after reset")
	}
}

// TestPathMatching 测试路径匹配功能
func TestPathMatching(t *testing.T) {
	Reset()
	Init("test-key",
		WithSkipPaths("/exact", "/prefix/*", "/wild*card"),
		WithCommonSkipPaths(),
	)

	testCases := []struct {
		path     string
		expected bool
	}{
		{path: "/exact", expected: true},
		{path: "/exact/sub", expected: false},
		{path: "/prefix", expected: true}, // 前缀路径本身也匹配
		{path: "/prefix/sub", expected: true},
		{path: "/prefix/sub/123", expected: true},
		{path: "/wildcard", expected: true},
		{path: "/wildcard/sub", expected: false}, // 通配符模式匹配整个路径段
		{path: "/health", expected: true},
		{path: "/metrics", expected: true},
		{path: "/normal/path", expected: false},
	}

	for _, tc := range testCases {
		result := IsPathSkipped(tc.path)
		if result != tc.expected {
			t.Errorf("Path %s: expected %v, got %v", tc.path, tc.expected, result)
		}
	}
}

// TestWildcardMatching 测试通配符匹配函数
func TestWildcardMatching(t *testing.T) {
	cases := []struct {
		str     string
		pattern string
		expect  bool
	}{
		{str: "test", pattern: "test", expect: true},
		{str: "test/abc", pattern: "test/*", expect: true},
		{str: "testabc", pattern: "test*", expect: true},
		{str: "test/middle/end", pattern: "test/*/end", expect: true},
		{str: "test/abc/xyz/end", pattern: "test/*/xyz/*", expect: true},
		{str: "test/different/end", pattern: "test/*/xyz", expect: false},
		{str: "short", pattern: "long*", expect: false},
	}

	for i, tc := range cases {
		result := matchWildcard(tc.str, tc.pattern)
		if result != tc.expect {
			t.Errorf("Case %d: matchWildcard(%q, %q) = %v; want %v", i, tc.str, tc.pattern, result, tc.expect)
		}
	}
}

// TestSign 测试Token签发功能
func TestSign(t *testing.T) {
	Reset()
	key := "test-secret-key"
	identityKey := "user_id"
	Init(key, WithIdentityKey(identityKey))

	// 测试正常签发
	identityValue := "test-user-123"
	tokenString, expireAt, err := Sign(identityValue)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	if tokenString == "" {
		t.Fatal("Sign returned empty token")
	}
	if expireAt.IsZero() {
		t.Fatal("Sign returned zero expiration time")
	}

	// 验证签发的token
	parsedIdentity, err := ParseIdentity(tokenString, key)
	if err != nil {
		t.Fatalf("ParseIdentity failed: %v", err)
	}
	if parsedIdentity != identityValue {
		t.Errorf("Expected identity %s, got %s", identityValue, parsedIdentity)
	}

	// 测试使用自定义claims签发
	customClaims := jwt.MapClaims{
		"custom_field": "custom_value",
		"number_field": 123,
	}
	customToken, _, err := SignWithClaims(customClaims)
	if err != nil {
		t.Fatalf("SignWithClaims failed: %v", err)
	}

	claims, err := GetClaims(customToken)
	if err != nil {
		t.Fatalf("GetClaims failed: %v", err)
	}
	if val, ok := claims["custom_field"]; !ok || val != "custom_value" {
		t.Errorf("Expected custom_field in claims")
	}
	if val, ok := claims["number_field"]; !ok || val.(float64) != 123 {
		t.Errorf("Expected number_field in claims")
	}
}

// TestSignTokens 测试同时签发 access token 和 refresh token
func TestSignTokens(t *testing.T) {
	Reset()
	key := "test-secret-key"
	identityKey := "user_id"
	Init(key, WithIdentityKey(identityKey))

	// 测试正常签发
	identityValue := "test-user-123"
	tokenPair, err := SignTokens(identityValue)
	if err != nil {
		t.Fatalf("SignTokens failed: %v", err)
	}

	// 验证 access token
	if tokenPair.AccessToken == "" {
		t.Fatal("access token is empty")
	}
	if tokenPair.AccessExpireAt.IsZero() {
		t.Fatal("access token expiration is zero")
	}

	// 验证 refresh token
	if tokenPair.RefreshToken == "" {
		t.Fatal("refresh token is empty")
	}
	if tokenPair.RefreshExpireAt.IsZero() {
		t.Fatal("refresh token expiration is zero")
	}

	// 验证 refresh token 过期时间比 access token 长
	if tokenPair.RefreshExpireAt.Before(tokenPair.AccessExpireAt) {
		t.Error("refresh token should expire later than access token")
	}

	// 验证 access token 的类型
	accessClaims, err := GetClaims(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("GetClaims for access token failed: %v", err)
	}
	if tokenType, ok := accessClaims["type"].(string); !ok || tokenType != "access" {
		t.Error("access token should have type 'access'")
	}

	// 验证 refresh token 的类型
	refreshClaims, err := GetClaims(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("GetClaims for refresh token failed: %v", err)
	}
	if tokenType, ok := refreshClaims["type"].(string); !ok || tokenType != "refresh" {
		t.Error("refresh token should have type 'refresh'")
	}
}

// TestRefreshTokens 测试刷新 token 功能
func TestRefreshTokens(t *testing.T) {
	Reset()
	key := "test-secret-key"
	identityKey := "user_id"
	Init(key, WithIdentityKey(identityKey))

	// 签发初始 tokens
	identityValue := "test-user-123"
	tokenPair, err := SignTokens(identityValue)
	if err != nil {
		t.Fatalf("SignTokens failed: %v", err)
	}

	// 刷新 tokens
	newTokenPair, err := RefreshTokens(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshTokens failed: %v", err)
	}

	// 验证新 tokens
	if newTokenPair.AccessToken == "" {
		t.Fatal("new access token is empty")
	}
	if newTokenPair.AccessExpireAt.IsZero() {
		t.Fatal("new access token expiration is zero")
	}
	if newTokenPair.RefreshToken == "" {
		t.Fatal("new refresh token is empty")
	}
	if newTokenPair.RefreshExpireAt.IsZero() {
		t.Fatal("new refresh token expiration is zero")
	}

	// 验证新 access token 的身份信息
	parsedIdentity, err := ParseIdentity(newTokenPair.AccessToken, key)
	if err != nil {
		t.Fatalf("ParseIdentity failed: %v", err)
	}
	if parsedIdentity != identityValue {
		t.Errorf("Expected identity %s, got %s", identityValue, parsedIdentity)
	}
}

// TestRefreshTokensWithInvalidTokens 测试使用无效的 refresh token
func TestRefreshTokensWithInvalidTokens(t *testing.T) {
	Reset()
	key := "test-secret-key"
	identityKey := "user_id"
	Init(key, WithIdentityKey(identityKey))

	// 测试用例
	cases := []struct {
		token      string
		expectErr  bool
		errMessage string
	}{
		{"", true, ErrEmptyToken.Error()}, // 空 token
		{"invalid-token", true, ""},       // 无效格式
		{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.", true, ""}, // 不完整的 token
	}

	for i, tc := range cases {
		_, err := RefreshTokens(tc.token)
		if !tc.expectErr && err != nil {
			t.Errorf("Case %d: Expected no error, got %v", i, err)
		}
		if tc.expectErr && err == nil {
			t.Errorf("Case %d: Expected error, got nil", i)
		}
		if tc.expectErr && tc.errMessage != "" && err != nil && err.Error() != tc.errMessage {
			t.Errorf("Case %d: Expected error message '%s', got '%s'", i, tc.errMessage, err.Error())
		}
	}

	// 测试使用 access token 作为 refresh token
	_, _, err := Sign("test-user")
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// 测试使用错误密钥签发的 token
	wrongKeyToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf":  time.Now().Unix(),
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"type": "refresh",
	}).SignedString([]byte("wrong-key"))
	if err != nil {
		t.Fatalf("Failed to create wrong key token: %v", err)
	}

	_, err = RefreshTokens(wrongKeyToken)
	if err == nil {
		t.Error("Expected error when using token with wrong key")
	}
}

// TestConfigAccessorsWithRefreshToken 测试配置访问函数，包括 refresh token 相关
func TestConfigAccessorsWithRefreshToken(t *testing.T) {
	Reset()
	refreshExpiration := 24 * time.Hour
	Init("test-key",
		WithIdentityKey("user_id"),
		WithExpiration(1*time.Hour),
		WithRefreshTokenExpiration(refreshExpiration),
	)

	if GetRefreshTokenExpiration() != refreshExpiration {
		t.Errorf("Expected refresh token expiration %v, got %v", refreshExpiration, GetRefreshTokenExpiration())
	}
}

// TestTokenParsing 测试Token解析功能
func TestTokenParsing(t *testing.T) {
	Reset()
	key := "test-secret-key"
	identityKey := "user_id"
	Init(key, WithIdentityKey(identityKey))

	// 签发测试token
	identityValue := "test-user-123"
	tokenString, _, err := Sign(identityValue)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// 测试正常解析
	parsedIdentity, err := ParseIdentity(tokenString, key)
	if err != nil {
		t.Fatalf("ParseIdentity failed: %v", err)
	}
	if parsedIdentity != identityValue {
		t.Errorf("Expected identity %s, got %s", identityValue, parsedIdentity)
	}

	// 测试解析失败的情况
	cases := []struct {
		token     string
		expectErr bool
	}{
		{token: "", expectErr: true},
		{token: "invalid-token", expectErr: true},
		{token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.", expectErr: true},
	}

	for i, tc := range cases {
		_, err := ParseIdentity(tc.token, key)
		if (err != nil) != tc.expectErr {
			t.Errorf("Case %d: expected error %v, got %v", i, tc.expectErr, err != nil)
		}
	}

	// 测试使用错误密钥解析
	_, err = ParseIdentity(tokenString, "wrong-key")
	if err == nil {
		t.Error("Expected error when parsing with wrong key")
	}
}

// TestGetClaims 测试获取Token中的claims
func TestGetClaims(t *testing.T) {
	Reset()
	key := "test-secret-key"
	Init(key)

	// 签发测试token
	customClaims := jwt.MapClaims{
		"custom1": "value1",
		"custom2": 42,
	}
	tokenString, _, err := SignWithClaims(customClaims)
	if err != nil {
		t.Fatalf("SignWithClaims failed: %v", err)
	}

	// 获取并验证claims
	claims, err := GetClaims(tokenString)
	if err != nil {
		t.Fatalf("GetClaims failed: %v", err)
	}

	// 验证必要的时间字段
	if _, ok := claims["nbf"]; !ok {
		t.Error("Missing nbf claim")
	}
	if _, ok := claims["iat"]; !ok {
		t.Error("Missing iat claim")
	}
	if _, ok := claims["exp"]; !ok {
		t.Error("Missing exp claim")
	}

	// 验证自定义字段
	if val, ok := claims["custom1"]; !ok || val != "value1" {
		t.Errorf("Expected custom1=value1, got %v", val)
	}
	if val, ok := claims["custom2"]; !ok || val.(float64) != 42 {
		t.Errorf("Expected custom2=42, got %v", val)
	}
}

// TestGinContextParsing 测试从Gin上下文解析Token
func TestGinContextParsing(t *testing.T) {
	Reset()
	key := "test-secret-key"
	identityKey := "user_id"
	Init(key, WithIdentityKey(identityKey))

	// 签发测试token
	identityValue := "test-user-123"
	tokenString, _, err := Sign(identityValue)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// 测试带有有效token的请求
	func() {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = req

		parsedIdentity, err := ParseRequest(ctx)
		if err != nil {
			t.Fatalf("ParseRequest failed: %v", err)
		}
		if parsedIdentity != identityValue {
			t.Errorf("Expected identity %s, got %s", identityValue, parsedIdentity)
		}
	}()

	// 测试跳过路径
	func() {
		Reset()
		Init(key, WithIdentityKey(identityKey), WithSkipPaths("/health"))

		req, _ := http.NewRequest("GET", "/health", nil)
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = req

		parsedIdentity, err := ParseRequest(ctx)
		if err != nil {
			t.Fatalf("ParseRequest failed: %v", err)
		}
		if parsedIdentity != "" {
			t.Errorf("Expected empty identity for skipped path, got %s", parsedIdentity)
		}
	}()

	// 测试没有token的请求
	func() {
		Reset()
		Init(key, WithIdentityKey(identityKey))

		req, _ := http.NewRequest("GET", "/test", nil)
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = req

		_, err := ParseRequest(ctx)
		if err != ErrEmptyAuthHeader {
			t.Errorf("Expected ErrEmptyAuthHeader, got %v", err)
		}
	}()

	// 测试格式错误的token
	func() {
		Reset()
		Init(key, WithIdentityKey(identityKey))

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat "+tokenString)
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = req

		_, err := ParseRequest(ctx)
		if err != ErrMalformedAuthHeader {
			t.Errorf("Expected ErrMalformedAuthHeader, got %v", err)
		}
	}()
}

// TestGRPCContextParsing 测试从gRPC上下文解析Token
func TestGRPCContextParsing(t *testing.T) {
	Reset()
	key := "test-secret-key"
	identityKey := "user_id"
	Init(key, WithIdentityKey(identityKey))

	// 签发测试token
	identityValue := "test-user-123"
	tokenString, _, err := Sign(identityValue)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// 创建带有有效token的gRPC上下文
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + tokenString,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// 解析请求
	parsedIdentity, err := ParseRequest(ctx)
	if err != nil {
		t.Fatalf("ParseRequest for gRPC failed: %v", err)
	}
	if parsedIdentity != identityValue {
		t.Errorf("Expected identity %s, got %s", identityValue, parsedIdentity)
	}

	// 测试没有token的gRPC上下文
	emptyCtx := context.Background()
	_, err = ParseRequest(emptyCtx)
	if err == nil {
		t.Error("Expected error for gRPC context without token")
	}
}

// TestConfigAccessors 测试配置访问函数
func TestConfigAccessors(t *testing.T) {
	Reset()
	Init("test-key",
		WithIdentityKey("user_id"),
		WithExpiration(3*time.Hour),
		WithSkipPaths("/test1", "/test2"),
	)

	if !IsIdentityRequired() {
		t.Error("Expected identity required to be true")
	}

	if GetExpiration() != 3*time.Hour {
		t.Errorf("Expected expiration 3h, got %v", GetExpiration())
	}

	skipPaths := GetSkipPaths()
	if len(skipPaths) != 2 {
		t.Errorf("Expected 2 skip paths, got %d", len(skipPaths))
	}

	// 测试修改返回的跳过路径副本不会影响原始配置
	skipPaths[0] = "/modified"
	originalSkipPaths := GetSkipPaths()
	if originalSkipPaths[0] == "/modified" {
		t.Error("Original skip paths should not be modified")
	}
}

// TestNoIdentityRequired 测试无身份验证模式
func TestNoIdentityRequired(t *testing.T) {
	Reset()
	Init("test-key", WithIdentityKey("")) // 设置空身份键以禁用身份验证

	// 签发token时不提供身份值
	tokenString, _, err := Sign("")
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// 解析token - 应该返回空身份
	parsedIdentity, err := ParseIdentity(tokenString, "test-key")
	if err != nil {
		t.Fatalf("ParseIdentity failed: %v", err)
	}
	if parsedIdentity != "" {
		t.Errorf("Expected empty identity, got %s", parsedIdentity)
	}

	if IsIdentityRequired() {
		t.Error("Expected identity required to be false")
	}
}

// TestParseWithKey 测试使用自定义密钥解析Token
func TestParseWithKey(t *testing.T) {
	Reset()
	customKey := "custom-secret-key"

	// 使用自定义密钥签发token
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-subject",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(customKey))
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	// 使用ParseWithKey解析
	claims, err := ParseWithKey(tokenString, customKey)
	if err != nil {
		t.Fatalf("ParseWithKey failed: %v", err)
	}

	if val, ok := claims["sub"]; !ok || val != "test-subject" {
		t.Errorf("Expected sub=test-subject, got %v", val)
	}

	// 测试使用错误密钥解析
	_, err = ParseWithKey(tokenString, "wrong-key")
	if err == nil {
		t.Error("Expected error when parsing with wrong key")
	}
}

// TestParseRequestIgnoreSkip 测试忽略跳过路径的解析
func TestParseRequestIgnoreSkip(t *testing.T) {
	Reset()
	key := "test-secret-key"
	identityKey := "user_id"
	Init(key,
		WithIdentityKey(identityKey),
		WithSkipPaths("/health"),
	)

	// 签发测试token
	identityValue := "test-user-123"
	tokenString, _, err := Sign(identityValue)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// 创建到跳过路径的请求
	req, _ := http.NewRequest("GET", "/health", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	// 使用忽略跳过路径的解析方法
	parsedIdentity, err := ParseRequestIgnoreSkip(ctx)
	if err != nil {
		t.Fatalf("ParseRequestIgnoreSkip failed: %v", err)
	}
	if parsedIdentity != identityValue {
		t.Errorf("Expected identity %s, got %s", identityValue, parsedIdentity)
	}
}
