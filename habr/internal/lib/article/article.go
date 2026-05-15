package articlehelper

import (
	"sso/internal/domain/models"
	"strconv"
)

func PackArticle(in []models.Article) []string {
	articlesStr := []string{}
	for _, article := range in {
		articlesStr = append(articlesStr, article.Body)
	}
	return articlesStr
}

func ToCache(in []models.Article) map[string]string {
	toCache := make(map[string]string)
	for _, article := range in {
		toCache["author"+strconv.Itoa(int(article.AuthorId))] = article.Body
	}
	return toCache
}
