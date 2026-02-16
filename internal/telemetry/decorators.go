package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order domain.OrderWithInformation) error
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error)
	WarmUp(ctx context.Context) error
}

type OrderRepository interface {
	Create(ctx context.Context, order domain.OrderWithInformation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error)
	GetAllLast24Hours(ctx context.Context) ([]domain.OrderWithInformation, error)
	Ping(ctx context.Context) error
}

type Cache interface {
	Set(ctx context.Context, key uuid.UUID, value domain.OrderWithInformation)
	Get(ctx context.Context, key uuid.UUID) (*domain.OrderWithInformation, bool)
}

func WrapOrderService(next OrderService) OrderService {
	return &orderServiceTelemetry{next: next}
}

func WrapOrderRepository(next OrderRepository) OrderRepository {
	return &orderRepositoryTelemetry{next: next}
}

func WrapCache(next Cache) Cache {
	return &cacheTelemetry{next: next}
}

type orderServiceTelemetry struct {
	next OrderService
}

func (t *orderServiceTelemetry) CreateOrder(ctx context.Context, order domain.OrderWithInformation) error {
	ctx, span := otel.Tracer("service").Start(ctx, "OrderService.CreateOrder")
	defer span.End()

	err := t.next.CreateOrder(ctx, order)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("create order failed", slog.Any("error", err))
	}

	return err
}

func (t *orderServiceTelemetry) GetOrder(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "OrderService.GetOrder")
	defer span.End()

	order, err := t.next.GetOrder(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("get order failed", slog.Any("error", err))
	}

	return order, err
}

func (t *orderServiceTelemetry) WarmUp(ctx context.Context) error {
	ctx, span := otel.Tracer("service").Start(ctx, "OrderService.WarmUp")
	defer span.End()

	err := t.next.WarmUp(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("cache warm-up failed", slog.Any("error", err))
	}

	return err
}

type orderRepositoryTelemetry struct {
	next OrderRepository
}

func (t *orderRepositoryTelemetry) Create(ctx context.Context, order domain.OrderWithInformation) error {
	ctx, span := otel.Tracer("repository").Start(ctx, "OrderRepository.Create")
	defer span.End()

	err := t.next.Create(ctx, order)
	if err != nil {
		IncStorageOp("db", "write", "error")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("repository create failed", slog.Any("error", err))
		return err
	}

	IncStorageOp("db", "write", "ok")
	return nil
}

func (t *orderRepositoryTelemetry) GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
	ctx, span := otel.Tracer("repository").Start(ctx, "OrderRepository.GetByID")
	defer span.End()

	order, err := t.next.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			IncStorageOp("db", "read", "miss")
			return nil, err
		}
		IncStorageOp("db", "read", "error")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("repository get failed", slog.Any("error", err))
		return nil, err
	}

	IncStorageOp("db", "read", "ok")
	return order, nil
}

func (t *orderRepositoryTelemetry) GetAllLast24Hours(ctx context.Context) ([]domain.OrderWithInformation, error) {
	ctx, span := otel.Tracer("repository").Start(ctx, "OrderRepository.GetAllLast24Hours")
	defer span.End()

	orders, err := t.next.GetAllLast24Hours(ctx)
	if err != nil {
		IncStorageOp("db", "read", "error")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("repository get all failed", slog.Any("error", err))
		return nil, err
	}

	IncStorageOp("db", "read", "ok")
	return orders, nil
}

func (t *orderRepositoryTelemetry) Ping(ctx context.Context) error {
	ctx, span := otel.Tracer("repository").Start(ctx, "OrderRepository.Ping")
	defer span.End()

	if err := t.next.Ping(ctx); err != nil {
		SetRepositoryUp("order_postgres", false)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Warn("repository ping failed", slog.Any("error", err))
		return err
	}

	SetRepositoryUp("order_postgres", true)
	return nil
}

type cacheTelemetry struct {
	next Cache
}

func (t *cacheTelemetry) Set(ctx context.Context, key uuid.UUID, value domain.OrderWithInformation) {
	ctx, span := otel.Tracer("cache").Start(ctx, "cache.set")
	defer span.End()

	t.next.Set(ctx, key, value)
	IncStorageOp("cache", "write", "ok")
}

func (t *cacheTelemetry) Get(ctx context.Context, key uuid.UUID) (*domain.OrderWithInformation, bool) {
	ctx, span := otel.Tracer("cache").Start(ctx, "cache.get")
	defer span.End()

	value, ok := t.next.Get(ctx, key)
	if !ok {
		span.SetAttributes(attribute.Bool("cache.hit", false))
		IncStorageOp("cache", "read", "miss")
		return nil, false
	}

	span.SetAttributes(attribute.Bool("cache.hit", true))
	IncStorageOp("cache", "read", "hit")
	return value, true
}
