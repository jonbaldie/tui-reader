package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

type statefulAction int

const (
	actionResize statefulAction = iota
	actionNextPage
	actionPrevPage
	actionFirstPage
	actionLastPage
	actionNextLink
	actionPrevLink
	actionFollowLink
	actionGoBack
	actionView
)

func TestStatefulActionSequences(t *testing.T) {
	for _, seed := range []int64{1, 7, 19, 37, 101} {
		t.Run(fmt.Sprintf("seed-%d", seed), func(t *testing.T) {
			path := writeTempFile(t, "stateful.md", statefulDocument())
			model := applyWindowSize(NewModel(path), 20, 5)
			rng := rand.New(rand.NewSource(seed))
			var trace []string

			for step := 0; step < 500; step++ {
				action := chooseStatefulAction(model, rng)
				trace = append(trace, action.String())
				model = applyStatefulAction(model, action, rng)
				assertStatefulInvariants(t, model, seed, step, trace)
			}
		})
	}
}

func TestStateful_BackHistoryRemainsValidAfterReflow(t *testing.T) {
	path := writeTempFile(t, "history-reflow.md", statefulDocument())
	model := applyWindowSize(NewModel(path), 20, 5)

	for follows := 0; follows < 2; follows++ {
		model = pressKey(model, "tab")
		model = pressKey(model, "enter")
		if len(model.History()) != follows+1 {
			t.Fatalf("follow %d did not add a history entry: %v", follows+1, model.History())
		}
	}

	model = applyWindowSize(model, 200, 80)
	assertStatefulInvariants(t, model, 0, 0, []string{"two follows", "resize to 200x80"})
}

func chooseStatefulAction(model Model, rng *rand.Rand) statefulAction {
	page := model.BookRef().Pages[model.CurrentPage()]
	actions := []statefulAction{
		actionResize,
		actionNextPage,
		actionPrevPage,
		actionFirstPage,
		actionLastPage,
		actionGoBack,
		actionView,
	}
	if len(page.Links) > 0 {
		actions = append(actions, actionNextLink, actionPrevLink)
	}
	if model.SelectedLink() >= 0 {
		actions = append(actions, actionFollowLink)
	}
	return actions[rng.Intn(len(actions))]
}

func applyStatefulAction(model Model, action statefulAction, rng *rand.Rand) Model {
	switch action {
	case actionResize:
		sizes := [][2]int{{1, 1}, {20, 5}, {40, 12}, {80, 24}, {200, 80}}
		size := sizes[rng.Intn(len(sizes))]
		return applyWindowSize(model, size[0], size[1])
	case actionNextPage:
		return pressKey(model, "right")
	case actionPrevPage:
		return pressKey(model, "left")
	case actionFirstPage:
		return pressKey(model, "home")
	case actionLastPage:
		return pressKey(model, "end")
	case actionNextLink:
		return pressKey(model, "tab")
	case actionPrevLink:
		return pressKey(model, "shift+tab")
	case actionFollowLink:
		return pressKey(model, "enter")
	case actionGoBack:
		return pressKey(model, "b")
	case actionView:
		_ = model.View()
		return model
	default:
		panic(fmt.Sprintf("unknown stateful action %d", action))
	}
}

func assertStatefulInvariants(t *testing.T, model Model, seed int64, step int, trace []string) {
	t.Helper()
	book := model.BookRef()
	if book == nil {
		t.Fatalf("seed %d step %d (%s): book is nil", seed, step, strings.Join(trace, ", "))
	}
	if len(book.Pages) == 0 {
		t.Fatalf("seed %d step %d (%s): book has no pages", seed, step, strings.Join(trace, ", "))
	}
	if model.CurrentPage() < 0 || model.CurrentPage() >= len(book.Pages) {
		t.Fatalf("seed %d step %d (%s): current page %d is outside [0, %d)", seed, step, strings.Join(trace, ", "), model.CurrentPage(), len(book.Pages))
	}

	page := book.Pages[model.CurrentPage()]
	if model.SelectedLink() < -1 || model.SelectedLink() >= len(page.Links) {
		t.Fatalf("seed %d step %d (%s): selected link %d is invalid for page with %d links", seed, step, strings.Join(trace, ", "), model.SelectedLink(), len(page.Links))
	}
	for i, historyPage := range model.History() {
		if historyPage < 0 || historyPage >= len(book.Pages) {
			t.Fatalf("seed %d step %d (%s): history entry %d is invalid after reflow: %d outside [0, %d)", seed, step, strings.Join(trace, ", "), i, historyPage, len(book.Pages))
		}
	}
}

func (a statefulAction) String() string {
	return []string{
		"resize",
		"next-page",
		"prev-page",
		"first-page",
		"last-page",
		"next-link",
		"prev-link",
		"follow-link",
		"go-back",
		"view",
	}[a]
}

func statefulDocument() string {
	var content strings.Builder
	content.WriteString("# Contents\n\n")
	for chapter := 1; chapter <= 8; chapter++ {
		fmt.Fprintf(&content, "[Chapter %d](#chapter-%d)\n", chapter, chapter)
	}
	for chapter := 1; chapter <= 8; chapter++ {
		fmt.Fprintf(&content, "\n# Chapter %d\n\n", chapter)
		if chapter < 8 {
			fmt.Fprintf(&content, "[Next chapter](#chapter-%d)\n\n", chapter+1)
		}
		for line := 0; line < 12; line++ {
			fmt.Fprintf(&content, "Chapter %d filler line %d has enough words to wrap in a narrow terminal.\n", chapter, line)
		}
	}
	return content.String()
}
