package role

import (
	"context"
	"testing"
)

type mockRepository struct {
	Repository
	roles map[uint]*Role
}

func (m *mockRepository) GetByID(ctx context.Context, id uint) (*Role, error) {
	role, ok := m.roles[id]
	if !ok {
		return nil, nil // Should return error in real repo, but for hasCycle nil parent is check enough
	}
	return role, nil
}

func TestHasCycle(t *testing.T) {
	repo := &mockRepository{
		roles: make(map[uint]*Role),
	}
	
	// Setup roles
	// 1 -> 2 -> 3
	parentID2 := uint(1)
	parentID3 := uint(2)
	repo.roles[1] = &Role{ID: 1}
	repo.roles[2] = &Role{ID: 2, ParentID: &parentID2}
	repo.roles[3] = &Role{ID: 3, ParentID: &parentID3}

	s := &service{repo: repo}

	tests := []struct {
		name     string
		parentID uint
		childID  uint
		want     bool
	}{
		{"No cycle - simple", 1, 3, false},
		{"No cycle - root", 0, 1, false},
		{"Cycle - self", 1, 1, false}, // Explicitly false here; handled by s.Update direct equality check
		{"Cycle - direct", 3, 2, true},
		{"Cycle - indirect", 3, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// In our service.Update, we check: 
			// if *req.ParentID == role.ID { return error } // Direct self-cycle handled in Update
			// if s.hasCycle(ctx, *req.ParentID, role.ID) { return error }
			
			// hasCycle(ctx, parentID, childID) checks if childID is an ancestor of parentID
			if got := s.hasCycle(context.Background(), tt.parentID, tt.childID); got != tt.want {
				t.Errorf("hasCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}
