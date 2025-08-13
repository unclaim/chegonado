package api

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/unclaim/chegonado/internal/auth/domain"
)

func TestNewAuthHandler(t *testing.T) {
	type args struct {
		authService domain.AuthServicePort
	}
	tests := []struct {
		name string
		args args
		want *AuthHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAuthHandler(tt.args.authService); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAuthHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthHandler_SendPasswordResetHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.SendPasswordResetHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_Signup(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.Signup(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_Signout(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.Signout(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.Login(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_CheckSessionHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.CheckSessionHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_GetActiveUserSessionsHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.GetActiveUserSessionsHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_RevokeSessionHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.RevokeSessionHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_ResetPasswordHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.ResetPasswordHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_ResetPasswordPageHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.ResetPasswordPageHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_PasswordHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.PasswordHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_SendEmailCodeForLoginHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.SendEmailCodeForLoginHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_VerifyEmailCodeForLoginHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.VerifyEmailCodeForLoginHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_SendEmailCodeForSignupHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.SendEmailCodeForSignupHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_VerifyEmailCodeForSignupHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.VerifyEmailCodeForSignupHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestAuthHandler_ResendVerificationCodeHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ah   *AuthHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ah.ResendVerificationCodeHandler(tt.args.w, tt.args.r)
		})
	}
}
