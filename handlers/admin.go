package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

var templates map[string]*template.Template

// InitTemplates pre-caches all admin templates at startup in production mode.
func InitTemplates() {
	if os.Getenv("ENV") != "production" {
		return
	}

	templates = make(map[string]*template.Template)
	files, err := os.ReadDir("views")
	if err != nil {
		log.Fatalf("Error reading views directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || file.Name() == "layout.html" || !strings.HasSuffix(file.Name(), ".html") {
			continue
		}

		t, err := template.ParseFiles("views/layout.html", "views/"+file.Name())
		if err != nil {
			log.Fatalf("Error parsing template %s: %v", file.Name(), err)
		}
		templates[file.Name()] = t
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	if os.Getenv("ENV") != "production" || templates == nil {
		t, err := template.ParseFiles("views/layout.html", "views/"+tmpl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.ExecuteTemplate(w, "layout.html", data)
		return
	}

	t, ok := templates[tmpl]
	if !ok {
		http.Error(w, "Template "+tmpl+" not found in cache", http.StatusInternalServerError)
		return
	}
	t.ExecuteTemplate(w, "layout.html", data)
}
