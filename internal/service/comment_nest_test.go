package service

import (
	"testing"
	"test/internal/model"
	"time"

	"github.com/google/uuid"
)

func TestNestCommentsAttachesRepliesToParents(t *testing.T) {
	t.Parallel()

	rootID := uuid.New()
	replyID := uuid.New()
	articleID := uuid.New()
	now := time.Now().UTC()

	comments := []model.Comment{
		{
			ID:         rootID,
			ArticleID:  articleID,
			AuthorName: "Ada",
			Body:       "Root",
			Status:     model.CommentStatusApproved,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         replyID,
			ArticleID:  articleID,
			ParentID:   &rootID,
			AuthorName: "Grace",
			Body:       "Reply",
			Status:     model.CommentStatusApproved,
			CreatedAt:  now.Add(time.Minute),
			UpdatedAt:  now.Add(time.Minute),
		},
	}

	nested := nestComments(comments)
	if len(nested) != 1 {
		t.Fatalf("expected 1 root, got %d", len(nested))
	}
	if nested[0].ID != rootID {
		t.Fatalf("unexpected root id %s", nested[0].ID)
	}
	if len(nested[0].Replies) != 1 {
		t.Fatalf("expected 1 reply nested under root, got %d", len(nested[0].Replies))
	}
	if nested[0].Replies[0].ID != replyID {
		t.Fatalf("unexpected reply id %s", nested[0].Replies[0].ID)
	}
	if nested[0].Replies[0].Body != "Reply" {
		t.Fatalf("unexpected reply body %q", nested[0].Replies[0].Body)
	}
}
