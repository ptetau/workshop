package main

import (
	"fmt"
	"strings"
)

// UpdateHTMLForm injects missing inputs into a form.
// It looks for </form> and inserts fields before it.
func UpdateHTMLForm(src string, params []field) (string, error) {
	insertionPoint := strings.LastIndex(src, "</button>")
	if insertionPoint == -1 {
		// Fallback to </form>
		insertionPoint = strings.LastIndex(src, "</form>")
	}
	if insertionPoint == -1 {
		return src, fmt.Errorf("could not find closing </form> or submit button tag")
	}

	var toInsert string
	for _, p := range params {
		// Check if exists
		// Heuristic: check for name="Field"
		if strings.Contains(src, fmt.Sprintf("name=\"%s\"", p.Name)) {
			continue
		}

		// Generate Input
		// Reusing logic from formHTMLTemplate roughly
		inputType := "text"
		if p.Type == "int" {
			inputType = "number"
		} else if p.Type == "bool" {
			inputType = "checkbox"
		}

		fieldHTML := fmt.Sprintf(`
        <div class="form-group">
            <label for="id_%s">%s</label>
            <input type="%s" id="id_%s" name="%s" required>
        </div>`, p.Name, p.Name, inputType, p.Name, p.Name)

		toInsert += fieldHTML
	}

	if toInsert == "" {
		return src, nil
	}

	// Insert before the button (assuming button is at end of form)
	// Or simply before </form> if button logic is fuzzy.
	// But usually form ends with <button>Submit</button></form>
	// Let's insert BEFORE the button to be nice.

	return src[:insertionPoint] + toInsert + "\n        " + src[insertionPoint:], nil
}

// UpdateHTMLView injects missing fields into a dl.
func UpdateHTMLView(src string, fields []field) (string, error) {
	insertionPoint := strings.LastIndex(src, "</dl>")
	if insertionPoint == -1 {
		return src, fmt.Errorf("could not find closing </dl> tag")
	}

	var toInsert string
	for _, f := range fields {
		// Check presence
		if strings.Contains(src, fmt.Sprintf("{{ .%s }}", f.Name)) {
			continue
		}

		// Generate DD/DT
		// We need to use {{ printf ... }} style if we are writing to a Go string that will be compiled?
		// No, we are writing to the HTML file on disk. The HTML file on disk contains `{{ .Field }}`.
		// Wait, the `formHTMLTemplate` returns a string literal that contains Go Template actions.
		// When we read the file back from disk, we read `{{ .Field }}`.
		// So we can insert `{{ .NewField }}` directly.

		fieldHTML := fmt.Sprintf(`
        <dt>%s</dt>
        <dd>{{ .%s }}</dd>`, f.Name, f.Name)

		toInsert += fieldHTML
	}

	if toInsert == "" {
		return src, nil
	}

	return src[:insertionPoint] + toInsert + "\n    " + src[insertionPoint:], nil
}
