package elsearch

// Construir query de busca
func (es *Client) buildSearchQuery(query string, from, size int) map[string]interface{} {
	return map[string]interface{}{
		"from": from,
		"size": size,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query": query,
						"fields": []string{
							"title^3",                   // Título com boost 3x
							"description^2",             // Descrição com boost 2x
							"search_text^2",             // Search text com boost 2x
							"assigned_agent.full_name",  // Nome do agente
							"company.name",              // Nome da empresa
							"created_by_user.full_name", // Nome do usuário
							"category.name",             // Nome da categoria
							"subcategory.name",          // Nome da subcategoria
							"product.name",              // Nome do produto
							"product.description",       // Descrição do produto
							"tags",                      // Tags
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
