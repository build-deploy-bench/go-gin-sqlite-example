package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gothinkster/golang-gin-realworld-example-app/articles"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/gothinkster/golang-gin-realworld-example-app/users"
	"gorm.io/gorm"
)

// commit is the full Git SHA of the build. bench/build.sh sets it with -ldflags "-X main.commit=<sha>".
var commit = "unknown"

func Migrate(db *gorm.DB) {
	users.AutoMigrate()
	db.AutoMigrate(&articles.ArticleModel{})
	db.AutoMigrate(&articles.TagModel{})
	db.AutoMigrate(&articles.FavoriteModel{})
	db.AutoMigrate(&articles.ArticleUserModel{})
	db.AutoMigrate(&articles.CommentModel{})
}

func main() {
	// The benchmark parsers read this marker as the start of app boot. The binary is the whole start command.
	fmt.Printf("::bench::start %d\n", time.Now().UnixMilli())

	db := common.Init()
	Migrate(db)
	sqlDB, err := db.DB()
	if err != nil {
		log.Println("failed to get sql.DB:", err)
	} else {
		defer sqlDB.Close()
	}

	r := gin.Default()

	// Disable automatic redirect for trailing slashes
	// This prevents POST body from being lost during redirects
	r.RedirectTrailingSlash = false

	v1 := r.Group("/api")
	users.UsersRegister(v1.Group("/users"))
	v1.Use(users.AuthMiddleware(false))
	articles.ArticlesAnonymousRegister(v1.Group("/articles"))
	articles.TagsAnonymousRegister(v1.Group("/tags"))
	users.ProfileRetrieveRegister(v1.Group("/profiles"))

	v1.Use(users.AuthMiddleware(true))
	users.UserRegister(v1.Group("/user"))
	users.ProfileRegister(v1.Group("/profiles"))

	articles.ArticlesRegister(v1.Group("/articles"))

	// The benchmark probe polls this until it returns the run ID of the new deployment. The harness sets
	// BENCH_RUN_ID before each deploy, and each deployment keeps the env it started with.
	runID := os.Getenv("BENCH_RUN_ID")
	if runID == "" {
		runID = "unknown"
	}
	// Railway gets no BENCH_RUN_ID: a changed variable invalidates part of its build cache. Its own deployment
	// ID identifies the deployment instead.
	deployment := os.Getenv("RAILWAY_DEPLOYMENT_ID")
	if deployment == "" {
		deployment = "unknown"
	}
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"commit": commit, "run": runID, "deployment": deployment})
	})

	testAuth := r.Group("/api/ping")

	testAuth.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Listen before serving, so the marker is printed only when the port accepts connections.
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("failed to start server:", err)
	}
	fmt.Printf("::bench::listening %d\n", time.Now().UnixMilli())
	if err := r.RunListener(listener); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
