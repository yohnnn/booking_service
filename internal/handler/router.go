package handler

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	v1 "github.com/yohnnn/booking_service/internal/handler/v1"
	mw "github.com/yohnnn/booking_service/internal/middleware"
	"github.com/yohnnn/booking_service/internal/service"
)

type Router struct {
	logger         *slog.Logger
	authService    service.Auth
	authHandler    *v1.AuthHandler
	concertHandler *v1.ConcertHandler
	bookingHandler *v1.BookingHandler
}

func NewRouter(
	logger *slog.Logger,
	authService service.Auth,
	authHandler *v1.AuthHandler,
	concertHandler *v1.ConcertHandler,
	bookingHandler *v1.BookingHandler,
) *Router {
	return &Router{
		logger:         logger,
		authService:    authService,
		authHandler:    authHandler,
		concertHandler: concertHandler,
		bookingHandler: bookingHandler,
	}
}

func (r *Router) InitRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)

	router.Route("/api", func(mr chi.Router) {

		mr.Group(func(ar chi.Router) {
			ar.Post("/auth/register", r.authHandler.Register)
			ar.Post("/auth/login", r.authHandler.Login)
		})

		mr.Get("/concerts", r.concertHandler.GetAll)
		mr.Get("/concerts/{id}", r.concertHandler.GetByID)

		mr.Group(func(pr chi.Router) {
			pr.Use(mw.Auth(r.authService))
			pr.Post("/bookings", r.bookingHandler.Create)
			pr.Get("/bookings", r.bookingHandler.GetUserBookings)
		})
	})

	return router
}
