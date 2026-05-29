package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginFormData struct {
	*FormFields
}

func (app *JMcCannBackPackApp) LoginForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")

	if r.Method == http.MethodGet {
		data := LoginFormData{
			FormFields: NewFormFields(),
		}

		app.render(w, "loginform.templ.html", 200, data)
		return
	}

	if r.Method == http.MethodPost {
		app.LoginFormPost(w, r)
		return
	}

	w.Header().Set("Allow", "GET, POST")
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (app *JMcCannBackPackApp) LoginFormPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.serverError(w, err)
		return
	}

	data := LoginFormData{
		FormFields: NewFormFields(),
	}
	data.Values["email"] = r.FormValue("email")
	data.Values["password"] = r.FormValue("password")

	data.ValidateFormFields(FFEmail | FFPassword)
	if data.HasErrors() {
		app.render(w, "loginform.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	user, err := app.DB.GetUserByEmail(data.Values["email"])
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.Values["password"]))
	if err != nil {
		data.Errors["password"] = "incorrect password provided"
		app.render(w, "loginform.templ.html", http.StatusUnauthorized, data)
		return
	}

	accessToken, err := auth.CreateAccessToken(user, app.Config.JWTSecret, SerConstAccessTokenExpiryMinutes)
	if err != nil {
		app.serverError(w, err)
		return
	}

	refreshToken, err := auth.CreateRefreshToken(user, app.Config.JWTSecret, SerConstRefreshTokenExpiryHours)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   app.Config.SecureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   SerConstAccessTokenExpiryMinutes * 60,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   app.Config.SecureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   SerConstRefreshTokenExpiryHours * 3600,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userID, role, err := auth.ExtractClaimsFromToken(cookie.Value, app.Config.JWTSecret)
		if err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			})
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Redirect(w, r, fmt.Sprintf("/refresh?redirect=%s", r.URL.String()), http.StatusTemporaryRedirect)
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), auth.ContextKeyUserID, userID)
		ctx = context.WithValue(ctx, auth.ContextKeyUserRole, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *JMcCannBackPackApp) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userID, _, err := auth.ExtractClaimsFromToken(cookie.Value, app.Config.JWTSecret)
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := app.DB.GetUserById(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	accessToken, err := auth.CreateAccessToken(user, app.Config.JWTSecret, SerConstAccessTokenExpiryMinutes)
	if err != nil {
		app.serverError(w, err)
		return
	}
	refreshToken, err := auth.CreateRefreshToken(user, app.Config.JWTSecret, SerConstRefreshTokenExpiryHours)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   app.Config.SecureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   SerConstAccessTokenExpiryMinutes * 60,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   app.Config.SecureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   SerConstRefreshTokenExpiryHours * 3600,
	})

	redirectTo := r.URL.Query().Get("redirect")
	if redirectTo == "" {
		redirectTo = "/dashboard"
	}
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}
