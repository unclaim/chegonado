package index

import (
	"net/http"

	"github.com/unclaim/chegonado/pkg/security/session"
)

func ServeIndexPage(w http.ResponseWriter, r *http.Request) {
	_, err := session.SessionFromContext(r.Context())
	if err != nil {
		http.Redirect(w, r, "/api/user/login", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/private/", http.StatusFound)
}
