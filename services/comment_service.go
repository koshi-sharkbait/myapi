package services

import (

	"github.com/yourname/reponame/apperrors"
	"github.com/yourname/reponame/models"
	"github.com/yourname/reponame/repositories"
)

// PostCommentHandlerで使用することを想定したサービス
func (s *MyAppService) PostCommentService(comment models.Comment) (models.Comment, error) {
	newComment, err := repositories.InsertComment(s.db, comment)
	if err != nil {
		err = apperrors.InsertDataFailed.Wrap(err, "fail to record data")
		return models.Comment{}, err
	}

	return newComment, nil

}
