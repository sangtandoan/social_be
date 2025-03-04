package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/sangtandoan/social/internal/models/dto"
	"github.com/sangtandoan/social/internal/utils"
)

type PostsStore struct {
	db *sql.DB
}

type Post struct {
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tags      []string  `json:"tags"`
	UserID    int       `json:"user_id"`
	ID        int       `json:"id"`
}

func (s *PostsStore) Create(ctx context.Context, post *Post) error {
	query := "INSERT INTO posts (title, content, user_id, tags) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at"

	// SQL query timeout
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	row := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.UserID,
		pq.Array(post.Tags),
	)

	// Scan need address of fields in that struct not address of that struct
	err := row.Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostsStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := "SELECT id, title, content, tags, created_at, updated_at FROM posts WHERE id = $1"

	var post Post
	row := s.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}

	return &post, nil
}

func (s *PostsStore) GetAll(ctx context.Context) ([]*Post, error) {
	query := "SELECT id, title, content, tags, created_at, updated_at FROM posts"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	var res []*Post
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var post Post
		var tags pq.StringArray

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&tags,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		post.Tags = []string(tags)

		res = append(res, &post)
	}

	return res, nil
}

type UpdatePostParams struct {
	Title   *string   `json:"title,omitempty"`
	Content *string   `json:"content,omitempty"`
	Tags    *[]string `json:"tags,omitempty"`
	ID      int64     `json:"id,omitempty"`
}

// Parial update should have dynamic query string to optimize query
// Instead of just update everything, just update needed things will reduce I/O operations
func (s *PostsStore) UpdatePost(ctx context.Context, arg *UpdatePostParams) (*Post, error) {
	query := "UPDATE posts SET "
	var params []any
	if arg.Title != nil {
		query += fmt.Sprintf("title = $%d, ", len(params)+1)
		params = append(params, arg.Title)
	}
	if arg.Content != nil {
		fmt.Println("I'm at content")
		query += fmt.Sprintf("content = $%d, ", len(params)+1)
		params = append(params, arg.Content)
	}
	if arg.Tags != nil {
		query += fmt.Sprintf("tags = $%d, ", len(params)+1)
		params = append(params, pq.Array(arg.Tags))
	}

	query = strings.Trim(query, " ")
	query = strings.Trim(query, ",")
	query += fmt.Sprintf(
		" WHERE id = $%d RETURNING id, title, content, tags, created_at, updated_at",
		len(params)+1,
	)
	params = append(params, arg.ID)

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	row := s.db.QueryRowContext(ctx, query, params...)
	var post Post

	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostsStore) DeleteByID(ctx context.Context, id int64) error {
	query := "DELETE FROM posts WHERE id = $1"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

type PostResponse struct {
	Username string
	Post
	CommentsCount int64
}

func (s *PostsStore) GetUserFeed(
	ctx context.Context,
	arg *dto.UserFeedRequest,
) ([]*PostResponse, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.title, p.content, p.created_at, p.tags,
			u.username,
			COUNT(c.id) as comments_count
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON u.id = p.user_id
		JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
		WHERE f.user_id = $1 AND
			(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
			(p.tags @> $5 OR $5 = '{}')
		GROUP BY p.id, u.username
		ORDER BY p.created_at DESC
		OFFSET $2
		LIMIT $3
	`
	rows, err := s.db.QueryContext(
		ctx,
		query,
		arg.ID,
		arg.Offset,
		arg.Limit,
		arg.Search,
		pq.Array(arg.Tags),
	)
	if err != nil {
		return nil, err
	}

	var arr []*PostResponse
	for rows.Next() {
		var response PostResponse

		err := rows.Scan(
			&response.ID,
			&response.UserID,
			&response.Title,
			&response.Content,
			&response.CreatedAt,
			pq.Array(&response.Tags),
			&response.Username,
			&response.CommentsCount,
		)
		if err != nil {
			return nil, err
		}

		arr = append(arr, &response)
	}

	return arr, nil
}
