package menu

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, menu *Menu) error
	GetByID(ctx context.Context, id uint) (*Menu, error)
	List(ctx context.Context) ([]Menu, error)
	GetAccessibleMenus(ctx context.Context, userID uint) ([]Menu, error)
	Update(ctx context.Context, menu *Menu) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Create(ctx context.Context, menu *Menu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Menu, error) {
	var menu Menu
	if err := r.db.WithContext(ctx).First(&menu, id).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *repository) List(ctx context.Context) ([]Menu, error) {
	var menus []Menu
	if err := r.db.WithContext(ctx).Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *repository) GetAccessibleMenus(ctx context.Context, userID uint) ([]Menu, error) {
	var menus []Menu

	query := `
	WITH RECURSIVE exact_authorized_menus AS (
		-- Base case: Menus explicitly authorized via user_roles -> role_permissions -> menu_permissions
		SELECT DISTINCT m.*
		FROM menus m
		JOIN menu_permissions mp ON m.id = mp.menu_id
		JOIN role_permissions rp ON mp.permission_id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?
	),
	nested_menus AS (
		-- Base nodes (authorized children and parents)
		SELECT * FROM exact_authorized_menus
		
		UNION
		
		-- Recursive step: find parent of currently found menus
		SELECT m.*
		FROM menus m
		JOIN nested_menus n ON m.id = n.parent_id
	)
	SELECT * FROM nested_menus ORDER BY sort_order ASC;
	`

	if err := r.db.WithContext(ctx).Raw(query, userID).Scan(&menus).Error; err != nil {
		return nil, err
	}

	return menus, nil
}

func (r *repository) Update(ctx context.Context, menu *Menu) error {
	return r.db.WithContext(ctx).Save(menu).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Menu{}, id).Error
}
