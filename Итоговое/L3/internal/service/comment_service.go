package service

import (
	"context"
	"time"

	"notification-service/internal/model"

	"github.com/google/uuid"
)

type CommentStorage interface {
	CreateComment(ctx context.Context, c *model.Comment) error
	DeleteComment(ctx context.Context, id string) error
	GetCommentTree(ctx context.Context, rootID string) ([]*model.Comment, error)
	GetRootComments(ctx context.Context, limit, offset int, sortDesc bool) ([]*model.Comment, error)
	SearchComments(ctx context.Context, keyword string) ([]*model.Comment, error)
}

type CommentService struct {
	storage CommentStorage
}

func NewCommentService(s CommentStorage) *CommentService {
	return &CommentService{storage: s}
}

func (s *CommentService) Create(ctx context.Context, parentID *string, author, text string) (*model.Comment, error) {
	if author == "" {
		author = "Аноним"
	}

	c := &model.Comment{
		ID:        uuid.New().String(),
		ParentID:  parentID,
		Author:    author,
		Text:      text,
		CreatedAt: time.Now(),
	}

	err := s.storage.CreateComment(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// GetTree — магия сборки дерева. Берет плоский список из БД и связывает через ParentID
func (s *CommentService) GetTree(ctx context.Context, rootID string) (*model.Comment, error) {
	flatComments, err := s.storage.GetCommentTree(ctx, rootID)
	if err != nil || len(flatComments) == 0 {
		return nil, err
	}

	commentMap := make(map[string]*model.Comment)
	for _, c := range flatComments {
		commentMap[c.ID] = c
	}

	var root *model.Comment

	for _, c := range flatComments {
		if c.ID == rootID {
			root = c
		} else if c.ParentID != nil {
			if parent, exists := commentMap[*c.ParentID]; exists {
				parent.Children = append(parent.Children, c)
			}
		}
	}

	return root, nil
}

// GetRoots — получает список корневых комментариев (для главной страницы обсуждения)
func (s *CommentService) GetRoots(ctx context.Context, limit, offset int) ([]*model.Comment, error) {
	return s.storage.GetRootComments(ctx, limit, offset, true)
}

// Search — полнотекстовый поиск
func (s *CommentService) Search(ctx context.Context, keyword string) ([]*model.Comment, error) {
	return s.storage.SearchComments(ctx, keyword)
}

// Delete — удаление (и каскадное удаление ответов благодаря БД)
func (s *CommentService) Delete(ctx context.Context, id string) error {
	return s.storage.DeleteComment(ctx, id)
}
