// API приложения GoNews.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"GoNews/pkg/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	db storage.DBInterface
	r  *chi.Mux
}

// Конструктор API.
func New(db storage.DBInterface) *API {
	a := API{db: db, r: chi.NewRouter()}
	a.endpoints()
	return &a
}

// Router возвращает маршрутизатор для использования
// в качестве аргумента HTTP-сервера.
func (api *API) Router() *chi.Mux {
	return api.r
}

// Регистрация методов API в маршрутизаторе запросов.
func (api *API) endpoints() {
	// Middleware для логирования запросов
	api.r.Use(middleware.Logger)

	// получить n последних новостей
	api.r.Route("/news", func(r chi.Router) {
		r.Get("/{n}", api.posts)
	})

	// веб-приложение
	api.r.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))

}

func (api *API) posts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// Получаем параметр n из URL
	s := chi.URLParam(r, "n")
	n, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	// Получаем новости из базы данных
	news, err := api.db.News(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем новости в формате JSON
	json.NewEncoder(w).Encode(news)
}
