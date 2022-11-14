package render

import (
	"net/http"
	"testing"

	"github.com/byt3er/bookings/internals/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData
	r, err := getSession()
	if err != nil {
		t.Error(err)
	}
	// put something in the session to test
	session.Put(r.Context(), "flash", "123")
	result := AddDefaultData(&td, r)

	// if result == nil {
	// 	t.Error("failed\\")
	// }
	if result.Flash != "123" {
		t.Error("flash value of 123 not found in session")
	}
}
func TestRenderTemplate(t *testing.T) {
	pathToTemplates = "./../../templates"
	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err) // this will fail the test
	}

	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Error(err) // fail the test
	}

	var ww myWriter

	err = RenderTemplate(&ww, r, "home.page.tmpl", &models.TemplateData{})
	if err != nil {
		t.Error("error writing template to browser")
	}
	// try to render non existing template
	err = RenderTemplate(&ww, r, "non-existent.page.tmpl", &models.TemplateData{})
	if err == nil {
		t.Error("render template that does not exist")
	}
}

func getSession() (*http.Request, error) {
	// create a request; type *http.Request
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}
	// Create the context
	// get the context out of the request object
	ctx := r.Context()
	// put session data in the context
	// r.Header.Get("X-Session") is the part that make it active session
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	// put context back into the request
	r = r.WithContext(ctx)
	return r, nil
}

func TestNewTemplates(t *testing.T) {
	NewTemplates(app)
}

// test for CreateTemplateCache
func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}
