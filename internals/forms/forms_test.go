package forms

import (
	"net/url"
	"testing"
)

func TestForm_valid(t *testing.T) {
	postedData := url.Values{}
	form := New(postedData) // create  a new form object
	isValid := form.Valid()

	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}

func TestForm_Required(t *testing.T) {
	postData := url.Values{}
	form := New(postData) //url.Values
	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required feilds are missing")
	}

	postData = url.Values{}
	postData.Add("a", "a")
	postData.Add("b", "a")
	postData.Add("c", "a")

	form = New(postData)

	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("shows does not have required fields when it does")
	}
}
func TestForm_Has(t *testing.T) {

	formData := url.Values{}
	formData.Add("a", "a")
	formData.Add("b", "a")

	form := New(formData)

	if !form.Has("a") {
		t.Error("has value but showing empty")
	}

	if form.Has("d") {
		t.Error("doesn't has value, empty but showing not empty")
	}

}

func TestForm_MinLength(t *testing.T) {
	postedValues := url.Values{}
	form := New(postedValues)

	form.MinLength("x", 10)
	if form.Valid() {
		t.Error("form show min length for non-existing field")
	}
	isError := form.Errors.Get("x")
	if isError == "" {
		t.Error("should have an error but not found one")
	}

	postedValues = url.Values{}
	postedValues.Add("some_field", "some value")

	form = New(postedValues)
	form.MinLength("some_field", 100)
	if form.Valid() {
		t.Error("shows minlength of 100 met when data is shorter")
	}

	postedValues = url.Values{}
	postedValues.Add("another_field", "abc123")
	form = New(postedValues)
	form.MinLength("another_field", 1)

	if !form.Valid() {
		t.Error("show minlenght of 1 is not met when it is")
	}
	isError = form.Errors.Get("another_field")
	if isError != "" {
		t.Error("should not have an error, but got one ")
	}

	// formData := url.Values{}
	// formData.Add("a", "aaaaa")
	// formData.Add("b", "bb")
	// form := New(formData)

	// r := httptest.NewRequest("POST", "/some-url", nil)
	// r.PostForm = formData

	// form.MinLength("a", 5, r)
	// if !form.Valid() {
	// 	t.Error("minimum length value is provided but still the Minlenth() return false")
	// }

}

func TestForm_IsEmail(t *testing.T) {
	postedValues := url.Values{}
	form := New(postedValues)

	form.IsEmail("x")
	if form.Valid() {
		t.Error("form shows valid email for non-existent field")
	}

	postedValues = url.Values{}
	postedValues.Add("email", "me@here.com")
	form = New(postedValues)
	form.IsEmail("email")
	if !form.Valid() {
		t.Error("got invalid email when we should not have")
	}
	postedValues = url.Values{}
	postedValues.Add("email", "me@")
	form = New(postedValues)
	form.IsEmail("email")
	if form.Valid() {
		t.Error("got valid for invalid email")
	}
	// formData := url.Values{}
	// formData.Add("email1", "m@gmail.com")
	// formData.Add("email2", "a@gmail")
	// formData.Add("email3", "abc")

	// form := New(formData)

	// if !form.IsEmail("email1") {
	// 	t.Error("showing valid email as invalid email")
	// }
	// if form.IsEmail("email2") {
	// 	t.Error("showing invalid email as valid email")
	// }

}
