package prd

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrNoTasks is returned when the parser cannot derive any tasks from the PRD content.
var ErrNoTasks = errors.New("no tasks detected in PRD")

// Node represents a hierarchical task extracted from a PRD document.
type Node struct {
	Title       string
	Description string
	Children    []*Node
}

// Summary contains lightweight data about a parsed task for UI display.
type Summary struct {
	Title        string
	Description  string
	SubtaskCount int
}

var bulletPattern = regexp.MustCompile(`^(?:\d+[\.)]|[-+*])\s+`)

// Parse converts a PRD document into a hierarchical set of Nodes.
func Parse(content string) ([]*Node, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	// Increase the scanner buffer to tolerate wide markdown tables or paragraphs.
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1*1024*1024)

	var (
		roots        []*Node
		headingStack []*Node
		bulletStack  []*Node
		lastNode     *Node
	)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		raw := strings.TrimRight(scanner.Text(), "\r")
		trimmed := strings.TrimSpace(raw)

		if trimmed == "" {
			appendParagraphBreak(lastNode)
			continue
		}

		if level, title := parseHeading(trimmed); level > 0 {
			node := &Node{Title: title}
			headingStack = trimStack(headingStack, level-1)
			if parent := last(headingStack); parent != nil {
				parent.Children = append(parent.Children, node)
			} else {
				roots = append(roots, node)
			}
			headingStack = append(headingStack, node)
			bulletStack = bulletStack[:0]
			lastNode = node
			continue
		}

		if indent, text, ok := parseBullet(raw); ok {
			node := &Node{Title: text}
			depth := indent / 2
			if depth < 0 {
				depth = 0
			}
			if depth == 0 {
				bulletStack = bulletStack[:0]
			} else if depth < len(bulletStack) {
				bulletStack = bulletStack[:depth]
			}

			var parent *Node
			if depth > 0 && depth <= len(bulletStack) {
				parent = bulletStack[depth-1]
			} else {
				parent = last(headingStack)
			}

			if parent == nil {
				roots = append(roots, node)
			} else {
				parent.Children = append(parent.Children, node)
			}

			if depth < len(bulletStack) {
				bulletStack[depth] = node
				bulletStack = bulletStack[:depth+1]
			} else {
				bulletStack = append(bulletStack, node)
			}
			lastNode = node
			continue
		}

		if lastNode == nil {
			// Treat miscellaneous text as a standalone task when nothing has been created yet.
			node := &Node{Title: trimmed}
			roots = append(roots, node)
			headingStack = append(headingStack, node)
			bulletStack = bulletStack[:0]
			lastNode = node
			continue
		}

		if lastNode.Description != "" {
			lastNode.Description += "\n"
		}
		lastNode.Description += trimmed
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read PRD (line %d): %w", lineNum, err)
	}

	if len(roots) == 0 {
		return nil, ErrNoTasks
	}

	for _, n := range roots {
		trimDescriptions(n)
	}
	return roots, nil
}

// Summaries flattens parsed nodes for presentation purposes.
func Summaries(nodes []*Node) []Summary {
	summaries := make([]Summary, 0, len(nodes))
	for _, n := range nodes {
		summaries = append(summaries, summarizeNode(n))
	}
	return summaries
}

func summarizeNode(n *Node) Summary {
	summary := Summary{
		Title:        n.Title,
		Description:  snippet(n.Description, 140),
		SubtaskCount: len(n.Children),
	}
	return summary
}

func snippet(text string, max int) string {
	text = strings.TrimSpace(text)
	if max <= 0 || len(text) <= max {
		return text
	}
	return text[:max-1] + "…"
}

func parseHeading(line string) (int, string) {
	hashes := 0
	for hashes < len(line) && line[hashes] == '#' {
		hashes++
	}
	if hashes == 0 {
		return 0, ""
	}
	title := strings.TrimSpace(line[hashes:])
	if title == "" {
		return 0, ""
	}
	if hashes > 6 {
		hashes = 6
	}
	return hashes, title
}

func parseBullet(line string) (int, string, bool) {
	indent := countLeadingSpaces(line)
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return 0, "", false
	}
	loc := bulletPattern.FindStringIndex(trimmed)
	if loc == nil {
		return 0, "", false
	}
	text := strings.TrimSpace(trimmed[loc[1]:])
	if text == "" {
		return 0, "", false
	}
	return indent, text, true
}

func trimDescriptions(n *Node) {
	n.Description = strings.TrimSpace(n.Description)
	for _, child := range n.Children {
		trimDescriptions(child)
	}
}

func appendParagraphBreak(node *Node) {
	if node == nil {
		return
	}
	if node.Description != "" && !strings.HasSuffix(node.Description, "\n\n") {
		node.Description += "\n\n"
	}
}

func trimStack(stack []*Node, keep int) []*Node {
	if keep < 0 {
		return stack[:0]
	}
	if keep > len(stack) {
		keep = len(stack)
	}
	return stack[:keep]
}

func last(stack []*Node) *Node {
	if len(stack) == 0 {
		return nil
	}
	return stack[len(stack)-1]
}

func countLeadingSpaces(line string) int {
	count := 0
	for _, r := range line {
		if r == ' ' {
			count++
			continue
		}
		if r == '\t' {
			count += 4
			continue
		}
		break
	}
	return count
}
