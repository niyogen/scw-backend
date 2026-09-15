package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"delivery-backend/internal/config"
	"delivery-backend/internal/database"
	"delivery-backend/internal/handler"
	"delivery-backend/internal/middleware"
	"delivery-backend/internal/repository"
	"delivery-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. Initialize PostgreSQL Database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 3. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	workerRepo := repository.NewWorkerRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	locationRepo := repository.NewLocationRepository(db)

	// 4. Initialize Services
	authSvc := service.NewAuthService(userRepo, cfg)
	serviceCatalogSvc := service.NewServiceCatalogService(serviceRepo)
	workerSvc := service.NewWorkerService(workerRepo, userRepo)
	voucherSvc := service.NewVoucherService(voucherRepo)
	locationSvc := service.NewLocationService(locationRepo)
	bookingSvc := service.NewBookingService(bookingRepo, voucherRepo, workerRepo, serviceRepo)
	mediaSvc := service.NewMediaService(cfg)

	// 5. Initialize Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	serviceHandler := handler.NewServiceHandler(serviceCatalogSvc)
	workerHandler := handler.NewWorkerHandler(workerSvc)
	bookingHandler := handler.NewBookingHandler(bookingSvc)
	voucherHandler := handler.NewVoucherHandler(voucherSvc)
	locationHandler := handler.NewLocationHandler(locationSvc)
	mediaHandler := handler.NewMediaHandler(mediaSvc)
	healthHandler := handler.NewHealthHandler(db)

	// 6. Router Setup
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// Health and Root
	r.GET("/", healthHandler.RootHandler)
	r.GET("/health", healthHandler.HealthCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public Auth routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		}

		// Service Catalog routes
		serviceGroup := v1.Group("/services")
		{
			serviceGroup.GET("", serviceHandler.GetAllServices)
			serviceGroup.GET("/:id", serviceHandler.GetServiceByID)
			serviceGroup.POST("", serviceHandler.CreateService)
			serviceGroup.PUT("/:id", serviceHandler.UpdateService)
			serviceGroup.PATCH("/:id/status", serviceHandler.ToggleServiceStatus)
			serviceGroup.DELETE("/:id", serviceHandler.DeleteService)
		}

		// Worker Discovery & Admin routes
		workerGroup := v1.Group("/workers")
		{
			workerGroup.GET("", workerHandler.GetAllWorkers)
			workerGroup.GET("/:id", workerHandler.GetWorkerByID)
			workerGroup.POST("", workerHandler.CreateWorker)
			workerGroup.POST("/register", workerHandler.RegisterWorker)
			workerGroup.PUT("/:id", workerHandler.UpdateWorker)
			workerGroup.PATCH("/:id/status", workerHandler.ToggleWorkerStatus)
			workerGroup.DELETE("/:id", workerHandler.DeleteWorker)
		}

		// Public Voucher routes
		voucherGroup := v1.Group("/vouchers")
		{
			voucherGroup.GET("", voucherHandler.GetActiveVouchers)
			voucherGroup.POST("/validate", voucherHandler.ValidateVoucher)
		}

		// Media & Profile Photo Uploads (Google Cloud Storage)
		v1.POST("/upload/photo", mediaHandler.UploadPhoto)
		v1.POST("/upload/avatar", mediaHandler.UploadPhoto)
		v1.GET("/media/*filepath", mediaHandler.ServeMedia)

		// Admin routes (Bookings, Customers)
		adminGroup := v1.Group("/admin")
		{
			// Bookings
			adminGroup.GET("/bookings", bookingHandler.GetAllBookings)
			adminGroup.PATCH("/bookings/:id/status", bookingHandler.UpdateBookingStatus)

			// Customers
			adminGroup.GET("/customers", authHandler.GetAllCustomers)
			adminGroup.GET("/customers/:id", authHandler.GetCustomerByID)
			adminGroup.POST("/customers", authHandler.CreateCustomer)
			adminGroup.PUT("/customers/:id", authHandler.UpdateCustomer)
			adminGroup.PATCH("/customers/:id/status", authHandler.ToggleCustomerStatus)
			adminGroup.DELETE("/customers/:id", authHandler.DeleteCustomer)

			// Review Moderation (Approve / Reject / Delete)
			adminGroup.GET("/reviews", workerHandler.GetAdminReviews)
			adminGroup.PATCH("/reviews/:id/status", workerHandler.UpdateReviewStatus)
			adminGroup.DELETE("/reviews/:id", workerHandler.DeleteReview)
		}

		// Protected User routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// User Profile
			userGroup := protected.Group("/user")
			{
				userGroup.GET("/profile", authHandler.GetProfile)
				userGroup.PUT("/profile", authHandler.UpdateProfile)
				userGroup.POST("/change-password", authHandler.ChangePassword)

				// Saved Locations
				userGroup.GET("/locations", locationHandler.GetLocations)
				userGroup.POST("/locations", locationHandler.AddLocation)
				userGroup.PUT("/locations/:id/default", locationHandler.SetDefaultLocation)
				userGroup.DELETE("/locations/:id", locationHandler.DeleteLocation)
			}

			// Worker Reviews
			protected.POST("/workers/:id/reviews", workerHandler.AddWorkerReview)

			// Booking Endpoints
			bookingGroup := protected.Group("/bookings")
			{
				bookingGroup.POST("", bookingHandler.CreateBooking)
				bookingGroup.GET("", bookingHandler.GetMyBookings)
				bookingGroup.GET("/:id", bookingHandler.GetBookingByID)
				bookingGroup.POST("/:id/cancel", bookingHandler.CancelBooking)
				bookingGroup.PATCH("/:id/status", bookingHandler.UpdateBookingStatus)
			}
		}
	}

	// 7. Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Delivery Service API is running on port :%s (env: %s)", cfg.Port, cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting successfully")
}
