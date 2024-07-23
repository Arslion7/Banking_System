package main

import (
	"fmt"
	"myapp/controllers"
	"myapp/migrations"
	"myapp/models"
	"myapp/templates"
	"myapp/views"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfid() (config, error){
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}
	cfg.PSQL = models.DefaultPostgresConfig()
	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")
	cfg.CSRF.Key = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	cfg.CSRF.Secure = false
	cfg.Server.Address = ":3000" 
	return cfg, nil
}

func main() {
	cfg, err := loadEnvConfid()
	if err != nil {
		panic(err)
	}
	// Setup the database
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Setup the middleware
	CSRF := csrf.Protect([]byte(cfg.CSRF.Key),csrf.Secure(cfg.CSRF.Secure))

	umw := controllers.UserMiddleware{
		SessionService: &models.SessionService{
			DB: db,
		},
	}

	// Setup the controllers
	usersC := controllers.Users{
		UserService: &models.UserService{
			DB: db,
		},
		SessionService: &models.SessionService{
			DB: db,
		},
		EmailService: models.NewMessageService(cfg.SMTP),
		PasswordResetService: &models.PasswordResetService{
			DB: db,
		},
	}
	usersC.Templates.New = views.Must(views.Parse(templates.FS, "signup.gohtml", "tailwindcss.gohtml"))
	usersC.Templates.SignIn = views.Must(views.Parse(templates.FS, "signin.gohtml", "tailwindcss.gohtml"))
	usersC.Templates.PasswordReset = views.Must(views.Parse(templates.FS, "forgot-pw.gohtml", "tailwindcss.gohtml"))
	usersC.Templates.CheckYourEmail = views.Must(views.Parse(templates.FS, "check-your-email.gohtml", "tailwindcss.gohtml"))
	usersC.Templates.ResetPassword = views.Must(views.Parse(templates.FS, "reset-pw.gohtml", "tailwindcss.gohtml"))

	// Setup the router and routes
	r := chi.NewRouter()
	r.Use(CSRF)
	r.Use(umw.SetUser)
	r.Get("/", controllers.StaticHandler(views.Must(views.Parse(templates.FS, "home.gohtml", "tailwindcss.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.Parse(templates.FS, "contact.gohtml", "tailwindcss.gohtml"))))
	r.Get("/faq", controllers.FAQ(views.Must(views.Parse(templates.FS, "faq.gohtml", "tailwindcss.gohtml"))))
	r.Get("/signup", usersC.New)
	r.Post("/users", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.Authenticate)
	r.Post("/signout", usersC.SignOut)
	r.Get("/forgot-pw", usersC.PasswordReset)
	r.Post("/forgot-pw", usersC.ProcessPasswordReset)
	r.Get("/reset-pw", usersC.ReserPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)
	r.Route("/users/me", func (r chi.Router){
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request){
		http.Error(w, "Page not found", http.StatusNotFound)
	})
	
	fmt.Printf("Starting the server on %s ...", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
}