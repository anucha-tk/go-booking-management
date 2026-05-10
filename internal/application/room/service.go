package room

import (
	"context"
	"fmt"
	"go-booking-management-init/internal/domain/room"
	"time"
)

type Service struct {
	repo      room.Repository
	providers []SearchProvider
}

func NewService(repo room.Repository, providers []SearchProvider) *Service {
	if providers == nil {
		providers = []SearchProvider{
			&DBSearchProvider{repo: repo},
		}
	}
	return &Service{
		repo:      repo,
		providers: providers,
	}
}

func (s *Service) CreateRoom(ctx context.Context, rm *room.Room) (*room.Room, error) {
	if rm.Status == "" {
		rm.Status = room.StatusAvailable
	}
	if !room.IsValidStatus(rm.Status) {
		return nil, room.ErrInvalidRoomStatus
	}

	return s.repo.Create(ctx, rm)
}

func (s *Service) GetRoom(ctx context.Context, id int32) (*room.Room, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetRoomDetail(ctx context.Context, id int32) (*room.Detail, error) {
	detail, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get room detail: %w", err)
	}
	if detail == nil {
		return nil, room.ErrRoomNotFound
	}
	return detail, nil
}

func (s *Service) ListRooms(ctx context.Context, filter room.Filter) ([]*room.Room, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) ListAvailableRooms(ctx context.Context, filter room.AvailabilityFilter) ([]*room.Room, error) {
	if filter.StartDate.IsZero() || filter.EndDate.IsZero() {
		return nil, room.ErrMissingDates
	}
	if !filter.StartDate.Before(filter.EndDate) {
		return nil, room.ErrInvalidDateRange
	}
	// Use midnight today UTC as reference to avoid timezone issues
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if filter.StartDate.UTC().Before(today) {
		return nil, room.ErrPastDate
	}

	// NFR1: Aggregator timeout 200ms
	searchCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	type result struct {
		rooms []*room.Room
		err   error
	}

	resChan := make(chan result, len(s.providers))

	for _, p := range s.providers {
		go func(p SearchProvider) {
			rooms, err := p.Search(searchCtx, filter)
			select {
			case resChan <- result{rooms, err}:
			case <-searchCtx.Done():
			}
		}(p)
	}

	roomMap := make(map[int32]*room.Room)
	for i := 0; i < len(s.providers); i++ {
		select {
		case res := <-resChan:
			if res.err != nil {
				continue
			}
			for _, r := range res.rooms {
				roomMap[r.ID] = r
			}
		case <-searchCtx.Done():
			// Timeout or parent context cancelled
			break
		}
	}

	allRooms := make([]*room.Room, 0, len(roomMap))
	for _, r := range roomMap {
		allRooms = append(allRooms, r)
	}

	return allRooms, nil
}

func (s *Service) UpdateRoom(ctx context.Context, rm *room.Room) (*room.Room, error) {
	if !room.IsValidStatus(rm.Status) {
		return nil, room.ErrInvalidRoomStatus
	}
	return s.repo.Update(ctx, rm)
}

func (s *Service) DeleteRoom(ctx context.Context, id int32) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) UpdateRoomStatus(ctx context.Context, id int32, status string) (*room.Room, error) {
	if !room.IsValidStatus(status) {
		return nil, room.ErrInvalidRoomStatus
	}
	return s.repo.UpdateStatus(ctx, id, status)
}
