package model

type SortByRelevanceScoreWebsites []Website

func (websites SortByRelevanceScoreWebsites) Len() int {
	return len(websites)
}

func (websites SortByRelevanceScoreWebsites) Less(i, j int) bool {
	return websites[i].RelevanceScore < websites[j].RelevanceScore
}

func (websites SortByRelevanceScoreWebsites) Swap(i, j int) {
	websites[i], websites[j] = websites[j], websites[i]
}
