package ui

import (
	"strings"
	"testing"
)

func TestGetStaticFile_CSS(t *testing.T) {
	content, contentType, found := GetStaticFile("css/style.css")

	if !found {
		t.Fatal("Expected CSS file to be found")
	}

	if contentType != "text/css" {
		t.Errorf("Expected content type 'text/css', got: %s", contentType)
	}

	if len(content) == 0 {
		t.Error("Expected non-empty CSS content")
	}

	cssString := string(content)
	if !strings.Contains(cssString, "body") {
		t.Error("Expected CSS to contain body styles")
	}
	if !strings.Contains(cssString, ":root") {
		t.Error("Expected CSS to contain CSS variables")
	}
	if !strings.Contains(cssString, "--primary-color") {
		t.Error("Expected CSS to contain primary color variable")
	}
}

func TestGetStaticFile_JS(t *testing.T) {
	content, contentType, found := GetStaticFile("js/app.js")

	if !found {
		t.Fatal("Expected JS file to be found")
	}

	if contentType != "application/javascript" {
		t.Errorf("Expected content type 'application/javascript', got: %s", contentType)
	}

	if len(content) == 0 {
		t.Error("Expected non-empty JS content")
	}

	jsString := string(content)
	if !strings.Contains(jsString, "function") {
		t.Error("Expected JS to contain functions")
	}
	if !strings.Contains(jsString, "loadList") {
		t.Error("Expected JS to contain loadList function")
	}
	if !strings.Contains(jsString, "handleForm") {
		t.Error("Expected JS to contain handleForm function")
	}
}

func TestGetStaticFile_NotFound(t *testing.T) {
	_, _, found := GetStaticFile("nonexistent/file.txt")

	if found {
		t.Error("Expected non-existent file to not be found")
	}
}

func TestGetStaticFile_DefaultContentType(t *testing.T) {
	_, contentType, found := GetStaticFile("unknown.xyz")

	if found {
		if contentType != "text/plain" {
			t.Errorf("Expected default content type 'text/plain', got: %s", contentType)
		}
	}
}

func TestGetStaticFiles(t *testing.T) {
	files := getStaticFiles()

	if len(files) == 0 {
		t.Error("Expected static files to be returned")
	}

	// Check that expected files exist
	if _, exists := files["css/style.css"]; !exists {
		t.Error("Expected css/style.css to exist")
	}

	if _, exists := files["js/app.js"]; !exists {
		t.Error("Expected js/app.js to exist")
	}
}

func TestGetCSS_Content(t *testing.T) {
	css := getCSS()

	if css == "" {
		t.Fatal("Expected non-empty CSS")
	}

	// Check for essential CSS components
	essentialRules := []string{
		"body",
		":root",
		".btn",
		".form-control",
		".sidebar",
		".main-content",
		"--primary-color",
		"--gray-",
		"@media",
		"@keyframes",
		".data-table",
		".pagination",
	}

	for _, rule := range essentialRules {
		if !strings.Contains(css, rule) {
			t.Errorf("Expected CSS to contain '%s'", rule)
		}
	}

	// Check for responsive design
	if !strings.Contains(css, "@media (max-width: 768px)") {
		t.Error("Expected CSS to contain mobile breakpoint")
	}

	// Check for CSS custom properties
	if !strings.Contains(css, "--primary-gradient") {
		t.Error("Expected CSS to contain custom gradient property")
	}
}

func TestGetJS_Content(t *testing.T) {
	js := getJS()

	if js == "" {
		t.Fatal("Expected non-empty JavaScript")
	}

	// Check for essential JavaScript functions
	essentialFunctions := []string{
		"loadList",
		"renderTable",
		"renderPagination",
		"handleForm",
		"deleteRecord",
		"showError",
		"formatFieldValue",
	}

	for _, fn := range essentialFunctions {
		if !strings.Contains(js, fn) {
			t.Errorf("Expected JavaScript to contain function '%s'", fn)
		}
	}

	// Check for API integration
	if !strings.Contains(js, "API_BASE") {
		t.Error("Expected JavaScript to contain API_BASE constant")
	}

	// Check for event listeners
	if !strings.Contains(js, "DOMContentLoaded") {
		t.Error("Expected JavaScript to contain DOMContentLoaded listener")
	}
	if !strings.Contains(js, "addEventListener") {
		t.Error("Expected JavaScript to contain event listeners")
	}

	// Check for fetch API usage
	if !strings.Contains(js, "fetch") {
		t.Error("Expected JavaScript to contain fetch calls")
	}

	// Check for form handling
	if !strings.Contains(js, "FormData") {
		t.Error("Expected JavaScript to contain FormData usage")
	}

	// Check for pagination variables
	if !strings.Contains(js, "currentPage") {
		t.Error("Expected JavaScript to contain pagination variables")
	}
	if !strings.Contains(js, "currentSearch") {
		t.Error("Expected JavaScript to contain search variables")
	}

	// Check for error handling
	if !strings.Contains(js, "catch") {
		t.Error("Expected JavaScript to contain error handling")
	}
}

