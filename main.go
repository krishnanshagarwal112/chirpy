package main

import (
	"log"
	"net/http"
	"os"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
	AuthorId int `json:"author_id"`
}

type User struct {
	Id   int    `json:"id"`
	Email string `json:"email"`
	Password string `json:"password"`
	Is_Chirpy_Red bool `json:"is_chirpy_red"`
}

func main() {
	godotenv.Load()
	JWTSecret := os.Getenv("JWT_SECRET")
	POLKAkey := os.Getenv("POLKA_KEY")
	const port = "8080"
	DB, err := NewDB("")
	if err != nil {
		return
	}
	r := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()
	corsMux := middlewareCors(r) // it wraps mux with the middlewareCors, this ensure that all request first passes through CORS MIDDLEWARE

	var srv http.Server   // we create a variable srv which is of type http.Server !
	srv.Addr = ":" + port // corrected server address
	srv.Handler = corsMux // the handler will be corsMux , it is to ensure that every request must first go through this middleware

	var apiCfg apiConfig
	apiCfg.fileserverHits = 0
	apiCfg.jwtSecret = []byte(JWTSecret)
	apiCfg.polkaKey = POLKAkey

	fileHandler := http.FileServer(http.Dir("."))

	r.Mount("/", apiCfg.middlewareMetricsInc(fileHandler))
	apiRouter.Get("/metrics", apiCfg.metricsHandler)
	apiRouter.Get("/healthz", handlerReadiness)

	apiRouter.Get("/chirps", func(w http.ResponseWriter, r *http.Request) {
		chirpsGet(w, r, DB)
	})

	apiRouter.Post("/chirps", func(w http.ResponseWriter, r *http.Request) {
		chirpsPost(w, r, DB, &apiCfg)
	})

	apiRouter.Get("/chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
		chirpsGetById(w, r, DB)
	})

	apiRouter.Post("/users",func(w http.ResponseWriter, r *http.Request) {
		userPost(w,r,DB)
	})

	apiRouter.Put("/users",func(w http.ResponseWriter, r *http.Request) {
		usersPut(w,r,DB,&apiCfg)
	})

	apiRouter.Post("/login",func(w http.ResponseWriter, r *http.Request) {
		userLogin(w,r,DB,&apiCfg)
	})

	apiRouter.Post("/refresh",func(w http.ResponseWriter, r *http.Request) {
		refresh(w,r,DB, &apiCfg)
	})

	apiRouter.Post("/revoke",func(w http.ResponseWriter, r *http.Request) {
		revoke(w,r,DB, &apiCfg)
	})

	apiRouter.Delete("/chirps/{chirpID}",func(w http.ResponseWriter, r *http.Request) {
		delete(w,r,DB, &apiCfg)
	})

	apiRouter.Post("/polka/webhooks",func(w http.ResponseWriter, r *http.Request,) {
		webhook(w,r, DB, &apiCfg)
	})


	adminRouter.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w,r,&apiCfg)
	})



	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
