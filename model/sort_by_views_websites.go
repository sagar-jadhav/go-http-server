package model

type SortByViewsWebsites []Website

func (websites SortByViewsWebsites) Len() int {
	return len(websites)
}

func (websites SortByViewsWebsites) Less(i, j int) bool {
	return websites[i].Views < websites[j].Views
}

func (websites SortByViewsWebsites) Swap(i, j int) {
	websites[i], websites[j] = websites[j], websites[i]
}
