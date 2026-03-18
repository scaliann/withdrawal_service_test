package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ver1 "github.com/scaliann/withdrawal_service_test/internal/controller/http/v1"
	"github.com/scaliann/withdrawal_service_test/internal/usecase"
	"github.com/scaliann/withdrawal_service_test/pkg/logger"
	"github.com/scaliann/withdrawal_service_test/pkg/metrics"
	"github.com/scaliann/withdrawal_service_test/pkg/otel"
)

func WithdrawalRouter(r *chi.Mux, uc *usecase.UseCase, m *metrics.HTTPServer) {
	v1 := ver1.New(uc)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/", func(r chi.Router) {
		r.Use(logger.Middleware)
		r.Use(metrics.NewMiddleware(m))
		r.Use(otel.Middleware)
		r.Route("/v1", func(r chi.Router) {
			r.Post("/auth/token", v1.IssueToken)
			r.Post("/auth/refresh", v1.RefreshToken)
			r.Get("/auth/verify", v1.VerifyToken)

			r.Group(func(r chi.Router) {
				r.Use(v1.AuthMiddleware)
				r.Get("/withdrawals", v1.GetWithdrawals)
				r.Get("/withdrawals/{id}", v1.GetWithdrawal)
				r.Post("/withdrawals", v1.CreateWithdrawal)
				r.Get("/reports/user-withdrawals-summary", v1.GetUserWithdrawalsSummary)
			})
		})
	})
}
