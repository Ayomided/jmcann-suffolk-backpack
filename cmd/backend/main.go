package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	app "adediiji.uk/jmcann-suffolk-backpack-task/internal"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/db"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/ui"
)

func main() {
	addr := flag.String("addr", ":3000", "HTTP Network Address")
	db_path := flag.String("db_path", "test.db", "DB Connection String")
	seed := flag.Bool("seed", false, "Seed the database with test data")
	migrate := flag.Bool("migrate", false, "Run inital database migration")
	flag.Parse()

	infoLog := log.New(os.Stdout, "[INFO]\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "[ERROR]\t", log.Ldate|log.Ltime|log.Lshortfile)

	dbConn, err := db.OpenDBConnection(*db_path)
	if err != nil {
		panic(err.Error())
	}
	defer dbConn.Close()
	infoLog.Println("DB startup successful")

	if *migrate {
		err = db.SetupDB(dbConn)
		if err != nil {
			panic(err.Error())
		}
		infoLog.Println("DB migration successful")
		return
	}

	storage := db.AppStorage{
		DB: dbConn,
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		errorLog.Fatal("JWT_SECRET not set")
	}

	if *seed {
		log.Println("Seeding database...")
		if err := db.Seed(&storage, jwtSecret); err != nil {
			errorLog.Fatalf("Seed failed: %v", err)
		}
		infoLog.Println("DB seed complete")
		return
	}

	templateCache, err := ui.NewTemplateCache()
	if err != nil {
		errorLog.Fatalf("Failed to build template cache: %v", err)
	}

	app := app.JMcCannBackPackApp{
		ErrorLog: errorLog,
		InfoLog:  infoLog,
		DB:       &storage,
		Config: app.AppConfig{
			JWTSecret:    jwtSecret,
			SecureCookie: false,
		},
		TemplateCache: templateCache,
	}

	mux := http.NewServeMux()

	fileserver := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileserver))
	mux.HandleFunc("POST /refresh", app.RefreshToken)
	mux.HandleFunc("GET /login", app.LoginForm)
	mux.HandleFunc("POST /login", app.LoginFormPost)

	protected := http.NewServeMux()

	protected.HandleFunc("GET /dashboard", app.Dashboard)
	protected.HandleFunc("GET /logout", app.Logout)

	protected.HandleFunc("GET /jobs", app.JobsList)
	protected.HandleFunc("GET /jobs/new", app.JobCreate)
	protected.HandleFunc("POST /jobs/new", app.JobCreatePost)
	protected.HandleFunc("GET /jobs/{id}", app.JobView)
	protected.HandleFunc("POST /jobs/{id}/status", app.JobUpdateStatus)
	protected.HandleFunc("POST /jobs/{id}/delete", app.JobDelete)
	protected.HandleFunc("POST /jobs/{id}/calculate", app.JobCalculateCost)

	protected.HandleFunc("GET /operatives", app.OperativesList)
	protected.HandleFunc("GET /operatives/new", app.OperativeCreate)
	protected.HandleFunc("POST /operatives/new", app.OperativeCreatePost)
	protected.HandleFunc("GET /operatives/{id}", app.OperativeView)
	protected.HandleFunc("POST /operatives/{id}", app.OperativeUpdate)
	protected.HandleFunc("POST /operatives/{id}/delete", app.OperativeDelete)
	protected.HandleFunc("GET /operatives/{id}/rates", app.OperativeRateHistory)
	protected.HandleFunc("POST /operatives/{id}/rates", app.OperativeRateCreate)

	protected.HandleFunc("GET /resources", app.ResourcesList)
	protected.HandleFunc("GET /resources/new", app.ResourceCreate)
	protected.HandleFunc("POST /resources/new", app.ResourceCreatePost)
	protected.HandleFunc("GET /resources/{id}", app.ResourceView)
	protected.HandleFunc("POST /resources/{id}", app.ResourceUpdate)
	protected.HandleFunc("POST /resources/{id}/delete", app.ResourceDelete)
	protected.HandleFunc("GET /resources/{id}/rates", app.ResourceRateHistory)
	protected.HandleFunc("POST /resources/{id}/rates", app.ResourceRateCreate)

	protected.HandleFunc("GET /sessions/new", app.SessionCreate)
	protected.HandleFunc("POST /sessions/new", app.SessionCreatePost)
	protected.HandleFunc("GET /sessions/{id}", app.SessionView)
	protected.HandleFunc("POST /sessions/{id}/submit", app.SessionSubmit)

	protected.HandleFunc("GET /sessions/{id}/resources/new", app.JobResourceCreate)
	protected.HandleFunc("POST /sessions/{id}/resources/new", app.JobResourceCreatePost)
	protected.HandleFunc("POST /sessions/{id}/resources/{resId}", app.JobResourceUpdate)

	mux.Handle("/", app.RequireAuth(protected))

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		infoLog.Printf("Start Server on %s\n", *addr)
		err = srv.ListenAndServe()
		errorLog.Fatalln(err)
	}()

	<-stop
	infoLog.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		infoLog.Printf("Graceful shutdown failed: %v", err.Error())
	} else {
		infoLog.Println("Server shut down gracefully")
	}
}
