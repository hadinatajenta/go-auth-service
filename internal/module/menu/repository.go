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
	SetPermission(ctx context.Context, menuID, permID uint) error
	GetPermissionID(ctx context.Context, menuID uint) (uint, error)
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
    SELECT DISTINCT m.*
		FROM menus m
		LEFT JOIN menu_permissions mp ON m.id = mp.menu_id
		LEFT JOIN role_permissions rp ON mp.permission_id = rp.permission_id
		LEFT JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?
		OR mp.permission_id IS NULL
		),
		nested_menus AS (
			SELECT * FROM exact_authorized_menus
			UNION
			SELECT m.*
			FROM menus m
			JOIN nested_menus n ON m.id = n.parent_id
		)
	SELECT * FROM nested_menus
	ORDER BY sort_order ASC;
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete permissions mapping first
		if err := tx.Where("menu_id = ?", id).Delete(&MenuPermission{}).Error; err != nil {
			return err
		}
		// Delete the menu
		return tx.Delete(&Menu{}, id).Error
	})
}

func (r *repository) SetPermission(ctx context.Context, menuID, permID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing mapping
		tx.Where("menu_id = ?", menuID).Delete(&MenuPermission{})

		// If permID is 0, we just leave it deleted (public)
		if permID == 0 {
			return nil
		}

		// Create new mapping
		mp := &MenuPermission{MenuID: menuID, PermissionID: permID}
		return tx.Create(mp).Error
	})
}

func (r *repository) GetPermissionID(ctx context.Context, menuID uint) (uint, error) {
	var mp MenuPermission
	err := r.db.WithContext(ctx).Where("menu_id = ?", menuID).First(&mp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return mp.PermissionID, nil
}
