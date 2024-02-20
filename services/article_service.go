package services

import (
	"database/sql"
	"errors"

	"github.com/yourname/reponame/apperrors"
	"github.com/yourname/reponame/models"
	"github.com/yourname/reponame/repositories"
)

// PostArticleHandlerで使うことを想定したサービス
func (s *MyAppService) PostArticleService(article models.Article) (models.Article, error) {
	newArticle, err := repositories.InsertArticle(s.db, article)
	if err != nil {
		return models.Article{}, err
	}

	return newArticle, err
}

// ArticleListHandlerで使うことを想定したサービス
func (s *MyAppService) GetArticleListService(page int) ([]models.Article, error) {
	articleList, err := repositories.SelectArticleList(s.db, page)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = apperrors.GetDataFailed.Wrap(err, "fail to get data")
		}
		if len(articleList) == 0 {
			err = apperrors.NAData.Wrap(ErrNoData, "no data")
			return nil, err
		}
		return []models.Article{}, err
	}

	return articleList, nil
}

// ArticleDetailHandlerで使うことを想定したサービス
func (s *MyAppService) GetArticleService(articleID int) (models.Article, error) {
	type articleResult struct {
		article models.Article
		err     error
	}
	articleChan := make(chan articleResult)
	defer close(articleChan)
	go func(ch chan<- articleResult) {
		newarticle, err := repositories.SelectArticleDetail(s.db, articleID)
		ch <- articleResult{article: newarticle, err: err}
	}(articleChan)

	type commentResult struct {
		commentList *[]models.Comment
		err         error
	}
	commentChan := make(chan commentResult)
	go func(ch chan commentResult) {
		newCommentList, err := repositories.SelectCommentList(s.db, articleID)
		ch <- commentResult{commentList: &newCommentList, err: err}
	}(commentChan)

	var article models.Article
	var commentList []models.Comment
	var articleGetErr, commentGetErr error

	for i := 0; i < 2; i++ {
		select {
		case ar := <-articleChan:
			article, articleGetErr = ar.article, ar.err
		case cr := <-commentChan:
			commentList, commentGetErr = *cr.commentList, cr.err
		}
	}

	if articleGetErr != nil {
		if errors.Is(articleGetErr, sql.ErrNoRows) {
			err := apperrors.NAData.Wrap(articleGetErr, "no data")
			return models.Article{}, err
		}
		err := apperrors.GetDataFailed.Wrap(articleGetErr, "fail to data")
		return models.Article{}, err
	}

	if commentGetErr != nil {
		err := apperrors.GetDataFailed.Wrap(commentGetErr, "fail to data")
		return models.Article{}, err
	}

	article.CommentList = append(article.CommentList, commentList...)

	return article, nil
}

// PostNiceHandlerで使うことを想定したサービス
func (s *MyAppService) PostNiceService(article models.Article) (models.Article, error) {
	err := repositories.UpdateNiceNum(s.db, article.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = apperrors.NoTargetData.Wrap(err, "doesn't exist target article")
			return models.Article{}, err
		}
		err = apperrors.UpdateDataFailed.Wrap(err, "fail to updata nice count")
		return models.Article{}, err
	}

	return models.Article{
		ID:        article.ID,
		Title:     article.Title,
		Contents:  article.Contents,
		UserName:  article.UserName,
		NiceNum:   article.NiceNum + 1,
		CreatedAt: article.CreatedAt,
	}, nil
}
