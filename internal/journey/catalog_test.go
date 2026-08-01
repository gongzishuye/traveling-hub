package journey

import "testing"

func TestCatalogContainsEveryFrontendJourneyTemplate(t *testing.T) {
	const expectedTemplates = 18
	if got := len(Templates()); got != expectedTemplates {
		t.Fatalf("len(Templates()) = %d, want %d frontend templates", got, expectedTemplates)
	}
	for _, template := range Templates() {
		if template.ID == "" || template.FoodID == "" || template.PostcardID == "" {
			t.Fatalf("invalid template: %#v", template)
		}
		for _, event := range template.Events {
			if event == "" {
				t.Fatalf("template %q has an empty event", template.ID)
			}
		}
	}
}
