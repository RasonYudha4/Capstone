package handlers

import (
	"net/http"

	"auth-service/config"
	"auth-service/models"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler groups the HTTP handlers for authentication endpoints.
// Dependencies (User, OTP, and JWT services) are injected through the constructor,
// which makes testing straightforward.
type AuthHandler struct {
	userService *services.UserService
	otpService  *services.OTPService
	jwtService  *services.JWTService
}

// NewAuthHandler creates an AuthHandler with all required services.
func NewAuthHandler(
	userService *services.UserService,
	otpService *services.OTPService,
	jwtService *services.JWTService,
) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		otpService:  otpService,
		jwtService:  jwtService,
	}
}

// =====================================================================
//  POST /auth/login
// =====================================================================

// Login handles email + password authentication.
//
// Flow:
//  1. Parse & validate the request body (email + password).
//  2. Verify credentials against the database (bcrypt comparison).
//  3. If role is "user"         → issue JWT immediately (no OTP).
//  4. If role is "admin" or "master_admin" → generate OTP, return requires_otp: true.
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	// Bind and validate the JSON body.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Authenticate with email + password.
	user, err := h.userService.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal server error.",
		})
		return
	}

	// Generic error message — never reveal whether the email exists.
	if user == nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid email or password.",
		})
		return
	}

	// --- Role-based login branching ---

	// For "staff" role: credentials alone are sufficient.
	if user.Role == config.RoleStaff {
		token, err := h.jwtService.GenerateToken(user.ID, user.Email, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to generate token.",
			})
			return
		}

		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Login successful.",
			Data: models.LoginResponse{
				RequiresOTP: false,
				Token:       token,
				ExpiresIn:   config.JWTExpiration.String(),
			},
		})
		return
	}

	// For "admin" and "master_admin": require OTP as a second factor.
	_, err = h.otpService.GenerateAndStore(user.Email, "login")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate OTP.",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "OTP has been sent to your email. Please verify to complete login.",
		Data: models.LoginResponse{
			RequiresOTP: true,
		},
	})
}

// =====================================================================
//  POST /auth/verify-otp
// =====================================================================

// VerifyOTP handles the second step of admin/master_admin login.
//
// Flow:
//  1. Parse & validate the request body (email + 6-digit OTP).
//  2. Check the OTP against the in-memory store (expiry, attempts, code match).
//  3. If valid → look up user, generate JWT, return token.
//  4. If invalid → return descriptive error.
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.OTPVerifyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Verify the OTP against the in-memory store.
	valid, errMsg, err := h.otpService.Verify(req.Email, req.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal error during OTP verification.",
		})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: errMsg,
		})
		return
	}

	// OTP is valid — look up the user and issue a JWT.
	user, err := h.userService.GetByEmail(req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve user information.",
		})
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate authentication token.",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful.",
		Data: models.TokenResponse{
			Token:     token,
			ExpiresIn: config.JWTExpiration.String(),
		},
	})
}

// =====================================================================
//  GET /auth/me
// =====================================================================

// Me returns the authenticated user's profile information.
// Requires a valid JWT (enforced by the JWTAuth middleware).
func (h *AuthHandler) Me(c *gin.Context) {
	// Extract claims set by the JWTAuth middleware.
	value, exists := c.Get(config.ContextKeyUser)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Not authenticated.",
		})
		return
	}

	claims := value.(*services.Claims)

	// Fetch full user details from the database.
	user, err := h.userService.GetByID(claims.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found.",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User info retrieved.",
		Data: models.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
	})
}
