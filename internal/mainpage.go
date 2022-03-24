package internal

import (
	"forum/dbs"
	"forum/models"
	"html/template"
	"log"
	"net/http"
)

// Index handler for main page.
func Index(templ *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(ctxKeyUser).(*models.User)

		if r.Method != http.MethodGet {
			Error(w, templ, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
			return
		}

		var (
			posts []*models.Post
			err   error
		)

		filters, ok := r.URL.Query()["filter"]
		if ok && len(filters[0]) != 0 {
			posts, err = filter(filters[0], user)
			if err != nil {
				if err == errUserNotAuth {
					Error(w, templ, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
				} else {
					Error(w, templ, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
				}
				return
			}

		} else {
			posts, err = dbs.GetRecentPosts(user.ID, 1000, dbs.SimplePost)
			if err != nil {
				log.Println(err)
				Error(w, templ, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
				return
			}

		}

		cats, err := dbs.GetCategories()
		if err != nil {
			log.Println(err)
			Error(w, templ, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}

		page := struct {
			User       *models.User
			Posts      []*models.Post
			Categories []*models.Category
		}{
			User:       user,
			Posts:      posts,
			Categories: cats,
		}

		if err := templ.ExecuteTemplate(w, "home", page); err != nil {
			Error(w, templ, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	})
}