func TestCSS_ResponsiveDesign(t *testing.T) {
	css := getCSS()

	// Check for mobile-first responsive design
	responsiveRules := []string{
		"@media (max-width: 768px)",
		"flex-direction: column",
		"width: 100%",
		"grid-template-columns: 1fr",
	}

	for _, rule := range responsiveRules {
		if !strings.Contains(css, rule) {
			t.Errorf("Expected CSS to contain responsive rule '%s'", rule)
		}
	}
}

func TestCSS_Animations(t *testing.T) {
	css := getCSS()

	// Check for animations and transitions
	animationRules := []string{
		"@keyframes",
		"transition:",
		"transform:",
		"animation:",
		"fadeIn",
		"spin",
	}

	for _, rule := range animationRules {
		if !strings.Contains(css, rule) {
			t.Errorf("Expected CSS to contain animation rule '%s'", rule)
		}
	}
}

func TestJS_FormHandling(t *testing.T) {
	js := getJS()

	// Check for comprehensive form handling
	formRules := []string{
		"checkbox",
		"password",
		"number",
		"date",
		"datetime-local",
		"time",
		"parseFloat",
		"form.elements",
	}

	for _, rule := range formRules {
		if !strings.Contains(js, rule) {
			t.Errorf("Expected JavaScript to contain form handling for '%s'", rule)
		}
	}
}

func TestJS_FieldFormatting(t *testing.T) {
	js := getJS()

	// Check for field value formatting functions
	if !strings.Contains(js, "formatFieldValue") {
		t.Error("Expected JavaScript to contain formatFieldValue function")
	}

	// Check for specific field type handling
	fieldTypes := []string{
		"password",
		"boolean",
		"date",
		"options",
	}

	for _, fieldType := range fieldTypes {
		if !strings.Contains(js, fieldType) {
			t.Errorf("Expected JavaScript to contain handling for '%s' fields", fieldType)
		}
	}
}

func TestJS_TableRendering(t *testing.T) {
	js := getJS()

	// Check for table rendering functions
	tableRules := []string{
		"renderTable",
		"tbody",
		"createElement",
		"appendChild",
		"textContent",
		"empty-state",
		"actions",
	}

	for _, rule := range tableRules {
		if !strings.Contains(js, rule) {
			t.Errorf("Expected JavaScript to contain table rendering rule '%s'", rule)
		}
	}
}

func TestJS_PaginationHandling(t *testing.T) {
	js := getJS()

	// Check for pagination functionality
	paginationRules := []string{
		"renderPagination",
		"Previous",
		"Next",
		"total_pages",
		"startPage",
		"endPage",
		"active",
		"disabled",
	}

	for _, rule := range paginationRules {
		if !strings.Contains(js, rule) {
			t.Errorf("Expected JavaScript to contain pagination rule '%s'", rule)
		}
	}
}

func TestCSS_ComponentStyles(t *testing.T) {
	css := getCSS()

	// Check for major component styles
	components := []string{
		".sidebar",
		".main-content",
		".nav-menu",
		".btn-primary",
		".btn-secondary",
		".btn-danger",
		".form-control",
		".data-table",
		".pagination",
		".loading",
		".error",
		".success-message",
		".stat-card",
	}

	for _, component := range components {
		if !strings.Contains(css, component) {
			t.Errorf("Expected CSS to contain component style '%s'", component)
		}
	}
}

func TestJS_SearchFunctionality(t *testing.T) {
	js := getJS()

	// Check for search functionality
	searchRules := []string{
		"searchTimeout",
		"setTimeout",
		"clearTimeout",
		"input",
		"search",
		"currentSearch",
	}

	for _, rule := range searchRules {
		if !strings.Contains(js, rule) {
			t.Errorf("Expected JavaScript to contain search functionality '%s'", rule)
		}
	}
}