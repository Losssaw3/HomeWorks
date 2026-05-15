package articleservice

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sso/internal/domain/models"
	articlehelper "sso/internal/lib/article"
	jsonHelper "sso/internal/lib/json"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"strconv"

	"github.com/gorilla/mux"
	ssov1 "github.com/los3-smurf/photos/gen/go/sso"
)

const userIdKey = "userId"

type ArticleCreator interface {
	Create(ctx context.Context, authorId int64, body string) (int64, error)
}

type ArticleGetter interface {
	GetByAuthor(ctx context.Context, authorId int64) ([]models.Article, error)
	GetById(ctx context.Context, id int64) (models.Article, error)
	GetArticleAuthor(ctx context.Context, id int64) (int64, error)
	GetFeed(ctx context.Context) ([]models.Article, error)
}

type ArticleCache interface {
	GetArticleFromCacheById(ctx context.Context, key string) (models.Article, error)
	GetArticleFromCacheByAuthor(ctx context.Context, key string) ([]models.Article, error)
	DeleteArticleFromCache(ctx context.Context, keys []string) error
	AddArticleToChache(ctx context.Context, key string, article models.Article) error
	AddArticlesToCacheByAuthor(ctx context.Context, key string, articles []models.Article) error
}

type ArticleEditor interface {
	Edit(ctx context.Context, articleId int64, newBody string) error
}

type ArticleAuthI interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password string) (int64, error)
}

type ArticleHandler struct {
	name          string
	appId         int64
	log           *slog.Logger
	createArticle ArticleCreator
	getArticle    ArticleGetter
	editArticle   ArticleEditor
	cacheArticle  ArticleCache
	auth          ArticleAuthI
}

func NewArticleHandler(
	name string,
	appId int64,
	log *slog.Logger,
	createArticle ArticleCreator,
	getArticle ArticleGetter,
	editArticle ArticleEditor,
	cacheArticle ArticleCache,
	auth ArticleAuthI,
) *ArticleHandler {
	return &ArticleHandler{
		name:          name,
		appId:         appId,
		log:           log,
		createArticle: createArticle,
		getArticle:    getArticle,
		editArticle:   editArticle,
		cacheArticle:  cacheArticle,
		auth:          auth,
	}
}

func (h *ArticleHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req ssov1.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "login failed", http.StatusUnauthorized)
		return
	}

	jsonHelper.ResponseJSON(w, http.StatusOK, map[string]string{"token": token})

}

func (h *ArticleHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req *ssov1.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	userId, err := h.auth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "register failed", http.StatusUnauthorized)
		return
	}
	jsonHelper.ResponseJSON(w, http.StatusCreated, map[string]int64{"UserId": userId})
}

func (h *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) {
	uidF, ok := r.Context().Value(userIdKey).(float64)
	if !ok {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	uid := int64(uidF)
	var payload models.ArticlePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	articleId, err := h.createArticle.Create(r.Context(), uid, payload.Body)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonHelper.ResponseJSON(w, http.StatusCreated, map[string]int64{"articleId": articleId})
}

func (h *ArticleHandler) GetRoute(w http.ResponseWriter, r *http.Request) {
	articles, err := h.getArticle.GetFeed(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	articlesStr := []string{}
	for _, article := range articles {
		articlesStr = append(articlesStr, article.Body)
	}
	jsonHelper.ResponseJSON(w, http.StatusOK, map[string][]string{"articles": articlesStr})
}

func (h *ArticleHandler) Edit(w http.ResponseWriter, r *http.Request) {
	uidF, ok := r.Context().Value(userIdKey).(float64)
	if !ok {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	uid := int64(uidF)

	var payload models.EditRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	authorId, err := h.getArticle.GetArticleAuthor(r.Context(), payload.Id)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if authorId != uid {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	err = h.editArticle.Edit(r.Context(), payload.Id, payload.Body)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	var keys []string
	keyArt := "article:" + strconv.Itoa(int(payload.Id))
	keyAuthor := "author:" + strconv.Itoa(int(uid))
	keys = append(keys, keyArt, keyAuthor)
	err = h.cacheArticle.DeleteArticleFromCache(r.Context(), keys)
	if err != nil {
		h.log.Error("error writing cache: %w", sl.Err(err))
	}
	jsonHelper.ResponseJSON(w, http.StatusOK, map[string]int64{"articleId": payload.Id})
}

func (h *ArticleHandler) GetArticleByAuthor(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	authorIdStr := vars["id"]
	authorId, err := strconv.ParseInt(authorIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	cacheKey := "author:" + authorIdStr
	articles, err := h.cacheArticle.GetArticleFromCacheByAuthor(r.Context(), cacheKey)
	if err == nil {
		res := articlehelper.PackArticle(articles)
		jsonHelper.ResponseJSON(w, http.StatusOK, map[string][]string{"articles": res})
		return
	}
	if !errors.Is(err, storage.ErrCacheNotFound) {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	articles, err = h.getArticle.GetByAuthor(r.Context(), authorId)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	res := articlehelper.PackArticle(articles)
	err = h.cacheArticle.AddArticlesToCacheByAuthor(r.Context(), cacheKey, articles)
	if err != nil {
		h.log.Error("error writing cache: %w", sl.Err(err))
	}
	jsonHelper.ResponseJSON(w, http.StatusOK, map[string][]string{"articles": res})
}

func (h *ArticleHandler) GetArticleById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleIdStr := vars["id"]
	articleId, err := strconv.ParseInt(articleIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	cacheKey := "article:" + articleIdStr
	article, err := h.cacheArticle.GetArticleFromCacheById(r.Context(), cacheKey)
	if err == nil {
		jsonHelper.ResponseJSON(w, http.StatusOK, map[string]string{"article": article.Body})
		return
	}

	if !errors.Is(err, storage.ErrCacheNotFound) {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	article, err = h.getArticle.GetById(r.Context(), articleId)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	err = h.cacheArticle.AddArticleToChache(r.Context(), cacheKey, article)
	if err != nil {
		h.log.Error("error writing cache: %w", sl.Err(err))
	}
	jsonHelper.ResponseJSON(w, http.StatusOK, map[string]string{"article": article.Body})
}
