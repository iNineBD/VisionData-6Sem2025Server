package elsearch

// Construir query de busca
func (es *Client) buildSearchQuery(query string, from, size int) map[string]interface{} {
	if query == "" {
		// Sem query: apenas paginação e ordenação
		return map[string]interface{}{
			"from": from,
			"size": size,
			"sort": []map[string]interface{}{
				{
					"dates.created_at": map[string]string{
						"order": "desc",
					},
				},
			},
		}
	}
	// Com query: busca normal
	return map[string]interface{}{
		"from": from,
		"size": size,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query": query,
						"fields": []string{
							"title^3",
							"description^2",
							"search_text^2",
							"assigned_agent.full_name",
							"company.name",
							"created_by_user.full_name",
							"category.name",
							"subcategory.name",
							"product.name",
							"product.description",
							"tags",
						},
						"type":                 "best_fields",
						"fuzziness":            "AUTO",
						"operator":             "or",
						"minimum_should_match": "2",
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{
				"_score": map[string]string{
					"order": "desc",
				},
			},
			{
				"dates.created_at": map[string]string{
					"order": "desc",
				},
			},
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"title":       map[string]interface{}{},
				"description": map[string]interface{}{},
			},
			"pre_tags":  []string{"<mark>"},
			"post_tags": []string{"</mark>"},
		},
	}
}
