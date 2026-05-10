package room

import (
	"context"
	"go-booking-management-init/internal/domain/room"
	"time"
)

type SearchProvider interface {
	Name() string
	Search(ctx context.Context, filter room.AvailabilityFilter) ([]*room.Room, error)
}

type DBSearchProvider struct {
	repo room.Repository
}

func (p *DBSearchProvider) Name() string { return "Local Database" }

func (p *DBSearchProvider) Search(ctx context.Context, filter room.AvailabilityFilter) ([]*room.Room, error) {
	return p.repo.ListAvailable(ctx, filter)
}

func NewDBSearchProvider(repo room.Repository) *DBSearchProvider {
	return &DBSearchProvider{repo: repo}
}

type SimulatedSearchProvider struct {
	name  string
	delay time.Duration
	rooms []*room.Room
}

func NewSimulatedSearchProvider(name string, delay time.Duration, rooms []*room.Room) *SimulatedSearchProvider {
	return &SimulatedSearchProvider{
		name:  name,
		delay: delay,
		rooms: rooms,
	}
}

func (p *SimulatedSearchProvider) Name() string { return p.name }

func (p *SimulatedSearchProvider) Search(ctx context.Context, _ room.AvailabilityFilter) ([]*room.Room, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(p.delay):
		return p.rooms, nil
	}
}
