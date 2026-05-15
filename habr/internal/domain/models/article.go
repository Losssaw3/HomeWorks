package models

type Article struct {
	Id       int64  `json:"id"`
	AuthorId int64  `json:"author_id"`
	Body     string `json:"body"`
}

type ArticlePayload struct {
	Body string `json:"body"`
}

type CreateRequest struct {
	ArticlePayload
}

type EditRequest struct {
	Id int64 `json:"id"`
	ArticlePayload
}
