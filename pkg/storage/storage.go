// Пакет для работы с БД приложения GoNews.
package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Интерфейс для работы с базой данных
type DBInterface interface {
	News(n int) ([]Post, error)
	StoreNews(news []Post) error
}

type DBConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	Port     int    `json:"port"`
	SSLMode  string `json:"sslmode"`
}

// База данных.
type DB struct {
	pool *pgxpool.Pool
}

type Post struct {
	ID      int    // номер записи
	Title   string // заголовок публикации
	Content string // содержание публикации
	PubTime int64  // время публикации
	Link    string // ссылка на источник
}

func New(connstr string) (*DB, error) {
	if connstr == "" {
		return nil, errors.New("не указано подключение к БД")
	}
	pool, err := pgxpool.Connect(context.Background(), connstr)
	if err != nil {
		return nil, err
	}
	db := DB{
		pool: pool,
	}
	return &db, nil
}

func (db *DB) StoreNews(news []Post) error {
	for _, post := range news {
		_, err := db.pool.Exec(context.Background(), `
		INSERT INTO news(title, content, pub_time, link)
		VALUES ($1, $2, $3, $4)`,
			post.Title,
			post.Content,
			post.PubTime,
			post.Link,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// News возвращает последние новости из БД.
func (db *DB) News(n int) ([]Post, error) {
	if n == 0 {
		n = 10
	}
	rows, err := db.pool.Query(context.Background(), `
	SELECT id, title, content, pub_time, link FROM news
	ORDER BY pub_time DESC
	LIMIT $1`,
		n,
	)
	if err != nil {
		return nil, err
	}
	var news []Post
	for rows.Next() {
		var p Post
		err = rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.PubTime,
			&p.Link,
		)
		if err != nil {
			return nil, err
		}
		news = append(news, p)
	}
	return news, rows.Err()
}
