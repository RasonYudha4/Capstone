package handlers

import (
	"net/http"
	"strings"

	"auth-service/config"

	"github.com/gin-gonic/gin"
)

const (
	cookieAccessToken  = "access_token"
	cookieRefreshToken = "refresh_token"
)

func cookieSameSite() http.SameSite {
	switch strings.ToLower(config.CookieSameSite) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	sameSite := cookieSameSite()
	secure := config.CookieSecure || sameSite == http.SameSiteNoneMode

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieAccessToken,
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(config.AccessTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieRefreshToken,
		Value:    refreshToken,
		Path:     "/auth",
		MaxAge:   int(config.RefreshTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func clearAuthCookies(c *gin.Context) {
	sameSite := cookieSameSite()
	secure := config.CookieSecure || sameSite == http.SameSiteNoneMode

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieAccessToken,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieRefreshToken,
		Value:    "",
		Path:     "/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

// refreshTokenFromRequest prefers the HttpOnly cookie, then optional JSON body.
func refreshTokenFromRequest(c *gin.Context) string {
	if token, err := c.Cookie(cookieRefreshToken); err == nil && token != "" {
		return token
	}

	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	// Empty body is fine (cookie-only clients).
	_ = c.ShouldBindJSON(&body)
	return body.RefreshToken
}
