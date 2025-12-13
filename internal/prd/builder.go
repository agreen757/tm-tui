package prd

// BuildTaskDocuments converts parsed nodes into JSON-friendly maps that align with
// the Task Master tasks.json format. The first root task will receive the
// provided startID value and subsequent tasks increment sequentially.
func BuildTaskDocuments(nodes []*Node, startID int) []map[string]interface{} {
	docs := make([]map[string]interface{}, 0, len(nodes))
	id := startID
	for _, node := range nodes {
		docs = append(docs, buildDocument(node, id))
		id++
	}
	return docs
}

func buildDocument(node *Node, id int) map[string]interface{} {
	doc := map[string]interface{}{
		"id":       id,
		"title":    node.Title,
		"status":   "pending",
		"priority": "medium",
	}

	if node.Description != "" {
		doc["description"] = node.Description
	}

	if len(node.Children) > 0 {
		subtasks := make([]map[string]interface{}, 0, len(node.Children))
		for i, child := range node.Children {
			childID := i + 1
			subtasks = append(subtasks, buildDocument(child, childID))
		}
		doc["subtasks"] = subtasks
	}

	return doc
}
