package prd

import "testing"

func TestParseHeadingsAndBullets(t *testing.T) {
	content := `# Feature A

Intro text describing the feature.

## Flow
- Step 1
  - Detail 1
  - Detail 2
- Step 2

# Feature B
- Item
`

	nodes, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 root nodes, got %d", len(nodes))
	}

	if got := nodes[0].Title; got != "Feature A" {
		t.Fatalf("unexpected first title: %s", got)
	}
	if nodes[0].Description == "" {
		t.Fatalf("expected description for Feature A")
	}

	sub := nodes[0].Children
	if len(sub) != 1 {
		t.Fatalf("expected Flow child, got %d", len(sub))
	}
	if len(sub[0].Children) != 2 {
		t.Fatalf("expected 2 bullet children, got %d", len(sub[0].Children))
	}
}

func TestParseNoTasks(t *testing.T) {
	if _, err := Parse("   \n\n"); err != ErrNoTasks {
		t.Fatalf("expected ErrNoTasks, got %v", err)
	}
}

func TestBuildTaskDocuments(t *testing.T) {
	nodes := []*Node{{
		Title:       "Task A",
		Description: "Desc",
		Children:    []*Node{{Title: "Child"}},
	}}

	docs := BuildTaskDocuments(nodes, 5)
	if len(docs) != 1 {
		t.Fatalf("expected single document, got %d", len(docs))
	}
	if docs[0]["id"].(int) != 5 {
		t.Fatalf("expected id 5, got %v", docs[0]["id"])
	}
	sub := docs[0]["subtasks"].([]map[string]interface{})
	if sub[0]["id"].(int) != 1 {
		t.Fatalf("expected child id 1, got %v", sub[0]["id"])
	}
}

func TestSummaries(t *testing.T) {
	nodes := []*Node{{Title: "Task", Description: "A long description"}}
	s := Summaries(nodes)
	if len(s) != 1 {
		t.Fatalf("expected summary entry")
	}
	if s[0].Title != "Task" {
		t.Fatalf("unexpected title %s", s[0].Title)
	}
}
