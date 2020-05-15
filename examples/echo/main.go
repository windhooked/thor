package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"

	//	_ "github.com/mattn/go-sqlite3"
	//_ "github.com/jinzhu/gorm/dialects/mysql"
	// _ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	// _ "github.com/jinzhu/gorm/dialects/mssql"

	"github.com/windhooked/thor/admin"
	"github.com/windhooked/thor/auth"
	"github.com/windhooked/thor/auth/auth_identity"
	"github.com/windhooked/thor/auth_themes/clean"
	"github.com/windhooked/thor/qor"
	"github.com/windhooked/thor/session/manager"
)

type User struct {
	gorm.Model
	Email    string
	Password string
	Name     string
	Gender   string
	Role     string
	Basic    auth_identity.Basic
	SignLog  auth_identity.SignLog
}

type AdminAuth struct{}

var (
	DB, _ = gorm.Open("sqlite3", "example.db")

	Auth = clean.New(&auth.Config{
		DB:                DB,
		UserModel:         &User{},
		AuthIdentityModel: &auth_identity.AuthIdentity{},
		ViewPaths:         []string{"auth/views"},
	})
)

func (user *User) DisplayName() string {
	return user.Name
}

func (AdminAuth) LoginURL(c *admin.Context) string {
	return "/auth/login"
}

func (AdminAuth) LogoutURL(c *admin.Context) string {
	return "/auth/logout"
}

func (AdminAuth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	currentUser := Auth.GetCurrentUser(c.Request)
	if currentUser != nil {
		qorCurrentUser, ok := currentUser.(qor.CurrentUser)
		if !ok {
			fmt.Printf("User %#v haven't implement qor.CurrentUser interface\n", currentUser)
		}
		return qorCurrentUser
	}
	return nil
}

func main() {
	// Setup DB
	DB.LogMode(true)
	DB.AutoMigrate(&User{}, &auth_identity.AuthIdentity{}, &auth_identity.Basic{}, &auth_identity.SignLogs{})

	// Setup Admin
	Admin := admin.New(&admin.AdminConfig{
		DB:   DB,
		Auth: &AdminAuth{},
	})
	Admin.AddResource(&User{})

	// Setup Echo
	e := echo.New()

	// Middleware
	authMiddleware := echo.WrapMiddleware(manager.SessionManager.Middleware)
	e.Use(authMiddleware)

	// Handler
	adminHandler := echo.WrapHandler(Admin.NewServeMux("/admin"))
	e.GET("/admin", adminHandler)
	e.Any("/admin/*", adminHandler)

	authHandler := echo.WrapHandler(Auth.NewServeMux())
	e.Any("/auth/*", authHandler)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
