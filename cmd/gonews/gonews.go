// Сервер GoNews.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"GoNews/pkg/api"
	"GoNews/pkg/rss"
	"GoNews/pkg/storage"
)

// конфигурация приложения
type config struct {
	URLS   []string         `json:"rss"`
	Period int              `json:"request_period"`
	DB     storage.DBConfig `json:"db"`
}

func main() {
	// чтение и раскодирование файла конфигурации
	b, err := os.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	var config config
	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Fatal(err)
	}
	// инициализация зависимостей приложения
	dbInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", config.DB.Host, config.DB.User, config.DB.Password, config.DB.DBName, config.DB.Port, config.DB.SSLMode)
	db, err := storage.New(dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	api := api.New(db)

	// запуск парсинга новостей в отдельном потоке
	// для каждой ссылки
	chPosts := make(chan []storage.Post)
	chErrs := make(chan error)
	for _, url := range config.URLS {
		go parseURL(url, db, chPosts, chErrs, config.Period)
	}

	// запись потока новостей в БД
	go func() {
		for posts := range chPosts {
			db.StoreNews(posts)
		}
	}()

	// обработка потока ошибок
	go func() {
		for err := range chErrs {
			log.Println("ошибка:", err)
		}
	}()

	// запуск веб-сервера с API и приложением
	err = http.ListenAndServe(":80", api.Router())
	if err != nil {
		log.Fatal(err)
	}
}

// Асинхронное чтение потока RSS. Раскодированные
// новости и ошибки пишутся в каналы.
func parseURL(url string, db *storage.DB, posts chan<- []storage.Post, errs chan<- error, period int) {
	for {
		news, err := rss.Parse(url)
		if err != nil {
			errs <- err
			continue
		}
		posts <- news
		time.Sleep(time.Minute * time.Duration(period))
	}
}
