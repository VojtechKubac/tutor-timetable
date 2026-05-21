package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const TeacherIDKey contextKey = "teacher_id"

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				http.Error(w, `{"error":"UNAUTHORIZED"}`, http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"UNAUTHORIZED"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"UNAUTHORIZED"}`, http.StatusUnauthorized)
				return
			}

			teacherID, ok := claims["sub"].(string)
			if !ok || teacherID == "" {
				http.Error(w, `{"error":"UNAUTHORIZED"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), TeacherIDKey, teacherID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TeacherID(r *http.Request) string {
	id, _ := r.Context().Value(TeacherIDKey).(string)
	return id
}
