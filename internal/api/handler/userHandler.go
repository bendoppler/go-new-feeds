package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	_ "news-feed/docs"
	"news-feed/internal/api/model"
	"news-feed/internal/entity"
	"news-feed/internal/service"
	"news-feed/pkg/logger"
	"news-feed/pkg/middleware"
	"time"
)

type UserHandlerInterface interface {
	Login() http.HandlerFunc
	Signup() http.HandlerFunc
	EditProfile() http.HandlerFunc
	UserHandler(w http.ResponseWriter, r *http.Request)
}

// UserHandler handles requests related to users.
type UserHandler struct {
	userService service.UserServiceInterface
}

func (h *UserHandler) UserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Signup()
	case http.MethodPut:
		middleware.JWTAuthMiddleware(h.EditProfile()).ServeHTTP(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Login handles user login.
//
// @Summary User login
// @Description Authenticates a user and returns a token.
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body model.LoginRequest true "User credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 401 {object} string "Unauthorized"
// @Router /v1/users/login [post]
func (h *UserHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials model.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			logger.LogError(fmt.Sprintf("Invalid request body: %v", err))
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		token, err := h.userService.Login(credentials.UserName, credentials.Password)
		if err != nil {
			logger.LogError(fmt.Sprintf("Login failed: %v", err))
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"token": token})
		if err != nil {
			logger.LogError(fmt.Sprintf("Encode failed: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Signup handles user signup.
//
// @Summary User signup
// @Description Registers a new user and returns a token.
// @Tags users
// @Accept json
// @Produce json
// @Param signupRequest body model.SignupRequest true "New user signup information"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/users/signup [post]
func (h *UserHandler) Signup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var signupRequest model.SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&signupRequest); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		birthday, err := h.convertStringToDate(signupRequest.Birthday)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Convert API model to entity model
		newUser := entity.User{
			Username:  signupRequest.UserName,
			Email:     signupRequest.Email,
			FirstName: signupRequest.FirstName,
			LastName:  signupRequest.LastName,
			Birthday:  birthday,
			Password:  signupRequest.Password,
		}

		token, err := h.userService.Signup(newUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"token": token})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// EditProfile handles editing user profile.
//
// @Summary Edit user profile
// @Description Updates the user profile information.
// @Tags users
// @Accept json
// @Produce json
// @Param profileUpdate body model.ProfileUpdateRequest true "Profile update information"
// @Success 200 {object} map[string]string "Profile updated successfully"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/users/profile [put]
func (h *UserHandler) EditProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var profileUpdate model.ProfileUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&profileUpdate); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		birthday, err := h.convertStringToDate(profileUpdate.Birthday)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Convert API model to entity model
		user := entity.User{
			FirstName: profileUpdate.FirstName,
			LastName:  profileUpdate.LastName,
			Birthday:  birthday,
			Password:  profileUpdate.Password,
		}

		err = h.userService.EditProfile(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"msg": "Profile updated successfully"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *UserHandler) convertStringToDate(dateStr string) (time.Time, error) {
	const layout = "2006-01-02" // Define the date layout format
	date, err := time.Parse(layout, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %v", err)
	}
	return date, nil
}
